// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"context"
	"encoding/json"
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
	"github.com/steadybit/extension-postman/config"
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
	EnvironmentIdOrName string
	Environment         []map[string]string
	Verbose             bool
	Bail                bool
	Timeout             int
	TimeoutRequest      int
	Iterations          int
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
		Id:          targetID + ".run",
		Label:       "Postman",
		Description: "Integrate a Postman Collection via Postman Cloud API.",
		Version:     extbuild.GetSemverVersionStringOrUnknown(),
		Kind:        action_kit_api.Check,
		Icon:        extutil.Ptr(icon),
		TargetSelection: extutil.Ptr(action_kit_api.TargetSelection{
			// The target type this action is for
			TargetType: targetID,
			// You can provide a list of target templates to help the user select targets.
			// A template can be used to pre-fill a selection
			SelectionTemplates: extutil.Ptr([]action_kit_api.TargetSelectionTemplate{
				{
					Label: "by collection name",
					Query: "postman.collection.name=\"\"",
				},
				{
					Label: "by collection id",
					Query: "postman.collection.id=\"\"",
				},
			}),
		}),
		TimeControl: action_kit_api.TimeControlInternal,
		Parameters: []action_kit_api.ActionParameter{
			{
				Name:         "duration",
				Label:        "Estimated duration",
				DefaultValue: extutil.Ptr("30s"),
				Description:  extutil.Ptr("As long as you have no timeout in place, the step will run as long as needed. You can set this estimation to size the step in the experiment editor for a better understanding of the time schedule."),
				Required:     extutil.Ptr(true),
				Type:         action_kit_api.Duration,
			},
			{
				Name:        "environmentIdOrName",
				Label:       "Environment ID or Name",
				Description: extutil.Ptr("UID or unique Name of the Postman Environment"),
				Required:    extutil.Ptr(false),
				Type:        action_kit_api.String,
			},
			{
				Name:        "environment",
				Label:       "Environment variables",
				Description: extutil.Ptr("Environment variables which will be passed to your Postman Collection"),
				Required:    extutil.Ptr(false),
				Type:        action_kit_api.KeyValue,
				Advanced:    extutil.Ptr(true),
			},
			{
				Name:         "iterations",
				Label:        "Iterations",
				Description:  extutil.Ptr("Number of iterations to run the collection"),
				Required:     extutil.Ptr(false),
				Type:         action_kit_api.Integer,
				DefaultValue: extutil.Ptr("1"),
				Advanced:     extutil.Ptr(true),
			},
			{
				Name:        "timeout",
				Label:       "Timeout",
				Description: extutil.Ptr("The time to wait for the entire collection run to complete execution. Hint: If you hit this timeout, no reports will be generated."),
				Required:    extutil.Ptr(false),
				Type:        action_kit_api.Duration,
				Advanced:    extutil.Ptr(true),
			},
			{
				Name:        "timeoutRequest",
				Label:       "Request Timeout",
				Description: extutil.Ptr("The Request Timeout for each request."),
				Required:    extutil.Ptr(false),
				Type:        action_kit_api.Duration,
				Advanced:    extutil.Ptr(true),
			},
			{
				Name:        "verbose",
				Label:       "Verbose",
				Description: extutil.Ptr("Show detailed information of collection run and each request sent."),
				Required:    extutil.Ptr(false),
				Type:        action_kit_api.Boolean,
				Advanced:    extutil.Ptr(true),
			},
			{
				Name:        "bail",
				Label:       "Bail",
				Description: extutil.Ptr("Stops the runner when a test case fails."),
				Required:    extutil.Ptr(false),
				Type:        action_kit_api.Boolean,
				Advanced:    extutil.Ptr(true),
			},
		},
		Prepare: action_kit_api.MutatingEndpointReference{},
		Start:   action_kit_api.MutatingEndpointReference{},
		Status:  extutil.Ptr(action_kit_api.MutatingEndpointReferenceWithCallInterval{}),
		Stop:    extutil.Ptr(action_kit_api.MutatingEndpointReference{}),
	}
}

func (f PostmanAction) Prepare(_ context.Context, state *PostmanState, raw action_kit_api.PrepareActionRequestBody) (*action_kit_api.PrepareResult, error) {
	var request PostmanConfig
	if err := extconversion.Convert(raw.Config, &request); err != nil {
		return nil, extension_kit.ToError("Failed to unmarshal the config.", err)
	}
	config := config.Config

	collectionIds := raw.Target.Attributes["postman.collection.id"]
	if len(collectionIds) == 0 {
		return nil, extension_kit.ToError("No collection id provided", nil)
	}
	if len(collectionIds) > 1 {
		return nil, extension_kit.ToError("More than one collection id provided", nil)
	}
	var collectionId = collectionIds[0]

	state.Timestamp = time.Now().Format(time.RFC3339)
	state.Command = []string{
		"newman",
		"run",
		fmt.Sprintf("%s/collections/%s?apikey=%s", config.PostmanBaseUrl, collectionId, config.PostmanApiKey),
	}
	if request.EnvironmentIdOrName != "" {
		environmentId, error := GetPostEnvironmentId(request.EnvironmentIdOrName)
		if error != nil {
			return nil, extension_kit.ToError("Failed to get environment id.", error)
		}
		state.Command = append(state.Command, "--environment")
		state.Command = append(state.Command, fmt.Sprintf("%s/environments/%s?apikey=%s", config.PostmanBaseUrl, environmentId, config.PostmanApiKey))
	}
	if request.Environment != nil {
		for _, value := range request.Environment {
			state.Command = append(state.Command, "--env-var")
			state.Command = append(state.Command, fmt.Sprintf("%s=%s", value["key"], value["value"]))
		}
	}
	if request.Verbose {
		state.Command = append(state.Command, "--verbose")
	}
	if request.Bail {
		state.Command = append(state.Command, "--bail")
	}
	if request.Timeout > 0 {
		state.Command = append(state.Command, "--timeout")
		state.Command = append(state.Command, fmt.Sprintf("%d", request.Timeout))
	}
	if request.TimeoutRequest > 0 {
		state.Command = append(state.Command, "--timeout-request")
		state.Command = append(state.Command, fmt.Sprintf("%d", request.TimeoutRequest))
	}

	state.Command = append(state.Command, "--reporters")
	state.Command = append(state.Command, "cli,json-summary,htmlextra")
	state.Command = append(state.Command, "--reporter-summary-json-export")
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result-summary_%s.json", state.Timestamp))
	state.Command = append(state.Command, "--reporter-htmlextra-export")
	state.Command = append(state.Command, fmt.Sprintf("/tmp/newman-result_%s.html", state.Timestamp))
	state.Command = append(state.Command, "--reporter-htmlextra-omitResponseBodies")

	if request.Iterations > 1 {
		state.Command = append(state.Command, "-n")
		state.Command = append(state.Command, fmt.Sprintf("%d", request.Iterations))
	}
	log.Info().Msgf("Prepared action. Command: %s", extutil.MaskString(strings.Join(state.Command, " "), config.PostmanApiKey, 4))
	return nil, nil
}

func (f PostmanAction) Start(_ context.Context, state *PostmanState) (*action_kit_api.StartResult, error) {
	log.Info().Msgf("Starting newman!")
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
			Status: extutil.Ptr(action_kit_api.Errored),
			Title:  fmt.Sprintf("Postman run failed, exit-code %d", exitCode),
		}

		// check if summary file and try to check if it is a failure
		resultSummaryFileName := fmt.Sprintf("/tmp/newman-result-summary_%s.json", state.Timestamp)
		_, err = os.Stat(resultSummaryFileName)
		if err == nil { // file exists
			byteValue, _ := os.ReadFile(resultSummaryFileName)
			var report NewmanJsonReport
			jsonErr := json.Unmarshal(byteValue, &report)
			if jsonErr != nil {
				return nil, extutil.Ptr(extension_kit.ToError("Failed to parse report json", err))
			}
			if report.Run.Stats.Assertions != nil && report.Run.Stats.Assertions.Failed > 0 {
				result.Error = &action_kit_api.ActionKitError{
					Status: extutil.Ptr(action_kit_api.Failed),
					Title:  fmt.Sprintf("%d assertions failed", report.Run.Stats.Assertions.Failed),
				}
			} else if report.Run.Stats.Requests != nil && report.Run.Stats.Requests.Failed > 0 {
				result.Error = &action_kit_api.ActionKitError{
					Status: extutil.Ptr(action_kit_api.Failed),
					Title:  fmt.Sprintf("%d requests failed", report.Run.Stats.Requests.Failed),
				}
			}
		}

		result.Completed = true
	}

	messages := getStdOutMessages(cmdState.GetLines(false))
	log.Debug().Msgf("Returning %d messages", len(messages))

	result.Messages = extutil.Ptr(messages)
	return &result, nil
}

func getStdOutMessages(lines []string) []action_kit_api.Message {
	messages := make([]action_kit_api.Message, 0)
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
	extcmd.RemoveCmdState(state.CmdStateID)

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
	if exitCode != 0 && exitCode != -1 {
		messages = append(messages, action_kit_api.Message{
			Level:   extutil.Ptr(action_kit_api.Error),
			Message: fmt.Sprintf("Postman run failed with exit code %d", exitCode),
		})
	}

	var summaryFileContent string
	var htmlResultFileContent string

	artifacts := make([]action_kit_api.Artifact, 0)

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
