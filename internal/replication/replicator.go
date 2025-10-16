package replication

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"

	"github.com/elysiandb/elysian-gate/internal/forward"
	"github.com/elysiandb/elysian-gate/internal/global"
	"github.com/elysiandb/elysian-gate/internal/logger"
)

func ReplicateMasterToNode(master *global.Node, node *global.Node) error {
	types, err := listNodeEntityTypes(master)
	logger.Info(fmt.Sprintf("Replicating the following entity types from master to node: %v\n", types))
	if err != nil {
		return err
	}
	for _, raw := range types {
		t := sanitizeType(raw)
		if t == "" {
			continue
		}
		if err := resetNodeEntity(node, t); err != nil {
			return err
		}
		entities, err := listNodeEntities(master, t)
		if err != nil {
			return err
		}
		for _, entity := range entities {
			if err := sendEntityToNode(entity, node, t); err != nil {
				return err
			}
		}
	}
	return nil
}

func listNodeEntityTypes(node *global.Node) ([]string, error) {
	urlStr := fmt.Sprintf("http://%s:%d/kv/api:entity:types:list", node.HTTP.Host, node.HTTP.Port)
	logger.Info("Listing entity types from " + urlStr)

	status, body, err := forward.ForwardRequest("GET", urlStr, "")
	if err != nil || status >= 300 {
		return nil, err
	}

	var data struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return nil, err
	}
	if data.Value == "" {
		return []string{}, nil
	}

	raw := strings.Split(data.Value, ",")
	out := make([]string, 0, len(raw))
	for _, v := range raw {
		t := sanitizeType(v)
		if t != "" {
			out = append(out, t)
		}
	}
	return out, nil
}

func listNodeEntities(node *global.Node, entity string) ([]map[string]interface{}, error) {
	e := url.PathEscape(sanitizeType(entity))
	urlStr := fmt.Sprintf("http://%s:%d/api/%s", node.HTTP.Host, node.HTTP.Port, e)
	status, body, err := forward.ForwardRequest("GET", urlStr, "")
	if err != nil || status >= 300 {
		return nil, err
	}

	var entities []map[string]interface{}
	if err := json.Unmarshal([]byte(body), &entities); err != nil {
		return nil, err
	}

	return entities, nil
}

func sendEntityToNode(entity map[string]interface{}, node *global.Node, entityType string) error {
	payload, err := json.Marshal(entity)
	if err != nil {
		return err
	}
	e := url.PathEscape(sanitizeType(entityType))
	urlStr := fmt.Sprintf("http://%s:%d/api/%s", node.HTTP.Host, node.HTTP.Port, e)
	_, _, err = forward.ForwardRequest("POST", urlStr, string(payload))
	return err
}

func resetNodeEntity(node *global.Node, entity string) error {
	e := url.PathEscape(sanitizeType(entity))
	urlStr := fmt.Sprintf("http://%s:%d/api/%s", node.HTTP.Host, node.HTTP.Port, e)
	_, _, err := forward.ForwardRequest("DELETE", urlStr, "")
	return err
}

func sanitizeType(s string) string {
	s = strings.ReplaceAll(s, "\r", "")
	s = strings.ReplaceAll(s, "\n", "")
	return strings.TrimSpace(s)
}
