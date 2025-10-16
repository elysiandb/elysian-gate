package replication

import (
	"encoding/json"
	"fmt"
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
	for _, t := range types {
		err := resetNodeEntity(node, t)
		if err != nil {
			return err
		}
		entities, err := listNodeEntities(master, t)
		if err != nil {
			return err
		}
		for _, entity := range entities {
			err := sendEntityToNode(entity, node, t)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func listNodeEntityTypes(node *global.Node) ([]string, error) {
	url := fmt.Sprintf("http://%s:%d/kv/api:entity:types:list", node.HTTP.Host, node.HTTP.Port)
	logger.Info("Listing entity types from " + url)

	status, body, err := forward.ForwardRequest("GET", url, "")
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

	parts := strings.Split(data.Value, ",")
	for i := range parts {
		parts[i] = strings.TrimSpace(parts[i])
	}

	return parts, nil
}

func listNodeEntities(node *global.Node, entity string) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("http://%s:%d/api/%s", node.HTTP.Host, node.HTTP.Port, entity)
	status, body, err := forward.ForwardRequest("GET", url, "")
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
	url := fmt.Sprintf("http://%s:%d/api/%s", node.HTTP.Host, node.HTTP.Port, entityType)
	_, _, err = forward.ForwardRequest("POST", url, string(payload))
	return err
}

func resetNodeEntity(node *global.Node, entity string) error {
	url := fmt.Sprintf("http://%s:%d/%s", node.HTTP.Host, node.HTTP.Port, entity)
	_, _, err := forward.ForwardRequest("DELETE", url, "")
	return err
}
