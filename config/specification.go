// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package config

import "time"

type Specification struct {
	PostmanBaseUrl                     string `json:"postmanBaseUrl" split_words:"true" required:"false" default:"https://api.getpostman.com"`
	PostmanApiKey                      string `json:"postmanApiKey" split_words:"true" required:"true"`
	PostmanCollectionDiscoveryInterval string `json:"postmanCollectionDiscoveryInterval" split_words:"true" required:"false" default:"3h"`
	// PostmanApiTimeout bounds a single attempt against the Postman API. It has to stay well
	// below the request budget the agent advertises via the Request-Timeout header (28s with
	// agent defaults), because overrunning that budget replaces whatever error we would have
	// reported with an opaque "503 Service Unavailable ... Timeout".
	PostmanApiTimeout time.Duration `json:"postmanApiTimeout" split_words:"true" required:"false" default:"8s"`
	// PostmanApiMaxAttempts is the total number of attempts (first try plus retries) made per
	// Postman API call before giving up. Attempts stop early once the shared budget is spent.
	PostmanApiMaxAttempts int `json:"postmanApiMaxAttempts" split_words:"true" required:"false" default:"3"`
}
