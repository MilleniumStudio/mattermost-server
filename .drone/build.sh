make config-reset
make build-cmd BUILD_NUMBER='${CIRCLE_BRANCH}-${CIRCLE_BUILD_NUM}'
make package BUILD_NUMBER='${CIRCLE_BRANCH}-${CIRCLE_BUILD_NUM}'
