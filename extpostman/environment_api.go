package extpostman

import (
	"context"
	"encoding/json"
	"fmt"

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
func DownloadEnvironment(ctx context.Context, environmentId, destPath string) error {
	return downloadPostmanResource(ctx, "environments", environmentId, "environment", destPath)
}

func GetPostEnvironmentId(ctx context.Context, environmentIdOrName string) (string, error) {
	log.Info().Msgf("Searching for environment with id or name '%s'", environmentIdOrName)
	environmentId, err := uuid.Parse(environmentIdOrName)
	if err == nil {
		log.Info().Msgf("Found environment id '%s'", environmentId.String())
		return environmentId.String(), nil
	}

	// Resolving a name needs the full list. Failing to fetch it is reported as such: it used to
	// be swallowed and surfaced as "failed to find environment", which sends anyone debugging a
	// transient Postman API problem looking for a misconfigured environment name instead.
	environments, err := GetPostmanEnvironments(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to list environments while resolving name '%s': %w", environmentIdOrName, err)
	}
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

func GetPostmanEnvironments(ctx context.Context) ([]PostmanEnvironment, error) {
	body, err := getPostmanApiResource(ctx, "environments")
	if err != nil {
		return nil, err
	}

	var result PostmanEnvironmentResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode environments response: %w", err)
	}
	return result.Environments, nil
}
