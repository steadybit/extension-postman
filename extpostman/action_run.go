// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"encoding/json"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-postman/utils"
	"net/http"
)

func RegisterHandlers() {
	utils.RegisterHttpHandler("/postman/collection/run", utils.GetterAsHandler(getActionDescription))
	utils.RegisterHttpHandler("/postman/collection/run/prepare", prepareCollectionRun)
	//utils.RegisterHttpHandler("/postman/collection/run/start", startCollectionRun)
}

func getActionDescription() attack_kit_api.AttackDescription {
	return attack_kit_api.AttackDescription{
		Id:          "com.github.steadybit.extension_postman.collection.run",
		Label:       "Postman",
		Description: "Integrate a Postman Collection via Postman Cloud API.",
		Version:     "1.0.0",
		Icon:        attack_kit_api.Ptr(icon),
		TargetType:  "host",               //FIXME - nil
		Category:    attack_kit_api.State, //FIXME - "Custom integrations"
		TimeControl: attack_kit_api.INTERNAL,
		Parameters: []attack_kit_api.AttackParameter{
			{
				Name:         "duration",
				Label:        "Estimated duration",
				DefaultValue: attack_kit_api.Ptr("30s"),
				Description:  attack_kit_api.Ptr("As long as you have no timeout in place, the step will run as long as needed. You can set this estimation to size the step in the experiment editor for a better understanding of the time schedule."),
				Required:     attack_kit_api.Ptr(true),
				Type:         "duration",
			},
			{
				Name:        "apiKey",
				Label:       "API-Key",
				Description: attack_kit_api.Ptr("Postman Cloud API Key"),
				Required:    attack_kit_api.Ptr(true),
				Type:        "password",
			},
			{
				Name:        "collectionId",
				Label:       "Collection ID",
				Description: attack_kit_api.Ptr("UID of the Postman Collection"),
				Required:    attack_kit_api.Ptr(true),
				Type:        "string",
			},
			{
				Name:        "environmentId",
				Label:       "Environment ID",
				Description: attack_kit_api.Ptr("UID of the Postman Environment"),
				Required:    attack_kit_api.Ptr(false),
				Type:        "string",
			},
			{
				Name:        "environment",
				Label:       "Environment variables",
				Description: attack_kit_api.Ptr("Environment variables which will be passed to your Postman Collection"),
				Required:    attack_kit_api.Ptr(false),
				Type:        "key-value",
				Advanced:    attack_kit_api.Ptr(true),
			},
			{
				Name:         "iterations",
				Label:        "Iterations",
				Description:  attack_kit_api.Ptr("Number of iterations to run the collection"),
				Required:     attack_kit_api.Ptr(false),
				Type:         "integer",
				DefaultValue: attack_kit_api.Ptr("1"),
				Advanced:     attack_kit_api.Ptr(true),
			},
			{
				Name:         "timeout",
				Label:        "Timeout",
				Description:  attack_kit_api.Ptr("The time to wait for the entire collection run to complete execution. Hint: If you hit this timeout, no reports will be generated."),
				Required:     attack_kit_api.Ptr(false),
				Type:         "duration",
				DefaultValue: attack_kit_api.Ptr("1"),
				Advanced:     attack_kit_api.Ptr(true),
			},
			{
				Name:        "timeoutRequest",
				Label:       "Request Timeout",
				Description: attack_kit_api.Ptr("The Request Timeout for each request."),
				Required:    attack_kit_api.Ptr(false),
				Type:        "duration",
				Advanced:    attack_kit_api.Ptr(true),
			},
			{
				Name:        "verbose",
				Label:       "Verbose",
				Description: attack_kit_api.Ptr("Show detailed information of collection run and each request sent."),
				Required:    attack_kit_api.Ptr(false),
				Type:        "boolean",
				Advanced:    attack_kit_api.Ptr(true),
			},
			{
				Name:        "bail",
				Label:       "Bail",
				Description: attack_kit_api.Ptr("Stops the runner when a test case fails."),
				Required:    attack_kit_api.Ptr(false),
				Type:        "boolean",
				Advanced:    attack_kit_api.Ptr(true),
			},
		},
		Prepare: attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/postman/collection/run/prepare",
		},
		Start: attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/postman/collection/run/start",
		},
	}
}

type State struct {
}

func prepareCollectionRun(w http.ResponseWriter, _ *http.Request, body []byte) {
	state, err := PrepareCollectionRun(body)
	if err != nil {
		utils.WriteError(w, *err)
	} else {
		utils.WriteAttackState(w, *state)
	}
}

func PrepareCollectionRun(body []byte) (*State, *attack_kit_api.AttackKitError) {
	var request attack_kit_api.PrepareAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse request body", err))
	}

	return attack_kit_api.Ptr(State{}), nil
}
