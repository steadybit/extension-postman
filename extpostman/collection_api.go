package extpostman

import (
	"encoding/json"
	"github.com/rs/zerolog/log"
	"github.com/steadybit/extension-postman/v2/config"
	"io"
	"net/http"
	"net/url"
)

type PostmanCollectionResult struct {
	Collections []PostmanCollection `json:"collections"`
}
type PostmanCollection struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func GetPostmanCollections() []PostmanCollection {
	var specification = config.Config
	var apiKey = specification.PostmanApiKey
	collectionsUrl, err := url.Parse(specification.PostmanBaseUrl)
	if err != nil {
		log.Error().Msgf("Failed to parse postman base url. Got error: %s", err)
		return nil
	}
	collectionsUrl.Path += "/collections"
	parameters := url.Values{}
	parameters.Add("apikey", apiKey)
	collectionsUrl.RawQuery = parameters.Encode()

	response, err := http.Get(collectionsUrl.String())
	if err != nil {
		log.Error().Msgf("Failed to get collections from postman api. Got error: %s", err)
		return nil
	}
	if response.StatusCode != 200 {
		log.Error().Msgf("Failed to get collections from postman api. Got status code: %s", response.Status)
		return nil
	}

	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			log.Error().Msgf("Failed to close response body. Got error: %s", err)
			return
		}
	}(response.Body)

	var result PostmanCollectionResult
	err = json.NewDecoder(response.Body).Decode(&result)
	if err != nil {
		log.Error().Msgf("Failed to decode response body. Got error: %s", err)
		return nil
	}

	return result.Collections
}
