/*
 * Copyright 2022 steadybit GmbH. All rights reserved.
 */

package extpostman

import (
	"encoding/json"
	"github.com/steadybit/attack-kit/go/attack_kit_api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func removeAtIndex(s []string, index int) []string {
	return append(s[:index], s[index+1:]...)
}

func TestPrepareCollectionRun(t *testing.T) {
	// Given
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Config: map[string]interface{}{
			"duration":      "60s",
			"apiKey":        "123456",
			"collectionId":  "645797",
			"environmentId": "env1",
			"iterations":    2,
			"environment": []map[string]string{
				{"key": "Test1", "value": "foo"},
				{"key": "Test2", "value": "bar"},
			},
			"timeout":        30000,
			"timeoutRequest": 30000,
			"verbose":        false,
			"bail":           true,
		},
		Target: &attack_kit_api.Target{},
	}
	requestBodyJson, err := json.Marshal(requestBody)
	require.Nil(t, err)

	// When
	state, attackErr := PrepareCollectionRun(requestBodyJson)

	// Then
	assert.Nil(t, attackErr)
	assert.Equal(t, "newman", state.Command[0])
	assert.Equal(t, []string([]string{"newman", "run", "https://api.getpostman.com/collections/645797?apikey=123456", "--environment", "https://api.getpostman.com/environments/env1?apikey=123456", "-env-var", "Test1=foo", "-env-var", "Test2=bar", "--verbose", "--bail", "--timeout", "30000", "--timeout-request", "30000", "--reporters", "cli,json-summary,htmlextra", "--reporter-summary-json-export", "--reporter-htmlextra-export", "--reporter-htmlextra-omitResponseBodies", "-n", "2"}), removeAtIndex(removeAtIndex(state.Command, 18), 19))
}

func TestPrepareCollectionRunWithEmptyEnvironment(t *testing.T) {
	// Given
	requestBody := attack_kit_api.PrepareAttackRequestBody{
		Config: map[string]interface{}{
			"duration":       "60s",
			"apiKey":         "123456",
			"collectionId":   "645797",
			"environmentId":  "env1",
			"iterations":     2,
			"environment":    []map[string]string{},
			"timeout":        30000,
			"timeoutRequest": 30000,
			"verbose":        false,
			"bail":           true,
		},
		Target: &attack_kit_api.Target{},
	}
	requestBodyJson, err := json.Marshal(requestBody)
	require.Nil(t, err)

	// When
	state, attackErr := PrepareCollectionRun(requestBodyJson)

	// Then
	assert.Nil(t, attackErr)
	assert.Equal(t, "newman", state.Command[0])
	assert.Equal(t, []string([]string{"newman", "run", "https://api.getpostman.com/collections/645797?apikey=123456", "--environment", "https://api.getpostman.com/environments/env1?apikey=123456", "--verbose", "--bail", "--timeout", "30000", "--timeout-request", "30000", "--reporters", "cli,json-summary,htmlextra", "--reporter-summary-json-export", "--reporter-htmlextra-export", "--reporter-htmlextra-omitResponseBodies", "-n", "2"}), removeAtIndex(removeAtIndex(state.Command, 14), 15))
}
