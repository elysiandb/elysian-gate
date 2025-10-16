# ElysianGate — Smart Gateway and Load Balancer for ElysianDB Clusters

### Overview

ElysianGate is a lightweight, high-performance gateway designed to orchestrate and balance multiple ElysianDB nodes. It manages request routing, replication, and real-time monitoring, enabling distributed key-value clusters to behave as one unified system. Currently, it supports ElysianDB's automatic REST API mode, with native key-value mode replication planned for future releases.

---

### Key Features

* Automatic Cluster Bootstrapping — Optionally launch all ElysianDB nodes at startup.
* Intelligent Read/Write Routing — Routes writes to the master and distributes reads across synchronized slaves.
* Replication Engine — Automatically synchronizes master data to slave nodes at boot and when new nodes join.
* Real-Time Health Monitoring — Continuously checks node state through both HTTP and TCP.
* Dual Transport Support — Each node can expose both HTTP and TCP interfaces.
* YAML-Based Configuration — Simple, declarative setup for quick cluster orchestration.
* Built-in Benchmarking — k6 test scripts available for stress and performance evaluation.

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
  synchronizationInterval: 1
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

* Configuration Loader — Parses YAML and loads gateway and node definitions.
* Cluster Manager — Maintains the registry of nodes and tracks their status.
* Replication Manager — Ensures all slave nodes are consistent with the master.
* Health Monitor — Periodically verifies node liveness and readiness.
* HTTP Gateway Server — Uses fasthttp for low-latency routing and concurrency.

---

### Monitoring Output Example

```
15:42:03
Node node1 (master) [HTTP 0.0.0.0:8090 | TCP 0.0.0.0:8890] : HTTP up | TCP up | Ready
Node node2 (slave) [HTTP 0.0.0.0:8091 | TCP 0.0.0.0:8891] : HTTP up | TCP up | Ready
Node node3 (slave) [HTTP 0.0.0.0:8092 | TCP 0.0.0.0:8892] : HTTP up | TCP up | Ready
```

---

### Philosophy

ElysianGate transforms distributed ElysianDB clusters into a single, coherent system. It focuses on simplicity, visibility, and reliability, ensuring every node stays synchronized and every query is consistently served.
