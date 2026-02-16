package opensearch

import (
	"crypto/tls"
	"fmt"
	"net/http"

	opensearch "github.com/opensearch-project/opensearch-go/v4"
	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

func NewClient(url, user, password string) (*opensearchapi.Client, error) {
	client, err := opensearchapi.NewClient(opensearchapi.Config{
		Client: opensearch.Config{
			Addresses: []string{url},
			Username:  user,
			Password:  password,
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("opensearch client: %w", err)
	}
	return client, nil
}
