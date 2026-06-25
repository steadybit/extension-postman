// SPDX-License-Identifier: MIT
// SPDX-FileCopyrightText: 2025 Steadybit GmbH

package extpostman

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-kit/extbuild"
	"github.com/steadybit/extension-postman/v2/config"
)

// DownloadCollection fetches the collection from the Postman API and writes it to destPath.
func DownloadCollection(collectionId, destPath string) error {
	return downloadPostmanResource("collections", collectionId, "collection", destPath)
}

// DownloadEnvironment fetches the environment from the Postman API and writes it to destPath.
func DownloadEnvironment(environmentId, destPath string) error {
	return downloadPostmanResource("environments", environmentId, "environment", destPath)
}

// downloadPostmanResource fetches a resource from the Postman API authenticating via the
// X-API-Key header (instead of an apikey query parameter) so the API key is never exposed
// on a child process command line or persisted in the serialized action state. The API
// wraps the resource in a single top-level key (e.g. {"collection": {...}}); when present,
// that inner object is unwrapped so newman receives the canonical file format.
func downloadPostmanResource(resourcePath, id, wrapperKey, destPath string) error {
	specification := config.Config

	resourceUrl, err := url.Parse(specification.PostmanBaseUrl)
	if err != nil {
		return fmt.Errorf("failed to parse postman base url: %w", err)
	}
	resourceUrl.Path += fmt.Sprintf("/%s/%s", resourcePath, id)

	req, err := http.NewRequest("GET", resourceUrl.String(), nil)
	if err != nil {
		return err
	}
	req.Header.Add("X-API-Key", specification.PostmanApiKey)
	req.Header.Add("Accept", "*/*")
	req.Header.Add("User-Agent", fmt.Sprintf("steadybit-extension-postman/%s", extbuild.GetSemverVersionStringOrUnknown()))

	client := &http.Client{}
	response, err := client.Do(req)
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
