// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-postman/utils"
	"net/http"
)

func RegisterHandlers() {
	utils.RegisterHttpHandler("/postman/collection/run", utils.GetterAsHandler(getActionDescription))
	utils.RegisterHttpHandler("/postman/collection/run/prepare", prepareCollectionRun)
	utils.RegisterHttpHandler("/postman/collection/run/start", startCollectionRun)
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
	Command []string `json:"command"`
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
	// create command
	var state State
	state.Command = []string{
		"newman",
		"run",
		fmt.Sprintf("https://api.getpostman.com/collections/%s?apikey=%s", request.Config["collectionId"], request.Config["apiKey"]),
	}
	if request.Config["environmentId"] != "" {
		state.Command = append(state.Command, fmt.Sprintf("--environment"))
		state.Command = append(state.Command, fmt.Sprintf("https://api.getpostman.com/environments/%s?apikey=%s", request.Config["environmentId"], request.Config["apiKey"]))
	}
	if request.Config["environment"] != nil {
		for key, value := range request.Config["environment"].(map[string]interface{}) {
			if value != nil {
				state.Command = append(state.Command, fmt.Sprintf("-env-var"))
				state.Command = append(state.Command, fmt.Sprintf("%s=%s", key, value))
			}
		}
	}
	if request.Config["verbose"] != nil {
		state.Command = append(state.Command, fmt.Sprintf("--verbose"))
	}
	if request.Config["bail"] != nil {
		state.Command = append(state.Command, fmt.Sprintf("--bail"))
	}
	if request.Config["timeout"] != nil {
		state.Command = append(state.Command, fmt.Sprintf("--timeout"))
		state.Command = append(state.Command, fmt.Sprintf("%.0f", request.Config["timeout"].(float64)))
	}
	if request.Config["timeoutRequest"] != nil {
		state.Command = append(state.Command, fmt.Sprintf("--timeout-request"))
		state.Command = append(state.Command, fmt.Sprintf("%.0f", request.Config["timeoutRequest"].(float64)))
	}

	state.Command = append(state.Command, fmt.Sprintf("--reporters"))
	state.Command = append(state.Command, fmt.Sprintf("cli,json-summary,htmlextra"))
	state.Command = append(state.Command, fmt.Sprintf("--reporter-summary-json-export"))
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result-summary.json"))
	state.Command = append(state.Command, fmt.Sprintf("--reporter-htmlextra-export"))
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result.html"))
	state.Command = append(state.Command, fmt.Sprintf("--reporter-htmlextra-omitResponseBodies"))

	if request.Config["iterations"] != nil && request.Config["iterations"].(float64) > 1 {
		state.Command = append(state.Command, fmt.Sprintf("-n"))
		state.Command = append(state.Command, fmt.Sprintf("%.0f", request.Config["iterations"].(float64)))
	}

	return &state, nil
}

func startCollectionRun(w http.ResponseWriter, r *http.Request, body []byte) {
	state, err := StartCollectionRun(r.Context(), body)
	if err != nil {
		utils.WriteError(w, *err)
	} else {
		utils.WriteAttackState(w, *state)
	}
}

func StartCollectionRun(ctx context.Context, body []byte) (*State, *attack_kit_api.AttackKitError) {
	var request attack_kit_api.StartAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse request body", err))
	}

	var state State
	err = utils.DecodeAttackState(request.State, &state)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse attack state", err))
	}

	return &state, nil
}
