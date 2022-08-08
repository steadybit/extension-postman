// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"fmt"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-postman/extpostman"
	"net/http"
	"os"
)

func main() {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	exthttp.RegisterHttpHandler("/", exthttp.GetterAsHandler(getAttackList))

	extpostman.RegisterHandlers()

	port := 8086
	log.Info().Msgf("Starting extension-postman server on port %d. Get started via /\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
	if err != nil {
		log.Fatal().Err(err).Msgf("Failed to start the http server of the postman extension")
	}
}

func getAttackList() attack_kit_api.AttackList {
	return attack_kit_api.AttackList{
		Attacks: []attack_kit_api.DescribingEndpointReference{
			{
				"GET",
				"/postman/collection/run",
			},
		},
	}
}
