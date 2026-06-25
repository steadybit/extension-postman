package extpostman

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/rs/zerolog/log"
)

type PostmanCollectionResult struct {
	Collections []PostmanCollection `json:"collections"`
}
type PostmanCollection struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// DownloadCollection fetches the collection from the Postman API and writes it to destPath.
func DownloadCollection(collectionId, destPath string) error {
	return downloadPostmanResource("collections", collectionId, "collection", destPath)
}

func GetPostmanCollections() []PostmanCollection {
	req, err := newPostmanApiRequest("collections")
	if err != nil {
		log.Error().Msgf("Failed to create request for postman api. Got error: %s", err)
		return nil
	}

	response, err := postmanHttpClient.Do(req)
	if err != nil {
		log.Error().Msgf("Failed to get collections from postman api. Got error: %s", err)
		return nil
	}
	if response.StatusCode != http.StatusOK {
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
