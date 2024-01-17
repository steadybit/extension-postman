// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/action-kit/go/action_kit_api/v2"
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	target, err := e2e.PollForTarget(ctx, e, "com.steadybit.extension_postman.collection", func(target discovery_kit_api.Target) bool {
		log.Info().Msgf("Checking target: %s", target)
		return e2e.HasAttribute(target, "steadybit.label", "shopping-demo")
	})

	require.NoError(t, err)
	assert.Equal(t, target.TargetType, "com.steadybit.extension_postman.collection")
	assert.Equal(t, target.Attributes["postman.collection.id"], []string{collectionId})
	assert.Equal(t, target.Attributes["postman.collection.name"], []string{"shopping-demo"})
}

func testRunPostman(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct {
	}{}

	target := action_kit_api.Target{
		Attributes: map[string][]string{
			"postman.collection.id": {collectionId},
		},
	}

	exec, err := e.RunAction("com.steadybit.extension_postman.collection.run.v2", &target, config, nil)
	require.NoError(t, err)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Starting newman!", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Postman run completed successfully", 90*time.Second)
	require.NoError(t, exec.Cancel())
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

	exec, err := e.RunAction("com.steadybit.extension_postman.collection.run.v2", &target, config, nil)
	require.NoError(t, err)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "--environment", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Starting newman!", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Postman run completed successfully", 90*time.Second)
	require.NoError(t, exec.Cancel())
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

	exec, err := e.RunAction("com.steadybit.extension_postman.collection.run.v2", &target, config, nil)
	require.NoError(t, err)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "--environment", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Starting newman!", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Postman run completed successfully", 90*time.Second)
	require.NoError(t, exec.Cancel())
}
