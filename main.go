// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package main

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-postman/utils"
	"net/http"
)

func main() {
	utils.RegisterHttpHandler("/", utils.GetterAsHandler(getAttackList))

	port := 8086
	log.Info().Msgf("Starting extension-postman server on port %d. Get started via /\n", port)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

type ExtensionListResponse struct {
	Attacks []attack_kit_api.DescribingEndpointReference `json:"attacks"`
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
