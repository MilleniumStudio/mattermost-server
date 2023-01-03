// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

package app

import (
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"

	"github.com/mattermost/mattermost-server/v6/model"
	"github.com/mattermost/mattermost-server/v6/product"
	"github.com/mattermost/mattermost-server/v6/store"
)

const (
	LicenseEnv                = "MM_LICENSE"
	LicenseRenewalURL         = "https://customers.mattermost.com/subscribe/renew"
	JWTDefaultTokenExpiration = 7 * 24 * time.Hour // 7 days of expiration
)

var RequestTrialURL = "https://customers.mattermost.com/api/v1/trials"

// ensure the license service wrapper implements `product.LicenseService`
var _ product.LicenseService = (*licenseWrapper)(nil)

// licenseWrapper is an adapter struct that only exposes the
// config related functionality to be passed down to other products.
type licenseWrapper struct {
	srv *Server
}

func (w *licenseWrapper) Name() ServiceKey {
	return LicenseKey
}

func (w *licenseWrapper) GetLicense() *model.License {
	return w.srv.License()
}

func (w *licenseWrapper) RequestTrialLicense(requesterID string, users int, termsAccepted bool, receiveEmailsAccepted bool) *model.AppError {
	if *w.srv.platform.Config().ExperimentalSettings.RestrictSystemAdmin {
		return model.NewAppError("RequestTrialLicense", "api.restricted_system_admin", nil, "", http.StatusForbidden)
	}

	if !termsAccepted {
		return model.NewAppError("RequestTrialLicense", "api.license.request-trial.bad-request.terms-not-accepted", nil, "", http.StatusBadRequest)
	}

	if users == 0 {
		return model.NewAppError("RequestTrialLicense", "api.license.request-trial.bad-request", nil, "", http.StatusBadRequest)
	}

	requester, err := w.srv.userService.GetUser(requesterID)
	if err != nil {
		var nfErr *store.ErrNotFound
		switch {
		case errors.As(err, &nfErr):
			return model.NewAppError("RequestTrialLicense", MissingAccountError, nil, "", http.StatusNotFound).Wrap(err)
		default:
			return model.NewAppError("RequestTrialLicense", "app.user.get_by_username.app_error", nil, "", http.StatusInternalServerError).Wrap(err)
		}
	}

	trialLicenseRequest := &model.TrialLicenseRequest{
		ServerID:              w.srv.TelemetryId(),
		Name:                  requester.GetDisplayName(model.ShowFullName),
		Email:                 requester.Email,
		SiteName:              *w.srv.platform.Config().TeamSettings.SiteName,
		SiteURL:               *w.srv.platform.Config().ServiceSettings.SiteURL,
		Users:                 users,
		TermsAccepted:         termsAccepted,
		ReceiveEmailsAccepted: receiveEmailsAccepted,
	}

	return w.srv.platform.RequestTrialLicense(trialLicenseRequest)
}

// JWTClaims custom JWT claims with the needed information for the
// renewal process
type JWTClaims struct {
	LicenseID   string `json:"license_id"`
	ActiveUsers int64  `json:"active_users"`
	jwt.StandardClaims
}

func (s *Server) License() *model.License {
	return s.platform.License()
}

func (s *Server) LoadLicense() {
	// ENV var overrides all other sources of license.
	licenseStr := os.Getenv(LicenseEnv)
	if licenseStr != "" {
		license, err := utils.LicenseValidator.LicenseFromBytes([]byte(licenseStr))
		if err != nil {
			mlog.Error("Failed to read license set in environment.", mlog.Err(err))
			return
		}

		// skip the restrictions if license is a sanctioned trial
		if !license.IsSanctionedTrial() && license.IsTrialLicense() {
			canStartTrialLicense, err := s.LicenseManager.CanStartTrial()
			if err != nil {
				mlog.Error("Failed to validate trial eligibility.", mlog.Err(err))
				return
			}

			if !canStartTrialLicense {
				mlog.Info("Cannot start trial multiple times.")
				return
			}
		}

		if s.ValidateAndSetLicenseBytes([]byte(licenseStr)) {
			mlog.Info("License key from ENV is valid, unlocking enterprise features.")
		}
		return
	}

	/*
	licenseId := ""
	props, nErr := s.Store.System().Get()
	if nErr == nil {
		licenseId = props[model.SystemActiveLicenseId]
	}

	if !model.IsValidId(licenseId) {
		// Lets attempt to load the file from disk since it was missing from the DB
		license, licenseBytes := utils.GetAndValidateLicenseFileFromDisk(*s.Config().ServiceSettings.LicenseFileLocation)

		if license != nil {
			if _, err := s.SaveLicense(licenseBytes); err != nil {
				mlog.Error("Failed to save license key loaded from disk.", mlog.Err(err))
			} else {
				licenseId = license.Id
			}
		}
	}

	record, nErr := s.Store.License().Get(licenseId)
	if nErr != nil {
		mlog.Error("License key from https://mattermost.com required to unlock enterprise features.", mlog.Err(nErr))
		s.SetLicense(nil)
		return
	}

	s.ValidateAndSetLicenseBytes([]byte(record.Bytes))
	*/

	// Enable all OSS features for everyone
	f := model.Features{}
	f.SetDefaults()
	*f.Users = 9999

	s.SetLicense(&model.License{
		Id:        model.NewId(),
		IssuedAt:  0,
		ExpiresAt: 4102491600000, // 1st january 2100 in ms
		Customer:  &model.Customer{
			Name:    "Mr Robot",
			Email:   "mrrobot@fsociety.com",
			Company: "fsociety",
		},
		Features:  &f,
		SkuName: "Enterprise",
		SkuShortName: model.LicenseShortSkuEnterprise,
	})

	mlog.Info("License key valid unlocking enterprise features.")
}

func (s *Server) SaveLicense(licenseBytes []byte) (*model.License, *model.AppError) {
	return s.platform.SaveLicense(licenseBytes)
}

func (s *Server) SetLicense(license *model.License) bool {
	return s.platform.SetLicense(license)
}

func (s *Server) ValidateAndSetLicenseBytes(b []byte) bool {
	return s.platform.ValidateAndSetLicenseBytes(b)
}

func (s *Server) SetClientLicense(m map[string]string) {
	s.platform.SetClientLicense(m)
}

func (s *Server) ClientLicense() map[string]string {
	return s.platform.ClientLicense()
}

func (s *Server) RemoveLicense() *model.AppError {
	return s.platform.RemoveLicense()
}

func (s *Server) AddLicenseListener(listener func(oldLicense, newLicense *model.License)) string {
	return s.platform.AddLicenseListener(listener)
}

func (s *Server) RemoveLicenseListener(id string) {
	s.platform.RemoveLicenseListener(id)
}

func (s *Server) GetSanitizedClientLicense() map[string]string {
	return s.platform.GetSanitizedClientLicense()
}

// GenerateRenewalToken returns a renewal token that expires after duration expiration
func (s *Server) GenerateRenewalToken(expiration time.Duration) (string, *model.AppError) {
	return s.platform.GenerateRenewalToken(expiration)
}

// GenerateLicenseRenewalLink returns a link that points to the CWS where clients can renew license
func (s *Server) GenerateLicenseRenewalLink() (string, string, *model.AppError) {
	return s.platform.GenerateLicenseRenewalLink()
}
