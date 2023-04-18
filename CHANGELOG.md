# Changelog

## v1.5.0.

 - Update dependencies
 - Refactored to use action-kit-sdk

## v1.4.1

 - Renamed the action, removed the `(extension)`

## v1.4.0

 - Print build information on extension startup.

## v1.3.0

 - Support creation of a TLS server through the environment variables `STEADYBIT_EXTENSION_TLS_SERVER_CERT` and `STEADYBIT_EXTENSION_TLS_SERVER_KEY`. Both environment variables must refer to files containing the certificate and key in PEM format.
 - Support mutual TLS through the environment variable `STEADYBIT_EXTENSION_TLS_CLIENT_CAS`. The environment must refer to a comma-separated list of files containing allowed clients' CA certificates in PEM format.

## v1.2.0

- Support for the `STEADYBIT_LOG_FORMAT` env variable. When set to `json`, extensions will log JSON lines to stderr.

## v1.1.3

 - Define language-related environment variables in Docker image for consistency to the original `postman/newman` Docker image.

## v1.1.2

 - The `postman/newman` Docker image does not support the linux/arm64 platform.

## v1.1.1
 - Improved error handling

## v1.1.0

 - The log level can now be configured through the `STEADYBIT_LOG_LEVEL` environment variable.

## v1.0.0

 - Initial release
