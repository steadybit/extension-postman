# Changelog

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
