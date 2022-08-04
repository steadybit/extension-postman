// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-postman/utils"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
)

func RegisterHandlers() {
	utils.RegisterHttpHandler("/postman/collection/run", utils.GetterAsHandler(getActionDescription))
	utils.RegisterHttpHandler("/postman/collection/run/prepare", prepareCollectionRun)
	utils.RegisterHttpHandler("/postman/collection/run/start", startCollectionRun)
	utils.RegisterHttpHandler("/postman/collection/run/status", statusCollectionRun)
	utils.RegisterHttpHandler("/postman/collection/run/stop", stopCollectionRun)
}

func getActionDescription() attack_kit_api.AttackDescription {
	return attack_kit_api.AttackDescription{
		Id:          "com.github.steadybit.extension_postman.collection.run",
		Label:       "Postman",
		Description: "Integrate a Postman Collection via Postman Cloud API.",
		Version:     "0.0.1",
		Icon:        attack_kit_api.Ptr(icon),
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
		Status: attack_kit_api.Ptr(attack_kit_api.MutatingEndpointReferenceWithCallInterval{
			Method: "POST",
			Path:   "/postman/collection/run/status",
		}),
		Stop: attack_kit_api.Ptr(attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/postman/collection/run/stop",
		}),
	}
}

type State struct {
	Command []string `json:"command"`
	Pid     int      `json:"pid"`
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
		for _, value := range request.Config["environment"].([]interface{}) {
			cast := value.(map[string]interface{})
			state.Command = append(state.Command, fmt.Sprintf("-env-var"))
			state.Command = append(state.Command, fmt.Sprintf("%s=%s", cast["key"], cast["value"]))
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

func StartCollectionRun(_ context.Context, body []byte) (*State, *attack_kit_api.AttackKitError) {
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

	// start command
	log.Info().Msgf("Starting attack with command: %s", strings.Join(state.Command, " "))
	cmd := exec.Command(state.Command[0], state.Command[1:]...)
	outfile, err := os.Create("/tmp/newmanStdOut.log")
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to create log file", err))
	}
	cmd.Stdout = outfile
	cmd.Stderr = outfile
	cmd.Start()
	state.Pid = cmd.Process.Pid
	go func() {
		cmdErr := cmd.Wait()
		outfile.Close()
		if cmdErr != nil {
			log.Error().Msgf("Failed to execute postman action: %s", cmdErr)
		}
		var exitCode string
		exitCode = fmt.Sprintf("%d", cmd.ProcessState.ExitCode())
		err = ioutil.WriteFile("/tmp/newmanExitCode", []byte(exitCode), 0644)
		if err != nil {
			log.Error().Msgf("Failed to write exit code to file: %s", err)
		}
	}()
	log.Info().Msgf("Started extension-postman")

	return &state, nil
}

func statusCollectionRun(w http.ResponseWriter, r *http.Request, body []byte) {
	state, err := StatusCollectionRun(r.Context(), body)
	if err != nil {
		utils.WriteError(w, *err)
	} else {
		utils.WriteBody(w, state)
	}
}

func StatusCollectionRun(_ context.Context, body []byte) (*attack_kit_api.StatusResult, *attack_kit_api.AttackKitError) {

	var attackStatusRequest attack_kit_api.AttackStatusRequestBody
	err := json.Unmarshal(body, &attackStatusRequest)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse request body", err))
	}

	log.Info().Msgf("Checking collection run status for %s\n", attackStatusRequest)

	completed := false

	// check if postman is still running
	_, err = os.Stat("/tmp/newmanStdOut.log")
	if err != nil {
		log.Info().Msgf("Postman is still running")
	} else {
		log.Info().Msgf("Postman is not running anymore")
		completed = true
	}

	return &attack_kit_api.StatusResult{
		Completed: completed,
	}, nil
}

func stopCollectionRun(w http.ResponseWriter, r *http.Request, body []byte) {
	state, err := StopCollectionRun(r.Context(), body)
	if err != nil {
		utils.WriteError(w, *err)
	} else {
		utils.WriteBody(w, state)
	}
}

func StopCollectionRun(_ context.Context, body []byte) (*attack_kit_api.StopResult, *attack_kit_api.AttackKitError) {
	var request attack_kit_api.StopAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse request body", err))
	}

	var state State
	err = utils.DecodeAttackState(request.State, &state)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to parse attack state", err))
	}

	process, err := os.FindProcess(state.Pid)
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to find process", err))
	}
	process.Kill()

	summary, err := file2Base64("/tmp/newman-result-summary.json")
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to open summary file", err))
	}

	html, err := file2Base64("/tmp/newman-result.html")
	if err != nil {
		return nil, attack_kit_api.Ptr(utils.ToError("Failed to open html file", err))
	}

	return &attack_kit_api.StopResult{
		Artifacts: attack_kit_api.Ptr([]attack_kit_api.Artifact{
			{
				Label: "$(experimentKey)_$(executionId)_postman.json.tar",
				Data:  summary,
			}, {
				Label: "$(experimentKey)_$(executionId)_postman.html.tar",
				Data:  html,
			},
		}),
	}, nil
}

func file2Base64(file string) (string, error) {
	f, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer f.Close()

	var buffer bytes.Buffer
	_, err = io.Copy(&buffer, f)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(buffer.Bytes()), nil
}
