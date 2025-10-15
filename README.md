# ElysianGate — Smart Gateway and Load Balancer for ElysianDB Clusters

### Overview

**ElysianGate** is a lightweight, high-performance gateway designed to orchestrate and balance multiple **ElysianDB** nodes. It manages request routing, replication, and real-time monitoring, enabling distributed key-value clusters to behave as one unified system.

---

### Key Features

* **Automatic Cluster Bootstrapping** — Optionally launch all ElysianDB nodes at startup.
* **Intelligent Read/Write Routing** — Routes writes to the master and distributes reads across fresh slaves.
* **Replication Engine** — Synchronizes recent write operations across slave nodes for consistency.
* **Real-Time Health Monitoring** — Continuously checks node status via HTTP and TCP.
* **Dual Transport Awareness** — Supports both **HTTP** and **TCP** endpoints per node.
* **Zero Configuration** — Simple YAML-based setup for instant startup.
* **k6 Benchmark Suite** — Included for stress and performance testing.

---

### Configuration Example (`elysiangate.yaml`)

```yaml
nodes:
  node1:
    role: master
    http: { enabled: true, host: 0.0.0.0, port: 8090 }
    tcp:  { enabled: true, host: 0.0.0.0, port: 8890 }
  node2:
    role: slave
    http: { enabled: true, host: 0.0.0.0, port: 8091 }
    tcp:  { enabled: true, host: 0.0.0.0, port: 8891 }
  node3:
    role: slave
    http: { enabled: true, host: 0.0.0.0, port: 8092 }
    tcp:  { enabled: true, host: 0.0.0.0, port: 8892 }
  node4:
    role: slave
    http: { enabled: true, host: 0.0.0.0, port: 8093 }
    tcp:  { enabled: true, host: 0.0.0.0, port: 8893 }

gateway:
  startsNodes: false
  http:
    host: "0.0.0.0"
    port: 8899
```

---

### Usage

#### Start the Gateway

```bash
go run . --config elysiangate.yaml
```

#### Start Fresh (clear previous data)

```bash
go run . --config elysiangate.yaml --clear
```

#### Launch the Cluster Manually

```bash
make cluster
```

---

### Architecture

* **Configuration Loader** — Parses YAML and loads gateway and node definitions.
* **Cluster Manager** — Maintains in-memory registry of node states and roles.
* **Replication Balancer** — Tracks pending write operations and synchronizes them with slaves.
* **Monitoring Engine** — Performs continuous health checks and logs node status changes.
* **HTTP Gateway Server** — Built with **fasthttp** for ultra-low latency routing.

---

### Monitoring Output Example

```
15:42:03
Node master (node1) [HTTP 0.0.0.0:8090 | TCP 0.0.0.0:8890] : 🟢 HTTP up | 🟢 TCP up
Node slave (node2) [HTTP 0.0.0.0:8091 | TCP 0.0.0.0:8891] : 🔴 HTTP down | 🟢 TCP up
───────────────────────────────────────────────
```

---

### Philosophy

> *ElysianGate turns distributed ElysianDB clusters into a single, coherent system — effortless setup, instant visibility, and consistent performance by design.*

