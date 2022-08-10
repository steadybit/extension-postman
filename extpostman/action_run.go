// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/exthttp"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-postman/utils"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"sync"
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
				Name:        "timeout",
				Label:       "Timeout",
				Description: extutil.Ptr("The time to wait for the entire collection run to complete execution. Hint: If you hit this timeout, no reports will be generated."),
				Required:    extutil.Ptr(false),
				Type:        "duration",
				Advanced:    extutil.Ptr(true),
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

type InternalState struct {
	Cmd *exec.Cmd
	mu  *sync.Mutex
	out *bytes.Buffer
}

func (is *InternalState) Write(p []byte) (n int, err error) {
	is.mu.Lock()
	defer is.mu.Unlock()
	return is.out.Write(p)
}

func (is *InternalState) Lines() []string {
	is.mu.Lock()
	defer is.mu.Unlock()
	var result []string
	for line, err := is.out.ReadString('\n'); err == nil; line, err = is.out.ReadString('\n') {
		result = append(result, line)
	}
	return result
}

var internalStates = make(map[string]InternalState)

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

	var internalState InternalState
	internalState.out = new(bytes.Buffer)
	internalState.mu = new(sync.Mutex)
	internalState.Cmd = cmd

	//uncomment, if you want to see the output of the executed command
	//cmd.Stdout = io.MultiWriter(os.Stdout, cmdOut)
	cmd.Stdout = &internalState
	//cmd.Stderr = io.MultiWriter(os.Stderr, cmdOut)
	cmd.Stderr = &internalState
	cmd.Start()

	//internalState.out.
	//internalState.outScanner = bufio.NewScanner(cmdOut)
	//internalState.out = cmdOut
	internalStates[state.Timestamp] = internalState

	state.Pid = cmd.Process.Pid
	go func() {
		cmdErr := cmd.Wait()
		if cmdErr != nil {
			log.Error().Msgf("Failed to execute postman action: %s", cmdErr)
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

	timestamp := attackStatusRequest.State["Timestamp"].(string)
	internalState := internalStates[timestamp]

	// check if postman is still running
	exitCode := internalState.Cmd.ProcessState.ExitCode()
	if exitCode == -1 {
		log.Info().Msgf("Postman is still running")
	} else if exitCode == 0 {
		log.Info().Msgf("Postman run completed successfully")
		completed = true
	} else {
		//TODO add error message and do not throw a 500 here / change API to let the user know that the run failed?
		return nil, extutil.Ptr(extension_kit.ToError(fmt.Sprintf("Postman run failed, exit-code %d", exitCode), nil))
	}

	messages := getStdOutMessages(internalState.Lines())
	log.Debug().Msgf("Returning %d messages", len(messages))

	return &attack_kit_api.StatusResult{
		Completed: completed,
		State:     &attackStatusRequest.State,
		Messages:  extutil.Ptr(messages),
	}, nil
}

func getStdOutMessages(lines []string) []attack_kit_api.Message {
	var messages []attack_kit_api.Message
	for _, line := range lines {
		messages = append(messages, attack_kit_api.Message{
			Level:   extutil.Ptr(attack_kit_api.Info),
			Message: line,
		})
	}
	return messages
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

	var timestamp = state.Timestamp
	internalState := internalStates[timestamp]

	// kill postman if it is still running
	var pid = state.Pid
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to find process", err))
	}
	process.Kill()

	// read Stout and Stderr and send it as Messages
	messages := getStdOutMessages(internalState.Lines())

	// read return code and send it as Message
	exitCode := internalState.Cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		messages = append(messages, attack_kit_api.Message{
			Level:   extutil.Ptr(attack_kit_api.Error),
			Message: "Postman run failed with exit code " + string(exitCode),
		})
	}

	var summaryFileContent string
	var htmlResultFileContent string

	var artifacts []attack_kit_api.Artifact

	// check if summary file exists and send it as artifact
	const ResultSummaryFileName = "/tmp/newman-result-summary_%s.json"
	_, err = os.Stat(fmt.Sprintf(ResultSummaryFileName, timestamp))

	if err == nil { // file exists
		summaryFileContent, err = utils.File2Base64(fmt.Sprintf(ResultSummaryFileName, timestamp))
		if err != nil {
			return nil, extutil.Ptr(extension_kit.ToError("Failed to open summaryFileContent file", err))
		}
		artifacts = append(artifacts, attack_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_postman.json",
			Data:  summaryFileContent,
		})
	}

	// check if html result file exists and send it as artifact
	const ResultFileName = "/tmp/newman-result_%s.html"
	_, err = os.Stat(fmt.Sprintf(ResultFileName, timestamp))

	if err == nil { // file exists
		htmlResultFileContent, err = utils.File2Base64(fmt.Sprintf(ResultFileName, timestamp))
		if err != nil {
			return nil, extutil.Ptr(extension_kit.ToError("Failed to open htmlResultFileContent file", err))
		}
		artifacts = append(artifacts, attack_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_postman.html",
			Data:  htmlResultFileContent,
		})
	}

	log.Debug().Msgf("Returning %d messages", len(messages))
	return &attack_kit_api.StopResult{
		Artifacts: extutil.Ptr(artifacts),
		Messages:  extutil.Ptr(messages),
	}, nil
}
