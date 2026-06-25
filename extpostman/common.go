// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2022 Steadybit GmbH

package extpostman

import (
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

// postmanHttpClient is the shared client for all Postman API calls. The timeout bounds the
// request so a slow or unresponsive Postman API cannot hang the action/discovery indefinitely.
var postmanHttpClient = &http.Client{Timeout: 30 * time.Second}

// newPostmanApiRequest builds an authenticated GET request against the Postman API. The API
// key is sent via the X-API-Key header (never as a query parameter), so it is not exposed on
// a child process command line or persisted in the serialized action state.
func newPostmanApiRequest(pathSegments ...string) (*http.Request, error) {
	specification := config.Config

	resourceUrl, err := url.Parse(specification.PostmanBaseUrl)
	if err != nil {
		return nil, fmt.Errorf("failed to parse postman base url: %w", err)
	}
	resourceUrl.Path += "/" + strings.Join(pathSegments, "/")

	req, err := http.NewRequest(http.MethodGet, resourceUrl.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Add("X-API-Key", specification.PostmanApiKey)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", fmt.Sprintf("steadybit-extension-postman/%s", extbuild.GetSemverVersionStringOrUnknown()))
	return req, nil
}

// downloadPostmanResource fetches a resource from the Postman API and writes it to destPath.
// The API wraps the resource in a single top-level key (e.g. {"collection": {...}}); when
// present, that inner object is unwrapped so newman receives the canonical file format.
func downloadPostmanResource(resourcePath, id, wrapperKey, destPath string) error {
	req, err := newPostmanApiRequest(resourcePath, id)
	if err != nil {
		return err
	}

	response, err := postmanHttpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to request %s from postman api: %w", resourcePath, err)
	}
	defer func() {
		if cerr := response.Body.Close(); cerr != nil {
			log.Error().Msgf("Failed to close response body. Got error: %s", cerr)
		}
	}()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to download %s from postman api, got status code %s", resourcePath, response.Status)
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return fmt.Errorf("failed to read %s response body: %w", resourcePath, err)
	}

	content := body
	var wrapper map[string]json.RawMessage
	if err := json.Unmarshal(body, &wrapper); err == nil {
		if inner, ok := wrapper[wrapperKey]; ok {
			content = inner
		}
	}

	if err := os.WriteFile(destPath, content, 0600); err != nil {
		return fmt.Errorf("failed to write %s to disk: %w", resourcePath, err)
	}
	return nil
}
