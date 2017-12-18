## How to update the `scoop` manifest

The `scoop` manifest for the faas-cli is part of the official [sccop](https://github.com/lukesampson/scoop/blob/master/bucket/faas-cli.json) repo on Github. It needs to be updated for each subsequent release.

#### Simple version bumps

```
git clone https://github.com/lukesampson/scoop
cd scoop
./bin/checkver.ps1 faas-cli -u
```

Test the updated manifest
```
scoop install .\bucket\faas-cli.json
```

Create a new branch and commit the manifest `faas-cli.json`, then create a PR to update the manifest in Scoop repository

