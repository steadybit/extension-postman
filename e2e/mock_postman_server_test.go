package e2e

import (
	"fmt"
	"github.com/rs/zerolog/log"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
)

const collectionId = "5f757f0d-de24-462c-867f-256bb696d2dd"

func createMockPostmanServer() *httptest.Server {
	listener, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		panic(fmt.Sprintf("httptest: failed to listen: %v", err))
	}
	server := httptest.Server{
		Listener: listener,
		Config: &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Info().Str("path", r.URL.Path).Str("method", r.Method).Str("query", r.URL.RawQuery).Msg("Request received")
			if strings.Contains(r.URL.Path, "collections/"+collectionId) {
				w.WriteHeader(http.StatusOK)
				w.Write(getCollection())
			} else if strings.Contains(r.URL.Path, "collections") {
				w.WriteHeader(http.StatusOK)
				w.Write(getCollections())
			} else if strings.Contains(r.URL.Path, "environments") {
				w.WriteHeader(http.StatusOK)
				w.Write(getEnvironments())
			} else {
				w.WriteHeader(http.StatusBadRequest)
			}
		})},
	}
	server.Start()
	log.Info().Str("url", server.URL).Msg("Started Mock-Server")
	return &server
}

func getCollection() []byte {
	log.Info().Msg("Return collection")
	return []byte(`{"collection":{"info":{"_postman_id":"1c89f353-9e9d-4daf-9442-bc64f4c1b29b","name":"Simple Get","schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json","updatedAt":"2023-09-14T11:22:30.000Z","uid":"21222108-1c89f353-9e9d-4daf-9442-bc64f4c1b29b"},"item":[{"name":"Is google online","event":[{"listen":"test","script":{"id":"13df8fb5-f68d-4822-93b4-a6a67c512420","exec":["pm.test(\"Body matches string\", function () {","    pm.expect(pm.response.text()).to.include(\"Google\");","});"],"type":"text/javascript"}}],"id":"fd6c6e49-dbf2-4f21-92ef-c0a2bf10b426","protocolProfileBehavior":{"disableBodyPruning":true},"request":{"method":"GET","header":[],"url":{"raw":"https://www.google.de","protocol":"https","host":["www","google","de"]}},"response":[],"uid":"21222108-fd6c6e49-dbf2-4f21-92ef-c0a2bf10b426"}]}}`)
}

func getCollections() []byte {
	log.Info().Msg("Return collection")
	return []byte(`{
    "collections": [
        {
            "id": "5f757f0d-de24-462c-867f-256bb696d2dd",
            "name": "shopping-demo",
            "owner": "1111111",
            "createdAt": "2022-06-01T13:20:10.000Z",
            "updatedAt": "2023-03-10T13:26:48.000Z",
            "uid": "1111111-5f757f0d-de24-462c-867f-256bb696d2dd",
            "isPublic": false
        }
    ]
}`)
}

func getEnvironments() []byte {
	log.Info().Msg("Return environments")
	return []byte(`{"environments":[{"id":"70cb2138-3443-4c33-a45c-73477a5fd903","name":"dev","createdAt":"2022-10-04T14:02:50.000Z","updatedAt":"2022-10-05T11:28:57.000Z","owner":"123456","uid":"21211108-11cb2538-1111-1111-a45c-11177a5fd903","isPublic":false}]}`)
}
