// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
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
	postmanApiRetryBackoff    = 500 * time.Millisecond
	postmanApiErrorBodyLimit  = 2048
	postmanApiLowMonthlyQuota = 0.1
	postmanApiQuotaWarnEvery  = time.Hour
)

var lastQuotaWarning atomic.Int64

// retryableError marks a failure worth retrying; wait is the delay Postman asked for, if any.
type retryableError struct {
	err  error
	wait time.Duration
}

func (e *retryableError) Error() string { return e.err.Error() }
func (e *retryableError) Unwrap() error { return e.err }

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
// Transient failures (network errors, 429 and 5xx) are retried with a linear backoff, or after the
// delay Postman asks for. Every attempt is bounded by the smaller of
// config.PostmanApiTimeout and the time left on ctx, and no further attempt is started once the
// budget can no longer accommodate one.
func getPostmanApiResource(ctx context.Context, pathSegments ...string) ([]byte, error) {
	resource := strings.Join(pathSegments, "/")
	attempts := config.Config.PostmanApiMaxAttempts
	if attempts < 1 {
		attempts = 1
	}

	var lastErr error
	var wait time.Duration
	for attempt := 1; attempt <= attempts; attempt++ {
		if attempt > 1 {
			if !sleepWithContext(ctx, max(wait, time.Duration(attempt-1)*postmanApiRetryBackoff)) {
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

		body, err := getPostmanApiResourceOnce(ctx, timeout, resource, pathSegments)
		if err == nil {
			return body, nil
		}
		lastErr = err
		var retryable *retryableError
		if !errors.As(err, &retryable) {
			return nil, err
		}
		wait = retryable.wait
		if attempt < attempts && wait > 0 && !fitsBudget(ctx, wait) {
			return nil, fmt.Errorf("rate limited by the postman api for %s, retrying is possible in %s, after the request budget: %w", resource, wait, err)
		}
		log.Warn().Err(err).Msgf("Transient failure calling the postman api for %s (attempt %d/%d)", resource, attempt, attempts)
	}
	return nil, fmt.Errorf("postman api call for %s failed after %d attempt(s): %w", resource, attempts, lastErr)
}

// getPostmanApiResourceOnce performs one attempt; retryable failures are *retryableError.
func getPostmanApiResourceOnce(ctx context.Context, timeout time.Duration, resource string, pathSegments []string) ([]byte, error) {
	attemptCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := newPostmanApiRequest(attemptCtx, pathSegments...)
	if err != nil {
		return nil, err
	}

	response, err := postmanHttpClient.Do(req)
	if err != nil {
		// The shared budget being gone is final; a single attempt timing out is worth retrying.
		if ctx.Err() != nil {
			return nil, fmt.Errorf("postman api request for %s aborted: %w", resource, context.Cause(ctx))
		}
		return nil, &retryableError{err: fmt.Errorf("failed to request %s from postman api: %w", resource, err)}
	}
	defer func() {
		if cerr := response.Body.Close(); cerr != nil {
			log.Error().Msgf("Failed to close response body. Got error: %s", cerr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return nil, postmanApiStatusError(response, resource)
	}
	warnOnLowMonthlyQuota(response.Header)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("reading the postman api response for %s was aborted: %w", resource, context.Cause(ctx))
		}
		return nil, &retryableError{err: fmt.Errorf("failed to read %s response body: %w", resource, err)}
	}
	return body, nil
}

func postmanApiStatusError(response *http.Response, resource string) error {
	err := fmt.Errorf("postman api returned status %s for %s", response.Status, resource)
	body, _ := io.ReadAll(io.LimitReader(response.Body, postmanApiErrorBodyLimit))
	if detail := postmanErrorDetail(body); detail != "" {
		err = fmt.Errorf("%w: %s", err, detail)
	}

	_, remainingMonth := monthlyQuota(response.Header)
	// Postman documents no status code for an exhausted monthly limit, only this body.
	exhausted := strings.Contains(strings.ToLower(string(body)), "service limit exhausted") ||
		(response.StatusCode == http.StatusTooManyRequests && remainingMonth == 0)
	switch {
	case exhausted:
		return fmt.Errorf("the monthly postman api call limit for this api key is exhausted: %w", err)
	case response.StatusCode == http.StatusTooManyRequests:
		return &retryableError{err: err, wait: retryAfter(response.Header)}
	case response.StatusCode >= http.StatusInternalServerError:
		return &retryableError{err: err}
	default:
		return err
	}
}

// postmanErrorDetail returns "name: message" of Postman's {"error": {...}} body, else the raw body.
func postmanErrorDetail(body []byte) string {
	var parsed struct {
		Error struct {
			Name    string `json:"name"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &parsed) == nil && (parsed.Error.Name != "" || parsed.Error.Message != "") {
		return strings.TrimPrefix(parsed.Error.Name+": "+parsed.Error.Message, ": ")
	}
	return strings.Join(strings.Fields(string(body)), " ")
}

// retryAfter returns the delay Postman asks for: RetryAfter, or the reset of an exhausted
// RateLimit header ("limit=300, remaining=0, reset=12"); zero if none.
func retryAfter(header http.Header) time.Duration {
	seconds := headerInt(header, "RetryAfter", "X-RateLimit-RetryAfter", "Retry-After")
	if seconds <= 0 && rateLimitParam(header, "remaining") == 0 {
		seconds = rateLimitParam(header, "reset")
	}
	return time.Duration(max(seconds, 0)) * time.Second
}

// rateLimitParam returns the named parameter of the RateLimit header, -1 if absent. Postman sends
// the header twice, e.g. "limit=10, remaining=9, reset=1" and "\"monthly\";r=9448".
func rateLimitParam(header http.Header, name string) int {
	for _, value := range header.Values("RateLimit") {
		for part := range strings.SplitSeq(value, ",") {
			if param, ok := strings.CutPrefix(strings.TrimSpace(part), name+"="); ok {
				if n, err := strconv.Atoi(param); err == nil {
					return n
				}
			}
		}
	}
	return -1
}

// monthlyQuota returns Postman's monthly call limit and remaining calls, -1 if absent.
func monthlyQuota(header http.Header) (limit, remaining int) {
	return headerInt(header, "RateLimit-Limit-Month", "X-RateLimit-Limit-Month"),
		headerInt(header, "RateLimit-Remaining-Month", "X-RateLimit-Remaining-Month")
}

func warnOnLowMonthlyQuota(header http.Header) {
	limit, remaining := monthlyQuota(header)
	if limit <= 0 || remaining < 0 || float64(remaining) >= float64(limit)*postmanApiLowMonthlyQuota {
		return
	}
	now := time.Now().UnixNano()
	if last := lastQuotaWarning.Load(); now-last >= int64(postmanApiQuotaWarnEvery) && lastQuotaWarning.CompareAndSwap(last, now) {
		log.Warn().Int("remaining", remaining).Int("limit", limit).Msg("The monthly postman api call limit for this api key is almost exhausted")
	}
}

// headerInt returns the first integer among the named headers, -1 if there is none.
func headerInt(header http.Header, names ...string) int {
	for _, name := range names {
		if value, err := strconv.Atoi(strings.TrimSpace(header.Get(name))); err == nil {
			return value
		}
	}
	return -1
}

// fitsBudget reports whether ctx leaves time to wait d and then make a full attempt.
func fitsBudget(ctx context.Context, d time.Duration) bool {
	remaining, ok := remainingBudget(ctx)
	return !ok || remaining >= d+postmanApiAttemptTimeout()
}

// remainingBudget returns the usable time left on ctx, false if ctx has no deadline.
func remainingBudget(ctx context.Context) (time.Duration, bool) {
	deadline, ok := ctx.Deadline()
	if !ok {
		return 0, false
	}
	return time.Until(deadline) - postmanApiDeadlineMargin, true
}

func postmanApiAttemptTimeout() time.Duration {
	if config.Config.PostmanApiTimeout <= 0 {
		return postmanApiFallbackBudget
	}
	return config.Config.PostmanApiTimeout
}

// attemptTimeout returns how long a single attempt may take, or a non-positive duration when the
// budget is spent.
func attemptTimeout(ctx context.Context) time.Duration {
	timeout := postmanApiAttemptTimeout()
	usable, ok := remainingBudget(ctx)
	if !ok {
		return timeout
	}
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
