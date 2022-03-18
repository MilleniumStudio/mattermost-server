git clone --depth=1 --no-single-branch https://github.com/mattermost/mattermost-webapp.git
cd mattermost-webapp
git checkout $CIRCLE_BRANCH || git checkout release-6.5
export WEBAPP_GIT_COMMIT=$(git rev-parse HEAD)
echo "$WEBAPP_GIT_COMMIT"

FILE_DIST=dist.tar.gz
curl --version

runtime="2 minute"
endtime=$(date -ud "$runtime" +%s)
while [[ $(date -u +%s) -le $endtime ]]; do
    if curl --max-time 30 -f -o $FILE_DIST https://pr-builds.mattermost.com/mattermost-webapp/commit/$WEBAPP_GIT_COMMIT/mattermost-webapp.tar.gz; then
        break
    fi
    echo "Waiting for webapp git commit $WEBAPP_GIT_COMMIT with sleep 5: `date +%H:%M:%S`"
    sleep 5
done
ls -al
if tar -tzf $FILE_DIST >/dev/null; then
    mkdir dist && tar -xvf $FILE_DIST -C dist --strip-components=1
else
    npm ci && make build
fi
cd -
