// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"sync/atomic"
	"syscall"
	"testing"
	"time"

	"github.com/steadybit/action-kit/go/action_kit_api/v2"
	"github.com/steadybit/extension-kit/extutil"
	"github.com/steadybit/extension-postman/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testCollectionId = "645797"

// configureApiStub points the extension at srv and bounds attempts so the timing-sensitive
// tests stay quick.
func configureApiStub(t *testing.T, url string, timeout string, maxAttempts string) {
	t.Helper()
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_API_KEY", "123456")
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_BASE_URL", url)
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_API_TIMEOUT", timeout)
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_API_MAX_ATTEMPTS", maxAttempts)
	config.ParseConfiguration()
}

func prepareRequest(collectionId string, actionConfig map[string]any) action_kit_api.PrepareActionRequestBody {
	if actionConfig == nil {
		actionConfig = map[string]any{"duration": 60000}
	}
	return extutil.JsonMangle(action_kit_api.PrepareActionRequestBody{
		Config: actionConfig,
		Target: &action_kit_api.Target{
			Attributes: map[string][]string{"postman.collection.id": {collectionId}},
		},
	})
}

// TestPrepareFailsWithinBudgetWhenPostmanApiStalls is the regression test for the nightly
// failure: a stalling Postman API used to hold the prepare handler past the budget the agent
// advertises via Request-Timeout, so the platform saw "503 Service Unavailable ... Timeout"
// instead of the real cause. Prepare must now give up inside the budget and say why.
func TestPrepareFailsWithinBudgetWhenPostmanApiStalls(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-r.Context().Done()
	}))
	t.Cleanup(server.Close)
	configureApiStub(t, server.URL, "200ms", "2")

	action := PostmanAction{}
	state := action.NewEmptyState()

	// The agent's Request-Timeout header becomes a deadline on the request context.
	budget := 2 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()

	start := time.Now()
	_, err := action.Prepare(ctx, &state, prepareRequest(testCollectionId, nil))
	elapsed := time.Since(start)
	t.Cleanup(func() { _ = os.RemoveAll(state.WorkDir) })

	require.Error(t, err)
	assert.Less(t, elapsed, budget, "prepare must fail inside the request budget, not be cut off by the handler timeout")
	assert.NoError(t, ctx.Err(), "the request budget must not have been exhausted")
}

func TestPostmanApiRetriesTransientFailures(t *testing.T) {
	for _, status := range []int{http.StatusInternalServerError, http.StatusTooManyRequests, http.StatusBadGateway} {
		t.Run(fmt.Sprintf("status %d", status), func(t *testing.T) {
			var calls atomic.Int32
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if calls.Add(1) == 1 {
					w.WriteHeader(status)
					return
				}
				_, _ = w.Write([]byte(`{"collection":{"info":{"name":"test"},"item":[]}}`))
			}))
			t.Cleanup(server.Close)
			configureApiStub(t, server.URL, "2s", "3")

			action := PostmanAction{}
			state := action.NewEmptyState()

			_, err := action.Prepare(context.Background(), &state, prepareRequest(testCollectionId, nil))
			t.Cleanup(func() { _ = os.RemoveAll(state.WorkDir) })

			require.NoError(t, err, "a transient %d must be retried", status)
			assert.Equal(t, int32(2), calls.Load())
			assert.FileExists(t, state.Command[2])
		})
	}
}

func TestPostmanApiDoesNotRetryClientErrors(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)
	configureApiStub(t, server.URL, "2s", "3")

	action := PostmanAction{}
	state := action.NewEmptyState()

	_, err := action.Prepare(context.Background(), &state, prepareRequest(testCollectionId, nil))
	t.Cleanup(func() { _ = os.RemoveAll(state.WorkDir) })

	require.Error(t, err)
	assert.Equal(t, int32(1), calls.Load(), "a 404 will not pass on its own and must not be retried")
}

// TestGetPostEnvironmentIdDistinguishesApiFailureFromMissingEnvironment guards the misleading
// error this used to produce: any failure listing environments was swallowed and reported as
// if the configured environment name did not exist.
func TestGetPostEnvironmentIdDistinguishesApiFailureFromMissingEnvironment(t *testing.T) {
	t.Run("api failure", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
		}))
		t.Cleanup(server.Close)
		configureApiStub(t, server.URL, "2s", "1")

		_, err := GetPostEnvironmentId(context.Background(), "dev")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to list environments")
		assert.NotContains(t, err.Error(), "failed to find environment with name")
	})

	t.Run("environment really is missing", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte(`{"environments":[]}`))
		}))
		t.Cleanup(server.Close)
		configureApiStub(t, server.URL, "2s", "1")

		_, err := GetPostEnvironmentId(context.Background(), "dev")

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to find environment with name 'dev'")
	})
}

func TestGetPostEnvironmentIdAcceptsUuidWithoutCallingTheApi(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
	}))
	t.Cleanup(server.Close)
	configureApiStub(t, server.URL, "2s", "1")

	id, err := GetPostEnvironmentId(context.Background(), "70cb2138-3443-4c33-a45c-73477a5fd903")

	require.NoError(t, err)
	assert.Equal(t, "70cb2138-3443-4c33-a45c-73477a5fd903", id)
	assert.Equal(t, int32(0), calls.Load())
}

// TestStatusReportsSignalKilledNewmanAsCompleted covers newman being terminated by a signal.
// ExitCode() is -1 for that as well as for "still running", and the branch telling them apart
// had its Completed flag overwritten, so the action never finished on its own.
func TestStatusReportsSignalKilledNewmanAsCompleted(t *testing.T) {
	action := PostmanAction{}
	state := PostmanState{Command: []string{"sleep", "60"}}

	_, err := action.Start(context.Background(), &state)
	require.NoError(t, err)
	require.NotZero(t, state.Pid)

	require.NoError(t, syscall.Kill(state.Pid, syscall.SIGKILL))

	var result *action_kit_api.StatusResult
	require.Eventually(t, func() bool {
		result, err = action.Status(context.Background(), &state)
		return err == nil && result != nil && result.Completed
	}, 10*time.Second, 50*time.Millisecond, "a signal-killed newman must complete the action")

	require.NotNil(t, result.Error)
	require.NotNil(t, result.Error.Status)
	assert.Equal(t, action_kit_api.Errored, *result.Error.Status)
	assert.True(t, strings.HasPrefix(result.Error.Title, "Postman process is not running anymore"))
}
