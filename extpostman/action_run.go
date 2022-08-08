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
	"github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-postman/utils"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

func RegisterHandlers() {
	exthttp.RegisterHttpHandler("/postman/collection/run", exthttp.GetterAsHandler(getActionDescription))
	exthttp.RegisterHttpHandler("/postman/collection/run/prepare", prepareCollectionRun)
	exthttp.RegisterHttpHandler("/postman/collection/run/start", startCollectionRun)
	exthttp.RegisterHttpHandler("/postman/collection/run/status", statusCollectionRun)
	exthttp.RegisterHttpHandler("/postman/collection/run/stop", stopCollectionRun)
}

func getActionDescription() attack_kit_api.AttackDescription {
	return attack_kit_api.AttackDescription{
		Id:          "com.github.steadybit.extension_postman.collection.run",
		Label:       "Postman (extension)",
		Description: "Integrate a Postman Collection via Postman Cloud API.",
		Version:     "0.0.1-SNAPSHOT",
		Kind:        attack_kit_api.CHECK,
		Icon:        extutil.Ptr(icon),
		TimeControl: attack_kit_api.INTERNAL,
		Parameters: []attack_kit_api.AttackParameter{
			{
				Name:         "duration",
				Label:        "Estimated duration",
				DefaultValue: extutil.Ptr("30s"),
				Description:  extutil.Ptr("As long as you have no timeout in place, the step will run as long as needed. You can set this estimation to size the step in the experiment editor for a better understanding of the time schedule."),
				Required:     extutil.Ptr(true),
				Type:         "duration",
			},
			{
				Name:        "apiKey",
				Label:       "API-Key",
				Description: extutil.Ptr("Postman Cloud API Key"),
				Required:    extutil.Ptr(true),
				Type:        "password",
			},
			{
				Name:        "collectionId",
				Label:       "Collection ID",
				Description: extutil.Ptr("UID of the Postman Collection"),
				Required:    extutil.Ptr(true),
				Type:        "string",
			},
			{
				Name:        "environmentId",
				Label:       "Environment ID",
				Description: extutil.Ptr("UID of the Postman Environment"),
				Required:    extutil.Ptr(false),
				Type:        "string",
			},
			{
				Name:        "environment",
				Label:       "Environment variables",
				Description: extutil.Ptr("Environment variables which will be passed to your Postman Collection"),
				Required:    extutil.Ptr(false),
				Type:        "key-value",
				Advanced:    extutil.Ptr(true),
			},
			{
				Name:         "iterations",
				Label:        "Iterations",
				Description:  extutil.Ptr("Number of iterations to run the collection"),
				Required:     extutil.Ptr(false),
				Type:         "integer",
				DefaultValue: extutil.Ptr("1"),
				Advanced:     extutil.Ptr(true),
			},
			{
				Name:         "timeout",
				Label:        "Timeout",
				Description:  extutil.Ptr("The time to wait for the entire collection run to complete execution. Hint: If you hit this timeout, no reports will be generated."),
				Required:     extutil.Ptr(false),
				Type:         "duration",
				DefaultValue: extutil.Ptr("1"),
				Advanced:     extutil.Ptr(true),
			},
			{
				Name:        "timeoutRequest",
				Label:       "Request Timeout",
				Description: extutil.Ptr("The Request Timeout for each request."),
				Required:    extutil.Ptr(false),
				Type:        "duration",
				Advanced:    extutil.Ptr(true),
			},
			{
				Name:        "verbose",
				Label:       "Verbose",
				Description: extutil.Ptr("Show detailed information of collection run and each request sent."),
				Required:    extutil.Ptr(false),
				Type:        "boolean",
				Advanced:    extutil.Ptr(true),
			},
			{
				Name:        "bail",
				Label:       "Bail",
				Description: extutil.Ptr("Stops the runner when a test case fails."),
				Required:    extutil.Ptr(false),
				Type:        "boolean",
				Advanced:    extutil.Ptr(true),
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
		Status: extutil.Ptr(attack_kit_api.MutatingEndpointReferenceWithCallInterval{
			Method: "POST",
			Path:   "/postman/collection/run/status",
		}),
		Stop: extutil.Ptr(attack_kit_api.MutatingEndpointReference{
			Method: "POST",
			Path:   "/postman/collection/run/stop",
		}),
	}
}

type State struct {
	Command         []string `json:"command"`
	Pid             int      `json:"pid"`
	Timestamp       string   `json:"timestamp"`
	StdOutLineCount int      `json:"stdOutLineCount"`
}

func prepareCollectionRun(w http.ResponseWriter, _ *http.Request, body []byte) {
	state, err := PrepareCollectionRun(body)
	if err != nil {
		exthttp.WriteError(w, *err)
	} else {
		utils.WriteAttackState(w, *state)
	}
}

func PrepareCollectionRun(body []byte) (*State, *extension_kit.ExtensionError) {
	var request attack_kit_api.PrepareAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to parse request body", err))
	}
	// create command
	var state State
	state.Timestamp = time.Now().Format(time.RFC3339)
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
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result-summary_%s.json", state.Timestamp))
	state.Command = append(state.Command, fmt.Sprintf("--reporter-htmlextra-export"))
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result_%s.html", state.Timestamp))
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
		exthttp.WriteError(w, *err)
	} else {
		utils.WriteAttackState(w, *state)
	}
}

func StartCollectionRun(_ context.Context, body []byte) (*State, *extension_kit.ExtensionError) {
	var request attack_kit_api.StartAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to parse request body", err))
	}

	var state State
	err = utils.DecodeAttackState(request.State, &state)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to parse attack state", err))
	}

	// start command
	log.Info().Msgf("Starting attack with command: %s", strings.Join(state.Command, " "))
	cmd := exec.Command(state.Command[0], state.Command[1:]...)
	outfile, err := os.Create(fmt.Sprintf("/tmp/newmanStdOut_%s.log", state.Timestamp))
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to create log file", err))
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
		err = ioutil.WriteFile(fmt.Sprintf("/tmp/newmanExitCode_%s", state.Timestamp), []byte(exitCode), 0644)
		if err != nil {
			log.Error().Msgf("Failed to write exit code to file: %s", err)
		}
	}()
	log.Info().Msgf("Started extension-postman")

	state.Command = nil
	return &state, nil
}

func statusCollectionRun(w http.ResponseWriter, r *http.Request, body []byte) {
	state, err := StatusCollectionRun(r.Context(), body)
	if err != nil {
		exthttp.WriteError(w, *err)
	} else {
		exthttp.WriteBody(w, state)
	}
}

func StatusCollectionRun(_ context.Context, body []byte) (*attack_kit_api.StatusResult, *extension_kit.ExtensionError) {

	var attackStatusRequest attack_kit_api.AttackStatusRequestBody
	err := json.Unmarshal(body, &attackStatusRequest)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to parse request body", err))
	}

	log.Info().Msgf("Checking collection run status for %s\n", attackStatusRequest)

	completed := false

	// check if postman is still running
	timestamp := attackStatusRequest.State["Timestamp"].(string)
	_, err = os.Stat(fmt.Sprintf("/tmp/newmanStdOut_%s.log", timestamp))
	if err != nil {
		log.Info().Msgf("Postman is still running")
	} else {
		log.Info().Msgf("Postman is not running anymore")
		completed = true
	}

	if completed {
		// read file with exit code
		exitCode, err := ioutil.ReadFile(fmt.Sprintf("/tmp/newmanExitCode_%s", timestamp))
		if err != nil {
			return nil, extutil.Ptr(extension_kit.ToError("Failed to open exit code file", err))
		}

		if string(exitCode) == "0" {
			log.Info().Msgf("Postman run completed successfully")
		} else {
			return nil, extutil.Ptr(extension_kit.ToError("Postman run failed", nil))
		}

	}

	var stdOutLineCount = int(attackStatusRequest.State["StdOutLineCount"].(float64))
	messages, lineCounter, err := getStdOutToMessages(stdOutLineCount, timestamp)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to get stdout", err))
	}
	attackStatusRequest.State["SendStdOutLine"] = lineCounter

	return &attack_kit_api.StatusResult{
		Completed: completed,
		State:     &attackStatusRequest.State,
		Messages:  extutil.Ptr(messages),
	}, nil
}

func getStdOutToMessages(stdOutLineCount int, timestamp string) ([]attack_kit_api.Message, int, error) {
	// send stdout and stderr to attack status
	var messages []attack_kit_api.Message

	// read file with stdout
	stdout, err := ioutil.ReadFile(fmt.Sprintf("/tmp/newmanStdOut_%s.log", timestamp))

	// iterate over stdout and convert to Message if it has not been already sent
	var lineCounter int
	for _, line := range strings.Split(string(stdout), "\n") {
		lineCounter++
		if lineCounter > stdOutLineCount {
			log.Debug().Msgf("Postman stdout: %s", line)
			messages = append(messages, attack_kit_api.Message{
				Level:   extutil.Ptr(attack_kit_api.Info),
				Message: line,
			})
		}
	}
	return messages, lineCounter, err
}

func stopCollectionRun(w http.ResponseWriter, r *http.Request, body []byte) {
	state, err := StopCollectionRun(r.Context(), body)
	if err != nil {
		exthttp.WriteError(w, *err)
	} else {
		exthttp.WriteBody(w, state)
	}
}

func StopCollectionRun(_ context.Context, body []byte) (*attack_kit_api.StopResult, *extension_kit.ExtensionError) {
	var request attack_kit_api.StopAttackRequestBody
	err := json.Unmarshal(body, &request)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to parse request body", err))
	}

	var state State
	err = utils.DecodeAttackState(request.State, &state)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to parse attack state", err))
	}

	var pid = state.Pid
	var timestamp = state.Timestamp
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to find process", err))
	}
	process.Kill()

	summary, err := file2Base64(fmt.Sprintf("/tmp/newman-result-summary_%s.json", timestamp))
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to open summary file", err))
	}

	html, err := file2Base64(fmt.Sprintf("/tmp/newman-result_%s.html", timestamp))
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to open html file", err))
	}

	var stdOutLineCount = state.StdOutLineCount
	messages, _, err := getStdOutToMessages(stdOutLineCount, timestamp)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to get stdout", err))
	}

	return &attack_kit_api.StopResult{
		Artifacts: extutil.Ptr([]attack_kit_api.Artifact{
			{
				Label: "$(experimentKey)_$(executionId)_postman.json",
				Data:  summary,
			}, {
				Label: "$(experimentKey)_$(executionId)_postman.html",
				Data:  html,
			},
		}),
		Messages: extutil.Ptr(messages),
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
