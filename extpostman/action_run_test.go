// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

/*
 * Copyright 2022 steadybit GmbH. All rights reserved.
 */

package extpostman

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-postman/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newPostmanApiStub(t *testing.T) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// The API key must be passed via the header, never as a query parameter.
		assert.Equal(t, "123456", r.Header.Get("X-API-Key"))
		assert.Empty(t, r.URL.Query().Get("apikey"))

		switch {
		case strings.HasPrefix(r.URL.Path, "/collections/"):
			_, _ = w.Write([]byte(`{"collection":{"info":{"name":"test"},"item":[]}}`))
		case strings.HasPrefix(r.URL.Path, "/environments/"):
			_, _ = w.Write([]byte(`{"environment":{"id":"5f757f0d-de24-462c-867f-256bb696d2dd","name":"env","values":[]}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func assertNoSecretInCommand(t *testing.T, command []string) {
	t.Helper()
	for _, arg := range command {
		assert.NotContains(t, arg, "123456", "api key must not appear in the command")
		assert.NotContains(t, arg, "apikey", "apikey query parameter must not appear in the command")
	}
}

func TestPrepareCollectionRun(t *testing.T) {
	server := newPostmanApiStub(t)
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_API_KEY", "123456")
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_BASE_URL", server.URL)
	config.ParseConfiguration()

	// Given
	requestBody := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]any{
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
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(state.WorkDir) })

	// Then
	assert.Nil(t, result)
	assertNoSecretInCommand(t, state.Command)
	assert.Equal(t, "newman", state.Command[0])
	assert.Equal(t, "run", state.Command[1])
	assert.True(t, strings.HasPrefix(state.Command[2], state.WorkDir))
	assert.True(t, strings.HasSuffix(state.Command[2], "collection.json"))

	// the collection and environment are downloaded to local files, not fetched by newman via url
	assert.FileExists(t, state.Command[2])
	assert.Contains(t, state.Command, "--environment")
	assert.Contains(t, state.Command, "--env-var")
	assert.Contains(t, state.Command, "Test1=foo")
	assert.Contains(t, state.Command, "--bail")
	assert.Contains(t, state.Command, "-n")
}

func TestPrepareCollectionRunWithEmptyEnvironment(t *testing.T) {
	server := newPostmanApiStub(t)
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_API_KEY", "123456")
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_BASE_URL", server.URL)
	config.ParseConfiguration()

	// Given
	requestBody := extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: map[string]any{
			"duration":       60000,
			"iterations":     2,
			"environment":    []map[string]string{},
			"timeout":        30000,
			"timeoutRequest": 30000,
			"verbose":        true,
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
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(state.WorkDir) })

	// Then
	assert.Nil(t, result)
	assertNoSecretInCommand(t, state.Command)
	assert.Equal(t, "newman", state.Command[0])
	assert.Equal(t, "run", state.Command[1])
	assert.True(t, strings.HasSuffix(state.Command[2], "collection.json"))
	assert.NotContains(t, state.Command, "--environment")
	assert.Contains(t, state.Command, "--verbose")
}
