# Using template from external repository

## Repository structure

The external repository must have a directory named ```template``` at the root directory, in which there are directories
containing templates. The directory for each template can be freely named with alphanumeric characters and hyphen.

Example:

```
template
├── csharp
│   ├── Dockerfile
│   └── template.yml
├── dockerfile
│   ├── Dockerfile
│   └── template.yml
├── node
│   ├── Dockerfile
│   └── template.yml
├── node-armhf
│   ├── Dockerfile
│   └── template.yml
├── python
│   ├── Dockerfile
│   └── template.yml
├── python-armhf
│   ├── Dockerfile
│   └── template.yml
├── php5
│   ├── Dockerfile
│   └── template.yml
├── php7
│   ├── Dockerfile
│   └── template.yml
└── ruby
    ├── Dockerfile
    └── template.yml
```

## template.yml schema

* `language` - template name i.e. `node`
* `fprocess` - optional, fprocess for watchdog
* `build_options` - array, optional to define a slice of `string, []string` to provide a number of build options and package installed for the named package

    Example:

    ```yaml
    build_options:
      - name: curl-tls
        packages:
        - curl
        - ca-certificates
    ```
* `welcome_message` - printed after `faas-cli new`, populate with a link to the user guide or how to add a module for package manager
* `handler_folder` - where to copy the function's build context into the Docker image, usually just `function`


## Download external repository

In order to build functions using 3rd party templates, you need to add 3rd templates before the build step, with the following command:

```bash
faas-cli template pull https://github.com/openfaas-incubator/golang-http-template
```

If you need to update the downloaded repository, just add the flag `--overwrite` to the download command:

```bash
faas-cli template pull https://github.com/openfaas-incubator/golang-http-template --override
```

You can specify the template URL with `OPENFAAS_TEMPLATE_URL` environmental variable. CLI overrides the environmental variable.

```bash
export OPENFAAS_TEMPLATE_URL="https://github.com/openfaas-incubator/golang-http-template"
faas-cli template pull
```

## Pin the template repository version

You may specify the branch or tag pulled by adding a URL fragment with the branch or tag name. For example, to pull the `1.0` tag of the default template repository, use

```bash
faas-cli template pull https://github.com/openfaas/templates#1.0
```

If a branch or tag is not specified, the repositories default branch is pulled (usually `master`).


## List locally available languages

```bash
faas-cli new --list
```

## Check template store

In order to check what templates are available in the template store type

```bash
faas-cli template store list
```

Pull the desired template by specifying `NAME` attribute only:

```bash
faas-cli template store pull go
```

or pull the template by mixing the repository and name the following way:

```bash
faas-cli template store pull openfaas/go
```

If you have your own store with templates, you can set that as your default official store by setting the environmental variable `OPENFAAS_TEMPLATE_STORE_URL` the following way:

```bash
export OPENFAAS_TEMPLATE_STORE_URL=https://raw.githubusercontent.com/user/openfaas-templates/templates.json
```

Now the source of the store is changed to the URL you have specified above.

To get specific information for a template use the following command:

```bash
faas-cli template store describe golang-middleware
```

or use the source and the name in case of name collision:

```bash
faas-cli template store describe openfaas-incubator/golang-middleware
```
