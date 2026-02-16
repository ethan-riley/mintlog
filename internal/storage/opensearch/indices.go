package opensearch

import (
	"context"
	"fmt"
	"time"

	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

// IndexName returns the per-tenant, date-partitioned index name.
func IndexName(tenantID string, ts time.Time) string {
	return fmt.Sprintf("mintlog-%s-%s", tenantID, ts.Format("2006.01.02"))
}

// EnsureIndex creates the index if it doesn't exist (the template handles mappings).
func EnsureIndex(ctx context.Context, client *opensearchapi.Client, name string) error {
	resp, err := client.Indices.Exists(ctx, opensearchapi.IndicesExistsReq{
		Indices: []string{name},
	})
	if err == nil && resp.StatusCode == 200 {
		return nil
	}

	_, err = client.Indices.Create(ctx, opensearchapi.IndicesCreateReq{
		Index: name,
	})
	if err != nil {
		return fmt.Errorf("create index %s: %w", name, err)
	}
	return nil
}
