package extpostman

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type PostmanEnvironmentResult struct {
	Environments []PostmanEnvironment `json:"environments"`
}
type PostmanEnvironment struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// DownloadEnvironment fetches the environment from the Postman API and writes it to destPath.
func DownloadEnvironment(environmentId, destPath string) error {
	return downloadPostmanResource("environments", environmentId, "environment", destPath)
}

func GetPostEnvironmentId(environmentIdOrName string) (string, error) {
	log.Info().Msgf("Searching for environment with id or name '%s'", environmentIdOrName)
	environmentId, err := uuid.Parse(environmentIdOrName)
	if err == nil {
		log.Info().Msgf("Found environment id '%s'", environmentId.String())
		return environmentId.String(), nil
	}

	environments := GetPostmanEnvironments()
	log.Info().Msgf("Found %d environments", len(environments))
	var uniqueEnvironmentId string
	counter := 0
	for _, environment := range environments {
		if environment.Name == environmentIdOrName {
			log.Info().Msgf("Found environment with name '%s' and id '%s'", environment.Name, environment.Id)
			counter++
			uniqueEnvironmentId = environment.Id
		}
	}
	if counter > 1 {
		return "", fmt.Errorf("found multiple environments with name '%s'", environmentIdOrName)
	}
	if uniqueEnvironmentId != "" {
		return uniqueEnvironmentId, nil
	}

	return "", fmt.Errorf("failed to find environment with name '%s'", environmentIdOrName)
}

func GetPostmanEnvironments() []PostmanEnvironment {
	req, err := newPostmanApiRequest("environments")
	if err != nil {
		log.Error().Msgf("Failed to create request for postman api. Got error: %s", err)
		return nil
	}

	response, err := postmanHttpClient.Do(req)
	if err != nil {
		log.Error().Msgf("Failed to get Environments from postman api. Got error: %s", err)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Msgf("Failed to close response body. Got error: %s", err)
			return
		}
	}(response.Body)

	var result PostmanEnvironmentResult
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		log.Error().Msgf("Failed to decode response body. Got error: %s", err)
		return nil
	}

	return result.Environments
}
