# ElysianGate — The Smart Gateway for ElysianDB Clusters

<p align="center">
  <img src="docs/logo.png" alt="ElysianGate Logo" width="180"/>
</p>

---

### Overview

**ElysianGate** is a lightweight, high-performance gateway designed to orchestrate multiple **ElysianDB** nodes. It provides smart request routing, node health monitoring, and seamless startup or discovery of distributed key-value stores.

---

### ✨ Key Features

* **Automatic Cluster Bootstrapping** — Optionally launch all ElysianDB nodes at startup.
* **Node Discovery Mode** — Detect and monitor already running nodes in real time.
* **Dual Transport Awareness** — Handles both **HTTP** and **TCP** connections per node.
* **Real-Time Health Checks** — Continuously monitors all nodes and logs status changes instantly.
* **Zero Configuration** — Simple YAML-based setup.

---

### Configuration Example (`elysiangate.yaml`)

```yaml
nodes:
  - "./elysiandb/config/elysian-1.yaml"
  - "./elysiandb/config/elysian-2.yaml"
  - "./elysiandb/config/elysian-3.yaml"
  - "./elysiandb/config/elysian-4.yaml"

gateway:
  startsNodes: false
  http:
    host: "0.0.0.0"
    port: 8899
```

---

### 🚀 Usage

#### Start the Gateway

```bash
go run . --config elysiangate.yaml
```

#### Optional: Start Fresh

```bash
go run . --config elysiangate.yaml --clear
```

#### Start the ElysianDB Cluster Manually

```bash
make cluster
```

---

### 🧠 Architecture

* **Configuration Layer** — Parses YAML and loads gateway & node settings.
* **Cluster Manager** — Keeps an in-memory registry of all nodes with HTTP/TCP health.
* **Monitoring Engine** — Pings each node periodically, logging state transitions.
* **Gateway Server** — A minimal **fasthttp** service handling API and orchestration requests.

---

### Monitoring Output Example

```
15:42:03
Node 1 [HTTP 0.0.0.0:8090 | TCP 0.0.0.0:8071] : 🟢 HTTP up | 🟢 TCP up
Node 2 [HTTP 0.0.0.0:8091 | TCP 0.0.0.0:8072] : 🔴 HTTP down | 🟢 TCP up
───────────────────────────────────────────────
```

---

### Philosophy

> *ElysianGate aims to make distributed ElysianDB clusters effortless — minimal setup, instant observability, and full control over performance and reliability.*

---

### 🧑‍💻 Author

**Taymour**
Creator of [ElysianDB](https://github.com/elysiandb/elysiandb)

---

### 📜 License

MIT License © 2025
