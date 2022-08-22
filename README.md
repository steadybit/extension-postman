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

| Environment Variable                   |
|----------------------------------------|
| No additional configuration parameters |

## Deployment

We recommend that you deploy the extension with our [official Helm chart](https://github.com/steadybit/helm-charts/tree/main/charts/steadybit-extension-postman)
.

## Agent Configuration

The Steadybit agent needs to be configured to interact with the postman extension by adding the following environment variables:

```shell
# Make sure to adapt the URLs and indices in the environment variables names as necessary for your setup

STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL=http://steadybit-extension-postman.steadybit-extension.svc.cluster.local:8086
```

When leveraging our official Helm charts, you can set the configuration through additional environment variables on the agent:

```
--set agent.env[0].name=STEADYBIT_AGENT_ACTIONS_EXTENSIONS_0_URL \
--set agent.env[0].value="http://steadybit-extension-postman.steadybit-extension.svc.cluster.local:8086" \
```
