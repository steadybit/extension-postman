<img src="./postman.png" width="300" align="right" alt="Kong logo">

# Steadybit extension-postman

A [Steadybit](https://www.steadybit.com/) extension to execute [Postman](https://www.postman.com/) collections via Postman Cloud Api.

Learn about the capabilities of this extension in our [Reliability Hub](https://hub.steadybit.com/extension/com.steadybit.extension_postman).

## Prerequisites

- A Postman account with a
	valid [API Key](https://www.postman.com/postman/workspace/postman-public-workspace/documentation/12959542-c8142d51-e97c-46b6-bd77-52bb66712c9a#authentication)
	is needed to execute collections.

## Configuration

Postman_Api_Key
## Configuration

| Environment Variable                  | Helm value             | Meaning                                                     | Required | Default |
|---------------------------------------|------------------------|-------------------------------------------------------------|----------|---------|
| `HTTPS_PROXY`                         | via extraEnv variables | Configure the proxy to be used for Postman communication.   | no       |         |
| `STEADYBIT_EXTENSION_POSTMAN_API_KEY` | postman.apiKey         | Configure the api-key to be used for Postman communication. | yes      |         |

The extension supports all environment variables provided by [steadybit/extension-kit](https://github.com/steadybit/extension-kit#environment-variables).

## Installation

### Using Docker

```sh
docker run \
  --rm \
  -p 8085 \
  --name steadybit-extension-postman \
  ghcr.io/steadybit/extension-postman:latest
```

### Using Helm in Kubernetes

```sh
helm repo add steadybit-extension-postman https://steadybit.github.io/extension-postman
helm repo update
helm upgrade steadybit-extension-postman \
    --install \
    --wait \
    --timeout 5m0s \
    --create-namespace \
    --namespace steadybit-agent \
    --set postman.apiKey=<YOUR_API_KEY> \
    steadybit-extension-postman/steadybit-extension-postman
```

## Register the extension

Make sure to register the extension at the steadybit platform. Please refer to
the [documentation](https://docs.steadybit.com/integrate-with-steadybit/extensions/extension-installation) for more information.

## Proxy
To communicate to Postman via a proxy, we need the environment variable `https_proxy` to be set.
This can be set via helm using the extraEnv variable

```bash
--set "extraEnv[0].name=HTTPS_PROXY" \
--set "extraEnv[0].value=https:\\user:pwd@CompanyProxy.com:8888"
```
