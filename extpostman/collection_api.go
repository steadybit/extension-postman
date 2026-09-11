package extpostman

import (
	"context"
	"encoding/json"
	"fmt"
)

type PostmanCollectionResult struct {
	Collections []PostmanCollection `json:"collections"`
}
type PostmanCollection struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

// DownloadCollection fetches the collection from the Postman API and writes it to destPath.
func DownloadCollection(ctx context.Context, collectionId, destPath string) error {
	return downloadPostmanResource(ctx, "collections", collectionId, "collection", destPath)
}

func GetPostmanCollections(ctx context.Context) ([]PostmanCollection, error) {
	body, err := getPostmanApiResource(ctx, "collections")
	if err != nil {
		return nil, err
	}

	var result PostmanCollectionResult
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to decode collections response: %w", err)
	}
	return result.Collections, nil
}
