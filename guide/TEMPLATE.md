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
├── php-56
│   ├── Dockerfile
│   └── template.yml
├── php-71
│   ├── Dockerfile
│   └── template.yml
├── php-latest
│   ├── Dockerfile
│   └── template.yml
└── ruby
    ├── Dockerfile
    └── template.yml
```

## Download external repository

In order to build functions using 3rd party templates, you need to add 3rd templates before the build step, with the following command:

```
./faas-cli template pull https://github.com/itscaro/openfaas-template-php
```

If you need to update the downloaded repository, just add the flag `--overwrite` to the download command:

```
./faas-cli template pull https://github.com/itscaro/openfaas-template-php --override
```
