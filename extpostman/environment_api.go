package extpostman

import (
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-postman/config"
	"io"
	"net/http"
	"net/url"
)

type PostmanEnvironmentResult struct {
	Environments []PostmanEnvironment `json:"Environments"`
}
type PostmanEnvironment struct {
	Id string `json:"id"`
	Name string `json:"name"`
}

func GetPostEnvironmentId(environmentIdOrName string) (string, error) {
	//check if uuid
	environmentId, err := uuid.Parse(environmentIdOrName)
	if err == nil {
		return environmentId.String(), nil
	}

	environments := GetPostmanEnvironments()
	var uniqueEnvironmentId string
	counter := 0
	for _, environment := range environments {
		if environment.Name == environmentIdOrName {
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
	var specification = config.Config
	var apiKey = specification.PostmanApiKey
	EnvironmentsUrl, err := url.Parse(specification.PostmanBaseUrl)
	if err != nil {
		log.Error().Msgf("Failed to parse postman base url. Got error: %s", err)
		return nil
	}
	EnvironmentsUrl.Path += "/environments"
	parameters := url.Values{}
	parameters.Add("apikey", apiKey)
	EnvironmentsUrl.RawQuery = parameters.Encode()


	response, err := http.Get(EnvironmentsUrl.String())
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
