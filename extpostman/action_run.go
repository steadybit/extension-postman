// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_sdk"
	"github.com/steadybit/extension-kit"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-kit/extcmd"
	"github.com/steadybit/extension-kit/extconversion"
	"github.com/steadybit/extension-kit/extfile"
	"github.com/steadybit/extension-kit/extutil"
	"os"
	"os/exec"
	"strings"
	"time"
)

type PostmanAction struct {
}

type PostmanState struct {
	Command         []string `json:"command"`
	Pid             int      `json:"pid"`
	CmdStateID      string   `json:"cmdStateId"`
	Timestamp       string   `json:"timestamp"`
	StdOutLineCount int      `json:"stdOutLineCount"`
}

type PostmanConfig struct {
	CollectionId   string
	ApiKey         string
	EnvironmentId  string
	Environment    []map[string]string
	Verbose        bool
	Bail           bool
	Timeout        int
	TimeoutRequest int
	Iterations     int
}

func NewPostmanAction() action_kit_sdk.Action[PostmanState] {
	return PostmanAction{}
}

// Make sure PostmanAction implements all required interfaces
var _ action_kit_sdk.Action[PostmanState] = (*PostmanAction)(nil)
var _ action_kit_sdk.ActionWithStatus[PostmanState] = (*PostmanAction)(nil)
var _ action_kit_sdk.ActionWithStop[PostmanState] = (*PostmanAction)(nil)

func (f PostmanAction) NewEmptyState() PostmanState {
	return PostmanState{}
}

func (f PostmanAction) Describe() action_kit_api.ActionDescription {
	return action_kit_api.ActionDescription{
		Id:          "com.github.steadybit.extension_postman.collection.run",
		Label:       "Postman",
		Description: "Integrate a Postman Collection via Postman Cloud API.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Kind:        action_kit_api.Check,
		Icon:        extutil.Ptr(icon),
		TimeControl: action_kit_api.Internal,
		Parameters: []action_kit_api.ActionParameter{
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
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Status:  extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{}),
		Stop:    extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (f PostmanAction) Prepare(_ context.Context, state *PostmanState, request action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	var config PostmanConfig
	if err := extconversion.Convert(request.Config, &config); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the config.", err)
	}

	state.Timestamp = time.Now().Format(time.RFC3339)
	state.Command = []string{
		"newman",
		"run",
		fmt.Sprintf("https://api.getpostman.com/collections/%s?apikey=%s", config.CollectionId, config.ApiKey),
	}
	if config.EnvironmentId != "" {
		state.Command = append(state.Command, "--environment")
		state.Command = append(state.Command, fmt.Sprintf("https://api.getpostman.com/environments/%s?apikey=%s", config.EnvironmentId, config.ApiKey))
	}
	if config.Environment != nil {
		for _, value := range config.Environment {
			state.Command = append(state.Command, "-env-var")
			state.Command = append(state.Command, fmt.Sprintf("%s=%s", value["key"], value["value"]))
		}
	}
	if config.Verbose {
		state.Command = append(state.Command, "--verbose")
	}
	if config.Bail {
		state.Command = append(state.Command, "--bail")
	}
	if config.Timeout > 0 {
		state.Command = append(state.Command, "--timeout")
		state.Command = append(state.Command, fmt.Sprintf("%d", config.Timeout))
	}
	if config.TimeoutRequest > 0 {
		state.Command = append(state.Command, "--timeout-request")
		state.Command = append(state.Command, fmt.Sprintf("%d", config.TimeoutRequest))
	}

	state.Command = append(state.Command, "--reporters")
	state.Command = append(state.Command, "cli,json-summary,htmlextra")
	state.Command = append(state.Command, "--reporter-summary-json-export")
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result-summary_%s.json", state.Timestamp))
	state.Command = append(state.Command, "--reporter-htmlextra-export")
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result_%s.html", state.Timestamp))
	state.Command = append(state.Command, "--reporter-htmlextra-omitResponseBodies")

	if config.Iterations > 1 {
		state.Command = append(state.Command, "-n")
		state.Command = append(state.Command, fmt.Sprintf("%d", config.Iterations))
	}
	return nil, nil
}

func (f PostmanAction) Start(_ context.Context, state *PostmanState) (*action_kit_api.StartResult, error) {
	log.Info().Msgf("Starting action with command: %s", strings.Join(state.Command, " "))
	cmd := exec.Command(state.Command[0], state.Command[1:]...)
	cmdState := extcmd.NewCmdState(cmd)
	state.CmdStateID = cmdState.Id
	err := cmd.Start()
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to start command.", err))
	}

	state.Pid = cmd.Process.Pid
	go func() {
		cmdErr := cmd.Wait()
		if cmdErr != nil {
			log.Error().Msgf("Failed to execute postman action: %s", cmdErr)
		}
	}()
	log.Info().Msgf("Started extension-postman")

	state.Command = nil
	return nil, nil
}

func (f PostmanAction) Status(_ context.Context, state *PostmanState) (*action_kit_api.StatusResult, error) {
	log.Info().Msgf("Checking collection run status for %d\n", state.Pid)

	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to find command state", err))
	}

	var result action_kit_api.StatusResult

	// check if postman is still running
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	if exitCode == -1 {
		log.Info().Msgf("Postman is still running")
		result.Completed = false
	} else if exitCode == 0 {
		log.Info().Msgf("Postman run completed successfully")
		result.Completed = true
	} else {
		result.Error = &action_kit_api.ActionKitError{
			Status: extutil.Ptr(action_kit_api.Failed),
			Title:  fmt.Sprintf("Postman run failed, exit-code %d", exitCode),
		}
		result.Completed = true
	}

	messages := getStdOutMessages(cmdState.GetLines(false))
	log.Debug().Msgf("Returning %d messages", len(messages))

	result.Messages = extutil.Ptr(messages)
	return &result, nil
}

func getStdOutMessages(lines []string) []action_kit_api.Message {
	var messages []action_kit_api.Message
	for _, line := range lines {
		messages = append(messages, action_kit_api.Message{
			Level:   extutil.Ptr(action_kit_api.Info),
			Message: line,
		})
	}
	return messages
}

func (f PostmanAction) Stop(_ context.Context, state *PostmanState) (*action_kit_api.StopResult, error) {
	var timestamp = state.Timestamp
	cmdState, err := extcmd.GetCmdState(state.CmdStateID)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to find command state", err))
	}

	// kill postman if it is still running
	var pid = state.Pid
	process, err := os.FindProcess(pid)
	if err != nil {
		return nil, extutil.Ptr(extension_kit.ToError("Failed to find process", err))
	}
	_ = process.Kill()

	// read Stout and Stderr and send it as Messages
	messages := getStdOutMessages(cmdState.GetLines(true))

	// read return code and send it as Message
	exitCode := cmdState.Cmd.ProcessState.ExitCode()
	if exitCode != 0 {
		messages = append(messages, action_kit_api.Message{
			Level:   extutil.Ptr(action_kit_api.Error),
			Message: fmt.Sprintf("Postman run failed with exit code %d", exitCode),
		})
	}

	var summaryFileContent string
	var htmlResultFileContent string

	var artifacts []action_kit_api.Artifact

	// check if summary file exists and send it as artifact
	const ResultSummaryFileName = "/tmp/newman-result-summary_%s.json"
	_, err = os.Stat(fmt.Sprintf(ResultSummaryFileName, timestamp))

	if err == nil { // file exists
		summaryFileContent, err = extfile.File2Base64(fmt.Sprintf(ResultSummaryFileName, timestamp))
		if err != nil {
			return nil, extutil.Ptr(extension_kit.ToError("Failed to open summaryFileContent file", err))
		}
		artifacts = append(artifacts, action_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_postman.json",
			Data:  summaryFileContent,
		})
	}

	// check if html result file exists and send it as artifact
	const ResultFileName = "/tmp/newman-result_%s.html"
	_, err = os.Stat(fmt.Sprintf(ResultFileName, timestamp))

	if err == nil { // file exists
		htmlResultFileContent, err = extfile.File2Base64(fmt.Sprintf(ResultFileName, timestamp))
		if err != nil {
			return nil, extutil.Ptr(extension_kit.ToError("Failed to open htmlResultFileContent file", err))
		}
		artifacts = append(artifacts, action_kit_api.Artifact{
			Label: "$(experimentKey)_$(executionId)_postman.html",
			Data:  htmlResultFileContent,
		})
	}

	log.Debug().Msgf("Returning %d messages", len(messages))
	return &action_kit_api.StopResult{
		Artifacts: extutil.Ptr(artifacts),
		Messages:  extutil.Ptr(messages),
	}, nil
}
