// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2023 Steadybit GmbH

package e2e

import (
	"fmt"
	"github.com/steadybit/action-kit/go/action_kit_test/e2e"
	"github.com/steadybit/extension-kit/extlogging"
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
				"--set", "extraEnv[0].name=STEADYBIT_EXTENSION_POSTMAN_BASE_URL",
				"--set", fmt.Sprintf("extraEnv[0].value=%s:%s", "http://host.minikube.internal:", port),
			}
		},
	}

	e2e.WithDefaultMinikube(t, &extFactory, []e2e.WithMinikubeTestCase{
		{
			Name: "run postman",
			Test: testRunPostman,
		},
	})
}

func testRunPostman(t *testing.T, m *e2e.Minikube, e *e2e.Extension) {
	config := struct {
		CollectionId string
		ApiKey       string
	}{
		CollectionId: "testCollectionId",
		ApiKey:       "testApiKey",
	}

	exec, err := e.RunAction("com.steadybit.extension_postman.collection.run", nil, config, nil)
	require.NoError(t, err)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Starting newman!", 90*time.Second)
	e2e.AssertLogContainsWithTimeout(t, m, e.Pod, "Postman run completed successfully", 90*time.Second)
	require.NoError(t, exec.Cancel())
}
