// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2026 Steadybit GmbH

package extpostman

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/steadybit/extension-postman/v2/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A single Postman 500 used to hide all collections until the next discovery interval.
func TestCollectionDiscoveryRecoversBeforeTheNextInterval(t *testing.T) {
	var calls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if calls.Add(1) == 1 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"collections":[{"id":"c63c9f2a","name":"shopping-demo"}]}`))
	}))
	t.Cleanup(server.Close)
	configureApiStub(t, server.URL, "2s", "1")
	t.Setenv("STEADYBIT_EXTENSION_POSTMAN_COLLECTION_DISCOVERY_RETRY_INTERVAL", "10ms")
	config.ParseConfiguration()

	discovery := NewPostmanCollectionDiscovery()

	require.EventuallyWithT(t, func(c *assert.CollectT) {
		targets, err := discovery.DiscoverTargets(context.Background())
		assert.NoError(c, err)
		if assert.Len(c, targets, 1) {
			assert.Equal(c, "shopping-demo", targets[0].Label)
		}
	}, time.Second, 5*time.Millisecond)
}
