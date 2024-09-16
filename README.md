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

### Kubernetes

Detailed information about agent and extension installation in kubernetes can also be found in
our [documentation](https://docs.steadybit.com/install-and-configure/install-agent/install-on-kubernetes).

#### Recommended (via agent helm chart)

All extensions provide a helm chart that is also integrated in the
[helm-chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-agent) of the agent.

You must provide additional values to activate this extension.

```
--set extension-postman.enabled=true \
--set extension-postman.postman.apiKey={{YOUR_POSTMAN_API_KEY}} \
```

Additional configuration options can be found in
the [helm-chart](https://github.com/steadybit/extension-postman/blob/main/charts/steadybit-extension-postman/values.yaml) of the
extension.

#### Alternative (via own helm chart)

If you need more control, you can install the extension via its
dedicated [helm-chart](https://github.com/steadybit/extension-postman/blob/main/charts/steadybit-extension-postman).

```bash
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

### Linux Package

This extension is currently not available as a Linux package.

## Extension registration

Make sure that the extension is registered with the agent. In most cases this is done automatically. Please refer to
the [documentation](https://docs.steadybit.com/install-and-configure/install-agent/extension-discovery) for more
information about extension registration and how to verify.

## Proxy
To communicate to Postman via a proxy, we need the environment variable `https_proxy` to be set.
This can be set via helm using the extraEnv variable

```bash
--set "extraEnv[0].name=HTTPS_PROXY" \
--set "extraEnv[0].value=https:\\user:pwd@CompanyProxy.com:8888"
```
