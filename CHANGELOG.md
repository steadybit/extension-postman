# Changelog

## v2.0.16

- Updated dependencies

## v2.0.15

- Updated dependencies

## v2.0.14

- Dummy Release to test creating a Github Release

## v2.0.13

- Updated dependencies
- Added a label for the postman collection name target attribubte

## v2.0.12

- Updated dependencies

## v2.0.11

- Fix Postman API access

## v2.0.10

- Updated dependencies

## v2.0.9

- update dependencies (CVE-2024-21538)

## v2.0.8

- update dependencies
- Use uid instead of name for user statement in Dockerfile

## v2.0.7

- Update newman dependency

## v2.0.6

- Set new `Technology` property in extension description
- Update dependencies (go 1.23)

## v2.0.5

- update dependencies

## v2.0.4

- Update dependencies (go 1.22)
- Node 18.20.2 - [CVE-2024-27980](https://www.cve.org/CVERecord?id=CVE-2024-27980)

## v2.0.3

- update dependencies

## v2.0.2

- update dependencies

## v2.0.1

- update dependencies

## v2.0.0

- Breaking Changes
- Configure ApiKey in the extension configuration
- Discovery of Postman Collections
- Support for Postman Environments to select the correct environment for the collection by name
- Use a Postman Collection a target for the action

## v1.5.8

- Update dependencies

## v1.5.7

- Update dependencies

## v1.5.6

- Added `pprof` endpoints for debugging purposes
- Update dependencies

## v1.5.4

- update dependencies
- added https_proxy support documentation

## v1.5.3

- migration to new unified steadybit actionIds and targetTypes

## v1.5.2

- update dependencies

## v1.5.1

 - Fail on assertion and request failures, error out on other errors

## v1.5.0

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
