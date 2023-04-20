<img src="./postman.png" width="300" align="right" alt="Kong logo">

# Steadybit extension-postman

A [Steadybit](https://www.steadybit.com/) extension to execute [Postman](https://www.postman.com/) collections via Postman Cloud Api.

:warning: The Postman extension is currently bundled in the out-of-the-box version of steadybit. This code should help you to understand the usage
of [Action kit](https://github.com/steadybit/action-kit).

## Prerequisites

- A Postman account with a
	valid [API Key](https://www.postman.com/postman/workspace/postman-public-workspace/documentation/12959542-c8142d51-e97c-46b6-bd77-52bb66712c9a#authentication)
	is needed to execute collections.

## Configuration

| Environment Variable                  | Meaning                                                                                                                                                                | Default                 |
|---------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------------|
| `STEADYBIT_EXTENSION_ROBOT_NAMES`     | Comma-separated list of discoverable robots                                                                                                                            | Bender,Terminator,R2-D2 |
| `STEADYBIT_EXTENSION_PORT`            | Port number that the HTTP server should bind to.                                                                                                                       | 8080                    |
| `STEADYBIT_EXTENSION_TLS_SERVER_CERT` | Optional absolute path to a TLS certificate that will be used to open an **HTTPS** server.                                                                             |                         |
| `STEADYBIT_EXTENSION_TLS_SERVER_KEY`  | Optional absolute path to a file containing the key to the server certificate.                                                                                         |                         |
| `STEADYBIT_EXTENSION_TLS_CLIENT_CAS`  | Optional comma-separated list of absolute paths to files containing TLS certificates. When specified, the server will expect clients to authenticate using mutual TLS. |                         |
| `STEADYBIT_LOG_FORMAT`                | Defines the log format that the extension will use. Possible values are `text` and `json`.                                                                             | text                    |
| `STEADYBIT_LOG_LEVEL`                 | Defines the active log level. Possible values are `debug`, `info`, `warn` and `error`.                                                                                 | info                    |

## Running the Extension

### Using Docker

```sh
$ docker run \
  --rm \
  -p 8085 \
  --name steadybit-extension-postman \
  ghcr.io/steadybit/extension-postman:latest
```

### Using Helm in Kubernetes

```sh
$ helm repo add steadybit-extension-postman https://steadybit.github.io/extension-postman
$ helm repo update
$ helm upgrade steadybit-extension-postman \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-extension \
    steadybit-extension-postman/steadybit-extension-postman
```

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.

