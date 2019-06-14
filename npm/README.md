# @openfaas/faas-cli [![NPM version](https://img.shields.io/npm/v/@openfaas/faas-cli.svg?style=flat)](https://www.npmjs.com/package/@openfaas/faas-cli) [![NPM monthly downloads](https://img.shields.io/npm/dm/@openfaas/faas-cli.svg?style=flat)](https://npmjs.org/package/@openfaas/faas-cli) [![NPM total downloads](https://img.shields.io/npm/dt/@openfaas/faas-cli.svg?style=flat)](https://npmjs.org/package/@openfaas/faas-cli)

> OpenFaaS CLI

## Install globally

Install with [npm](https://www.npmjs.com/) as a global module:

```sh
$ npm install --global @openfaas/faas-cli
```

## Install locally

Install with [npm](https://www.npmjs.com/) as a local dev dependency:

```sh
$ npm install --save-dev @openfaas/faas-cli
```

## Usage globally

```sh
$ faas
```

## Usage locally

When installed locally, `faas` can easily be used in npm scripts in the `package.json`

```json
{
  "scripts": {
    "faas": "faas",
    "build": "faas build",
    "push": "faas push",
    "deploy": "faas deploy",
    "up": "faas up"
  }
}
```

The scripts can be run from the commandline in the local folder with the following commands:

```sh
# shows the help information
$ npm run faas

# shows the list of templates in the template store
$ npm run faas template store list

# builds the function containers
$ npm run build

# pushes the built containers to a container registry
$ npm run push

# deploys the function to the openfaas gateway
$ npm run deploy

# combines the last 3 commands into one to do build, push, and deploy together
$ npm run up
```

See the [main faas-cli README.md](https://github.com/openfaas/faas-cli) for more information on `faas` commands.
