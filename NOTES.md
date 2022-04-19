Upgrade Mattermost
====

We don't care about `master` branch we work only on `release-x.x` branches

In this example we will create the release 6.6.0 containing our license changes

## Create a new major/medium release x.x.0

0/ Be sure to have official MM git repo as remote
```
git remote add mattermost git@github.com:mattermost/mattermost-server.git
```

1/ Fetch all changes and create new branch from official repo
```
git fetch --all
git checkout -b release-6.6 --track mattermost/release-6.6
```

2/ Reset branch to commit tagged to the version you which. It's generally few commits behind
```
git reset --hard e05e0f9379e1324618601e3ba6111d78710e6319
```

3/ Apply our changes
```
git cherry-pick d68ae55e015e426f1ceb251d13db144d5e934e65
```

4/ Overwrite MM tag to our commit
```
git tag v6.6.0
```

5/ Push branch and tag
```
git push origin release-6.6
git push origin v6.6.0
```

## Create a new minor release 6.6.x

TODO


## Notes

Check `setup` step in the Drone build and make sure the script if fetching correct branch of the webapp and not master or other

```
+ bash .drone/setup.sh
Cloning into 'mattermost-webapp'...
Branch 'release-6.6' set up to track remote branch 'release-6.6' from 'origin'.
Switched to a new branch 'release-6.6'
```