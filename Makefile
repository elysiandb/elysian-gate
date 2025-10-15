.PHONY: clear cluster stop restart

BIN=./elysiandb/bin/elysiandb
CONF_DIR=./elysiandb/config

clear:
	rm -rf /tmp/elysian*

cluster:
	@echo "Starting ElysianDB cluster..."
	@$(BIN) --config $(CONF_DIR)/elysian-1.yaml &
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-2.yaml &
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-3.yaml &
	@sleep 0.5
	@$(BIN) --config $(CONF_DIR)/elysian-4.yaml &
	@sleep 2
	@echo "✅ Cluster started."

stop:
	@echo "Stopping ElysianDB cluster..."
	-@pkill -f "$(BIN)" >/dev/null 2>&1 || true
	@sleep 2
	@echo "🛑 Cluster stopped."

restart:
	@echo "🔄 Restarting cluster..."
	@$(MAKE) stop || true
	@$(MAKE) clear
	@sleep 2
	@$(MAKE) cluster
	@echo "♻️  Restart complete."
