// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

/*
 * Copyright 2022 steadybit GmbH. All rights reserved.
 */

package extpostman

import (
	"context"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-postman/v2/config"
	"github.com/steadybit/extension-postman/v2/utils"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func TestPrepareCollectionRun(t *testing.T) {
	os.Setenv("STEADYBIT_EXTENSION_POSTMAN_API_KEY", "123456")
	config.ParseConfiguration()
	// Given
	requestBody := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration":            60000,
			"EnvironmentIdOrName": "5f757f0d-de24-462c-867f-256bb696d2dd",
			"iterations":          2,
			"environment": []map[string]string{
				{"key": "Test1", "value": "foo"},
				{"key": "Test2", "value": "bar"},
			},
			"timeout":        30000,
			"timeoutRequest": 30000,
			"verbose":        false,
			"bail":           true,
		},
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"postman.collection.id": {"645797"},
			},
		},
	})
	action := NewPostmanAction()
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)

	// Then
	assert.Nil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "newman", state.Command[0])
	assert.Equal(t, []string{"newman", "run", "https://api.getpostman.com/collections/645797?apikey=123456", "--environment", "https://api.getpostman.com/environments/5f757f0d-de24-462c-867f-256bb696d2dd?apikey=123456", "--env-var", "Test1=foo", "--env-var", "Test2=bar", "--bail", "--timeout", "30000", "--timeout-request", "30000", "--reporters", "cli,json-summary,htmlextra", "--reporter-summary-json-export", "--reporter-htmlextra-export", "--reporter-htmlextra-omitResponseBodies", "-n", "2"}, utils.RemoveAtIndex(utils.RemoveAtIndex(state.Command, 17), 18))
}

func TestPrepareCollectionRunWithEmptyEnvironment(t *testing.T) {
	os.Setenv("STEADYBIT_EXTENSION_POSTMAN_API_KEY", "123456")
	config.ParseConfiguration()
	// Given
	requestBody := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]interface{}{
			"duration":            60000,
			"EnvironmentIdOrName": "5f757f0d-de24-462c-867f-256bb696d2dd",
			"iterations":          2,
			"environment":         []map[string]string{},
			"timeout":             30000,
			"timeoutRequest":      30000,
			"verbose":             true,
		},
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{
				"postman.collection.id": {"645797"},
			},
		},
	})
	action := NewPostmanAction()
	state := action.NewEmptyState()

	// When
	result, err := action.Prepare(context.TODO(), &state, requestBody)

	// Then
	assert.Nil(t, err)
	assert.Nil(t, result)
	assert.Equal(t, "newman", state.Command[0])
	assert.Equal(t, []string{"newman", "run", "https://api.getpostman.com/collections/645797?apikey=123456", "--environment", "https://api.getpostman.com/environments/5f757f0d-de24-462c-867f-256bb696d2dd?apikey=123456", "--verbose", "--timeout", "30000", "--timeout-request", "30000", "--reporters", "cli,json-summary,htmlextra", "--reporter-summary-json-export", "--reporter-htmlextra-export", "--reporter-htmlextra-omitResponseBodies", "-n", "2"}, utils.RemoveAtIndex(utils.RemoveAtIndex(state.Command, 13), 14))
}
