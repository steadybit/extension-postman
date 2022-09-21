// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/steadybit/extension-postman/extpostman"
	"net/http"
)

func main() {
	extlogging.InitZeroLog()

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getActionList))

	extpostman.RegisterHandlers()

	port := 8086
	log.Info().Msgf("Starting extension-postman server on port %d. Get started via /\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to start the http server of the postman extension")
	}
}

func getActionList() action_kit_api.ActionList {
	return action_kit_api.ActionList{
		Actions: []action_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/postman/collection/run",
			},
		},
	}
}
