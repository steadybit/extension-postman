// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-postman/v2/config"
)

const (
	targetID = "com.steadybit.extension_postman.collection"
	icon     = "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iMjQiIGhlaWdodD0iMjQiIHZpZXdCb3g9IjAgMCAyNCAyNCIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cGF0aCBkPSJNMTMuNTI3LjA5OUM2Ljk1NS0uNzQ0Ljk0MiAzLjkuMDk5IDEwLjQ3M2MtLjg0MyA2LjU3MiAzLjggMTIuNTg0IDEwLjM3MyAxMy40MjggNi41NzMuODQzIDEyLjU4Ny0zLjgwMSAxMy40MjgtMTAuMzc0QzI0Ljc0NCA2Ljk1NSAyMC4xMDEuOTQzIDEzLjUyNy4wOTl6bTIuNDcxIDcuNDg1YS44NTUuODU1IDAgMDAtLjU5My4yNWwtNC40NTMgNC40NTMtLjMwNy0uMzA3LS42NDMtLjY0M2M0LjM4OS00LjM3NiA1LjE4LTQuNDE4IDUuOTk2LTMuNzUzem0tNC44NjMgNC44NjFsNC40NC00LjQ0YS42Mi42MiAwIDExLjg0Ny45MDNsLTQuNjk5IDQuMTI1LS41ODgtLjU4OHptLjMzLjY5NGwtMS4xLjIzOGEuMDYuMDYgMCAwMS0uMDY3LS4wMzIuMDYuMDYgMCAwMS4wMS0uMDczbC42NDUtLjY0NS41MTIuNTEyem0tMi44MDMtLjQ1OWwxLjE3Mi0xLjE3Mi44NzkuODc4LTEuOTc5LjQyNmEuMDc0LjA3NCAwIDAxLS4wODUtLjAzOS4wNzIuMDcyIDAgMDEuMDEzLS4wOTN6bS0zLjY0NiA2LjA1OGEuMDc2LjA3NiAwIDAxLS4wNjktLjA4My4wNzcuMDc3IDAgMDEuMDIyLS4wNDZoLjAwMmwuOTQ2LS45NDYgMS4yMjIgMS4yMjItMi4xMjMtLjE0N3ptMi40MjUtMS4yNTZhLjIyOC4yMjggMCAwMC0uMTE3LjI1NmwuMjAzLjg2NWEuMTI1LjEyNSAwIDAxLS4yMTEuMTE3aC0uMDAzbC0uOTM0LS45MzQtLjI5NC0uMjk1IDMuNzYyLTMuNzU4IDEuODItLjM5My44NzQuODc0Yy0xLjI1NSAxLjEwMi0yLjk3MSAyLjIwMS01LjEgMy4yNjh6bTUuMjc5LTMuNDI4aC0uMDAybC0uODM5LS44MzkgNC42OTktNC4xMjVhLjk1Mi45NTIgMCAwMC4xMTktLjEyN2MtLjE0OCAxLjM0NS0yLjAyOSAzLjI0NS0zLjk3NyA1LjA5MXptMy42NTctNi40NmwtLjAwMy0uMDAyYTEuODIyIDEuODIyIDAgMDEyLjQ1OS0yLjY4NGwtMS42MSAxLjYxM2EuMTE5LjExOSAwIDAwMCAuMTY5bDEuMjQ3IDEuMjQ3YTEuODE3IDEuODE3IDAgMDEtMi4wOTMtLjM0M3ptMi41NzggMGExLjcxNCAxLjcxNCAwIDAxLS4yNzEuMjE4aC0uMDAxbC0xLjIwNy0xLjIwNyAxLjUzMy0xLjUzM2MuNjYxLjcyLjYzNyAxLjgzMi0uMDU0IDIuNTIyem0tLjEtMS41NDRhLjE0My4xNDMgMCAwMC0uMDUzLjE1Ny40MTYuNDE2IDAgMDEtLjA1My40NS4xNC4xNCAwIDAwLjAyMy4xOTcuMTQxLjE0MSAwIDAwLjA4NC4wMy4xNC4xNCAwIDAwLjEwNi0uMDUuNjkxLjY5MSAwIDAwLjA4Ny0uNzUxLjEzOC4xMzggMCAwMC0uMTk0LS4wMzN6IiBmaWxsPSJjdXJyZW50Q29sb3IiLz48L3N2Zz4="
)

const (
	// postmanApiBudgetFraction is the share of the remaining request budget that all Postman
	// API calls of a single action may consume together. The remainder is what lets us return
	// a meaningful error: the agent bounds every call into the extension via the
	// Request-Timeout header, and once that budget is gone the handler is cut off and the
	// platform only ever sees "503 Service Unavailable ... Timeout".
	postmanApiBudgetFraction = 0.8
	// postmanApiFallbackBudget applies when the caller has no deadline at all, e.g. the
	// periodic discovery refresh.
	postmanApiFallbackBudget = 20 * time.Second
	// postmanApiDeadlineMargin is kept free at the end of the budget so an in-flight attempt
	// cannot run right up to the deadline.
	postmanApiDeadlineMargin = 500 * time.Millisecond
	// postmanApiRetryBackoff is multiplied by the number of attempts already made.
	postmanApiRetryBackoff = 500 * time.Millisecond
)

// postmanHttpClient is the shared client for all Postman API calls. It deliberately carries no
// Timeout of its own: every request is issued with a context whose deadline is derived from the
// caller's budget, which both bounds the attempt and cancels the in-flight request.
var postmanHttpClient = &http.Client{}

// PostmanApiContext bounds all Postman API calls of a single action to one shared budget derived
// from the caller's deadline. Without it a slow Postman API is charged per call, and the three
// sequential calls a prepare may make (list environments, download environment, download
// collection) can overrun the request budget on their own.
func PostmanApiContext(ctx context.Context) (context.Context, context.CancelFunc) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return context.WithTimeout(ctx, postmanApiFallbackBudget)
	}
	budget := time.Duration(float64(time.Until(deadline)) * postmanApiBudgetFraction)
	if budget <= 0 {
		// Already out of time; let the calls fail fast rather than start work we cannot finish.
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, budget)
}

// newPostmanApiRequest builds an authenticated GET request against the Postman API. The API
// key is sent via the X-API-Key header (never as a query parameter), so it is not exposed on
// a child process command line or persisted in the serialized action state.
func newPostmanApiRequest(ctx context.Context, pathSegments ...string) (*http.Request, error) {
	specification := config.Config

	resourceUrl, err := url.Parse(specification.PostmanBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postman base url: %w", err)
	}
	resourceUrl.Path += "/" + strings.Join(pathSegments, "/")

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, resourceUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-API-Key", specification.PostmanApiKey)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", fmt.Sprintf("steadybit-extension-postman/%s", extbuild.GetSemverVersionStringOrUnknown()))
	return req, nil
}

// getPostmanApiResource GETs a resource from the Postman API and returns the raw response body.
// Transient failures (network errors, 429 and 5xx) are retried with a linear backoff. Every
// attempt is bounded by the smaller of config.PostmanApiTimeout and the time left on ctx, and no
// further attempt is started once the budget can no longer accommodate one.
func getPostmanApiResource(ctx context.Context, pathSegments ...string) ([]byte, error) {
	resource := strings.Join(pathSegments, "/")
	attempts := config.Config.PostmanApiMaxAttempts
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			if !sleepWithContext(ctx, time.Duration(attempt-1)*postmanApiRetryBackoff) {
				return nil, fmt.Errorf("postman api request for %s aborted while backing off: %w", resource, context.Cause(ctx))
			}
		}

		timeout := attemptTimeout(ctx)
		if timeout <= 0 {
			if lastErr != nil {
				return nil, fmt.Errorf("postman api budget for %s exhausted after %d attempt(s), last error: %w", resource, attempt-1, lastErr)
			}
			return nil, fmt.Errorf("no time left in the request budget to call the postman api for %s", resource)
		}

		body, retryable, err := getPostmanApiResourceOnce(ctx, timeout, resource, pathSegments)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retryable {
			return nil, err
		}
		log.Warn().Err(err).Msgf("Transient failure calling the postman api for %s (attempt %d/%d)", resource, attempt, attempts)
	}
	return nil, fmt.Errorf("postman api call for %s failed after %d attempt(s): %w", resource, attempts, lastErr)
}

// getPostmanApiResourceOnce performs a single attempt. The bool reports whether the failure is
// worth retrying.
func getPostmanApiResourceOnce(ctx context.Context, timeout time.Duration, resource string, pathSegments []string) ([]byte, bool, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := newPostmanApiRequest(attemptCtx, pathSegments...)
	if err != nil {
		return nil, false, err
	}

	response, err := postmanHttpClient.Do(req)
	if err != nil {
		// The shared budget being gone is final; a single attempt timing out is worth retrying.
		if ctx.Err() != nil {
			return nil, false, fmt.Errorf("postman api request for %s aborted: %w", resource, context.Cause(ctx))
		}
		return nil, true, fmt.Errorf("failed to request %s from postman api: %w", resource, err)
	}
	defer func() {
		if cerr := response.Body.Close(); cerr != nil {
			log.Error().Msgf("Failed to close response body. Got error: %s", cerr)
		}
	}()

	if isRetryableStatus(response.StatusCode) {
		return nil, true, fmt.Errorf("postman api returned status %s for %s", response.Status, resource)
	}
	if response.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("postman api returned status %s for %s", response.Status, resource)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		if ctx.Err() != nil {
			return nil, false, fmt.Errorf("reading the postman api response for %s was aborted: %w", resource, context.Cause(ctx))
		}
		return nil, true, fmt.Errorf("failed to read %s response body: %w", resource, err)
	}
	return body, false, nil
}

// isRetryableStatus reports whether a status code indicates a condition that may pass on its
// own: rate limiting and server-side errors.
func isRetryableStatus(statusCode int) bool {
	return statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError
}

// attemptTimeout returns how long a single attempt may take, or a non-positive duration when the
// budget is spent.
func attemptTimeout(ctx context.Context) time.Duration {
	timeout := config.Config.PostmanApiTimeout
	if timeout <= 0 {
		timeout = postmanApiFallbackBudget
	}

	deadline, ok := ctx.Deadline()
	if !ok {
		return timeout
	}
	usable := time.Until(deadline) - postmanApiDeadlineMargin
	if usable <= 0 {
		return 0
	}
	return min(timeout, usable)
}

// sleepWithContext waits for d, reporting false if ctx was done first.
func sleepWithContext(ctx context.Context, d time.Duration) bool {
	if d <= 0 {
		return ctx.Err() == nil
	}
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

// downloadPostmanResource fetches a resource from the Postman API and writes it to destPath.
// The API wraps the resource in a single top-level key (e.g. {"collection": {...}}); when
// present, that inner object is unwrapped so newman receives the canonical file format.
func downloadPostmanResource(ctx context.Context, resourcePath, id, wrapperKey, destPath string) error {
	body, err := getPostmanApiResource(ctx, resourcePath, id)
	if err != nil {
		return err
	}

	content := body
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapper); err == nil {
		if inner, ok := wrapper[wrapperKey]; ok {
			content = inner
		} else {
			log.Warn().Msgf("Postman API response for %s/%s did not contain expected wrapper key %q; using raw body", resourcePath, id, wrapperKey)
		}
	}

	if err := os.WriteFile(destPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write %s to disk: %w", resourcePath, err)
	}
	return nil
}
