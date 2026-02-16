package opensearch

import (
	"context"
	"fmt"
	"strings"

	opensearchapi "github.com/opensearch-project/opensearch-go/v4/opensearchapi"
)

const indexTemplateName = "mintlog-logs"

const indexTemplateBody = `{
  "index_patterns": ["mintlog-*"],
  "template": {
    "settings": {
      "number_of_shards": 1,
      "number_of_replicas": 0,
      "refresh_interval": "5s"
    },
    "mappings": {
      "properties": {
        "id":        { "type": "keyword" },
        "tenant_id": { "type": "keyword" },
        "timestamp": { "type": "date" },
        "level":     { "type": "keyword" },
        "message":   { "type": "text", "analyzer": "standard" },
        "service":   { "type": "keyword" },
        "host":      { "type": "keyword" },
        "trace_id":  { "type": "keyword" },
        "span_id":   { "type": "keyword" },
        "tags":      { "type": "keyword" },
        "fields":    { "type": "object", "enabled": true }
      }
    }
  },
  "priority": 100
}`

func EnsureIndexTemplate(ctx context.Context, client *opensearchapi.Client) error {
	_, err := client.IndexTemplate.Create(ctx, opensearchapi.IndexTemplateCreateReq{
		IndexTemplate: indexTemplateName,
		Body:          strings.NewReader(indexTemplateBody),
	})
	if err != nil {
		return fmt.Errorf("create index template: %w", err)
	}
	return nil
}
