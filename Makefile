.PHONY: clear cluster stop restart api_benchmark test test-cover

BIN=./elysiandb/bin/elysiandb
CONF_DIR=./elysiandb/config

COVERPKG := $(shell go list ./internal/... | paste -sd, -)

clear:
	rm -rf /tmp/elysian*

cluster:
	@echo "Starting ElysianDB cluster..."
	@$(BIN) --config $(CONF_DIR)/elysian-1.yaml & # slave 1
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-2.yaml & # slave 2
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-3.yaml & # slave 3
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-4.yaml & # slave 4
	@sleep 2
	@echo "✅ Cluster started."

group-1:
	@echo "Starting ElysianDB group 1..."
	@$(BIN) --config $(CONF_DIR)/elysian-1.yaml & # slave 1
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-2.yaml & # slave 2
	@sleep 2
	@echo "✅ Group 1 started."

group-2:
	@echo "Starting ElysianDB group 2..."
	@$(BIN) --config $(CONF_DIR)/elysian-3.yaml & # slave 3
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-4.yaml & # slave 4
	@sleep 2
	@echo "✅ Group 1 started."

cluster-slaves:
	@echo "Starting ElysianDB cluster slaves..."
	@$(BIN) --config $(CONF_DIR)/elysian-2.yaml & # slave 1
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-3.yaml & # slave 2
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-4.yaml & # slave 3
	@sleep 2
	@echo "✅ Slaves started."

stop:
	@echo "Stopping full ElysianDB cluster..."
	-@pkill -f "$(BIN)" >/dev/null 2>&1 || true
	@sleep 2
	@echo "🛑 Cluster stopped."

stop-slaves:
	@echo "Stopping ElysianDB slave nodes..."
	-@pkill -f "$(CONF_DIR)/elysian-2.yaml" >/dev/null 2>&1 || true
	-@pkill -f "$(CONF_DIR)/elysian-3.yaml" >/dev/null 2>&1 || true
	-@pkill -f "$(CONF_DIR)/elysian-4.yaml" >/dev/null 2>&1 || true
	@sleep 2
	@echo "🛑 Slaves stopped."

restart:
	@echo "🔄 Restarting cluster..."
	@$(MAKE) stop || true
	@$(MAKE) clear
	@sleep 2
	@$(MAKE) cluster
	@echo "♻️  Restart complete."

restart-slaves:
	@echo "🔄 Restarting cluster slaves..."
	@$(MAKE) stop-slaves || true
	@$(MAKE) clear
	@sleep 2
	@$(MAKE) cluster-slaves
	@echo "♻️  Slaves restart complete."

api_benchmark:
	BASE_URL=http://localhost:8899 KEYS=5000 VUS=200 DURATION=30s k6 run elysian_api_k6.js

test:
	@go test ./tests/... -v

test-cover:
	@go test -coverpkg=$(COVERPKG) ./tests/... -coverprofile=coverage.out -count=1
	@go tool cover -func=coverage.out | tail -n1