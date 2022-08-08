# AttackKit Go API

This module exposes Go types that you will find helpful when implementing an AttackKit extension.

The types are generated automatically from the AttackKit [OpenAPI specification](https://github.com/steadybit/attack-kit/tree/main/openapi).

## Installation

Add the following to your `go.mod` file:

```
go get github.com/steadybit/attack-kit/go/attack_kit_api@v0.1.0
```

## Usage

```go
import (
	"github.com/steadybit/attack-kit/go/attack_kit_api"
)

attackList := attack_kit_api.AttackList{
    Attacks: []attack_kit_api.DescribingEndpointReference{
        {
            "GET",
            "/attacks/rollout-restart",
        },
    },
}
```