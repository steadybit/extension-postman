// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/action-kit/go/action_kit_test/client"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/steadybit/discovery-kit/go/discovery_kit_api"
	"github.com/steadybit/discovery-kit/go/discovery_kit_test/validate"
	"github.com/steadybit/extension-kit/extlogging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"strings"
	"testing"
	"time"
)

func TestWithMinikube(t *testing.T) {
	extlogging.InitZeroLog()
	server := createMockPostmanServer()
	defer server.Close()
	split := strings.SplitAfter(server.URL, ":")
	port := split[len(split)-1]

	extFactory := e2e.HelmExtensionFactory{
		Name: "extension-postman",
		Port: 8086,
		ExtraArgs: func(m *e2e.Minikube) []string {
			return []string{
				"--set", "logging.level=debug",
				"--set", "postman.apiKey=testApiKey",
				"--set", "extraEnv[0].name=STEADYBIT_EXTENSION_POSTMAN_BASE_URL",
				"--set", fmt.Sprintf("extraEnv[0].value=%s:%s", "http://host.minikube.internal", port),
			}
		},
	}

	e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "validate discovery",
			Test: validateDiscovery,
		},
		{
			Name: "target discovery",
			Test: testDiscovery,
		},
		{
			Name: "run postman",
			Test: testRunPostman,
		},
		{
			Name: "run postman with env name",
			Test: testRunPostmanWithEnvName,
		},
		{
			Name: "run postman with env id",
			Test: testRunPostmanWithEnvId,
		},
	})
}

func validateDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	assert.NoError(t, validate.ValidateEndpointReferences("/", e.Client))
}

func testDiscovery(t *testing.T, _ *e2e.Minikube, e *e2e.Extension) {
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	target, err := e2e.PollForTarget(ctx, e, "com.steadybit.extension_postman.collection", func(target discovery_kit_api.Target) bool {
		log.Info().Msgf("Checking target: %s", target)
		return e2e.HasAttribute(target, "postman.collection.name", "shopping-demo")
	})

	require.NoError(t, err)
	assert.Equal(t, target.TargetType, "com.steadybit.extension_postman.collection")
	assert.Equal(t, target.Attributes["postman.collection.id"], []string{collectionId})
}

// assertPostmanArtifacts checks the two files a finished collection run hands back.
func assertPostmanArtifacts(t *testing.T, exec client.ActionExecution) {
	t.Helper()
	artifacts := make(map[string][]byte)
	for _, artifact := range exec.Artifacts() {
		data, err := base64.StdEncoding.DecodeString(artifact.Data)
		require.NoError(t, err, "artifact %s must be valid base64", artifact.Label)
		artifacts[artifact.Label] = data
	}

	require.Contains(t, artifacts, "$(experimentKey)_$(executionId)_postman.json")
	require.Contains(t, artifacts, "$(experimentKey)_$(executionId)_postman.html")
	assert.True(t, bytes.HasPrefix(artifacts["$(experimentKey)_$(executionId)_postman.json"], []byte("{")),
		"the summary artifact must be the json newman wrote")
	assert.Contains(t, string(artifacts["$(experimentKey)_$(executionId)_postman.html"]), "<html",
		"the report artifact must be the html newman wrote")
}

func testRunPostman(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct {
	}{}

	target := action_kit_api.Target{
		Attributes: map[string][]string{
			"postman.collection.id": {collectionId},
		},
	}

	exec, err := e.RunAction("com.steadybit.extension_postman.collection.run", &target, config, nil)
	require.NoError(t, err)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Starting newman!", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Postman run completed successfully", 210*time.Second)
	// Stop kills newman and then only attaches result files it finds, so cancelling here could hand
	// back nothing.
	require.NoError(t, exec.Wait())
	assertPostmanArtifacts(t, exec)
}

func testRunPostmanWithEnvId(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct {
		EnvironmentIdOrName string
	}{
		EnvironmentIdOrName: "70cb2138-3443-4c33-a45c-73477a5fd903",
	}

	target := action_kit_api.Target{
		Attributes: map[string][]string{
			"postman.collection.id": {collectionId},
		},
	}

	exec, err := e.RunAction("com.steadybit.extension_postman.collection.run", &target, config, nil)
	require.NoError(t, err)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "--environment", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Starting newman!", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Postman run completed successfully", 210*time.Second)
	// Stop kills newman and then only attaches result files it finds, so cancelling here could hand
	// back nothing.
	require.NoError(t, exec.Wait())
	assertPostmanArtifacts(t, exec)
}

func testRunPostmanWithEnvName(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct {
		EnvironmentIdOrName string
	}{
		EnvironmentIdOrName: "dev",
	}

	target := action_kit_api.Target{
		Attributes: map[string][]string{
			"postman.collection.id": {collectionId},
		},
	}

	exec, err := e.RunAction("com.steadybit.extension_postman.collection.run", &target, config, nil)
	require.NoError(t, err)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "--environment", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Starting newman!", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Postman run completed successfully", 210*time.Second)
	// Stop kills newman and then only attaches result files it finds, so cancelling here could hand
	// back nothing.
	require.NoError(t, exec.Wait())
	assertPostmanArtifacts(t, exec)
}
