| Command | Description |
| ------- | ----------- |
| **build verb** |
| `faas-cli` | prints available verbs and global flags (think of docker/kubectl etc) |
| `faas-cli --help` | as above |
| `faas-cli build` | help for build verb |
| `faas-cli build --help` | as above |
| `faas-cli build .` | Build using a local `Faasfile` (ie default name for the YAML) with context set to `.` (handlers are relative to this) |
| `faas-cli build -f /path/to/yaml .` | As above but explictly pointing to YAML path or URL |
| `faas-cli build . url-ing` | Context set to `.` but only building a specific function (thinking of the `samples.yml` with multiple fns). I'm a bit torn by this, probably better to avoid being clever. |
| **deploy verb** |
| `faas-cli deploy` | help for build verb |
| `faas-cli deploy --help` | as above |
| `faas-cli deploy .` | Deploys using the `Faasfile` in the PWD |
| `faas-cli deploy -f /path/to/yaml` | Deploys using the YAML at the specified path or URL |
| `faas-cli deploy -f /path/to/yaml ruby-echo` | as above but only deploys the specified function |
| `faas-cli deploy -f /path/to/yaml ruby-echo --force` | overwrites an existing function if it exists (default would be to warn that function already exists) |

Spitballing some new stuff..

| Command | Description |
| ------- | ----------- |
| **image verb** |
| `faas-cli image` | prints available subverbs below and global flags|
| *list sub-verb* |
| `faas-cli image list` | list all FaaS built images, would be based on build adding a magic label |
| *rm sub-verb* |
| `faas-cli image rm alexellis/faas-url-ping` | deletes the `alexellis/faas-url-ping` image only if it was created by FaaS, ie has a magic label |
| **function verb** |
| `faas-cli function` | prints available subverbs below and global flags|
| *list sub-verb* |
| `faas-cli function list` | list all running FaaS function containers, would be based on deploy adding a magic label |
| *rm sub-verb* |
| `faas-cli function rm shrink-image` | deletes the `func_shrink-image.xxxx` containers |
| *describe sub-verb* |
| `faas-cli function describe shrink-image` | describes the `func_shrink-image.xxxx` containers, could allow the user to add a description to the functions yaml definition that gets added as a label to either the image of the container |
| **provider verb** |
| `faas-cli provider list` | lists known FaaS providers, say `prod`/`staging`/`local` etc |
| `faas-cli provider add prod https://prod:8080 --network func_functions` | Adds a new provider called `prod`, likely cache this locally, perhaps in `~/.faas-cli/` |
| `faas-cli provider rm prod` | Remove above from the cache |
| `faas-cli provider login prod` | Prompt the user to authenticate with the provider, cache locally (Apache Brooklyns cli does something similar) |
| This could then enable stuff like.. |
| `faas-cli deploy -f /path/to/yaml prod` | override the provider in the YAML and deploy to `prod` |
