# EdgDB_project_P-A_Boucher

## How to Run

### Prerequisites
- Docker & Docker Compose installed.
- Go installed (if running outside Docker).
- Python 3.7+ (for using `port-opener.py`).

### Running the Service

#### 1. Using Docker Compose
```sh
docker-compose up -d --build
```
This will:
- Build and start the Go project.
- Start the Python server with 200 connections on port 8500 (default).
- Launch Prometheus for collecting metrics.

#### 2. Manual Execution
SSH inside the container and run
```sh
go run .
```
Ensure that the environment variables are set:
```sh
export CONSUL_PORT=8500
export UPDATE_DELAY=1s
```
Metrics will be written to:
```
/tmp/node-exporter/tcp_connections.prom
```

---

## Package Overview (`main.go`)
The Go program monitors active TCP connections on port 8500 and exports the count as a Prometheus metric.

---

## Design Decisions & Trade-offs
- Data Collection: Used `netstat -tan` to count established TCP connections.  
  `netstat` was chosen for simplicity.  

- Metrics Export: Used the Prometheus Node Exporter textfile collector method.  

- IPv6 Compatibility: Added support for both IPv4 & IPv6.  
  Metrics are split for better tracking:
  ```
  tcp_connections{port="8500", protocol="ipv4"} 150
  tcp_connections{port="8500", protocol="ipv6"} 10
  ```

- Handling `TIME_WAIT` connections: Ignored since Consul does not count them towards the connection limit.

---

## Test Files & Modifications
- `port-opener.py` (provided script):  
  - No change was done

- Manual verification:  
  - Metrics can be checked via Prometheus at:  
    ```
    http://localhost:9090
    ```

## Feedback
### Understanding the Problem

- The problem was well described and easy to understand.

- The provided Python test script made it easier to validate the solution.

### Challenges Faced
- Ensuring compatibility with Prometheus Node Exporter.
- Handling IPv6 and `0.0.0.0` correctly.

### Suggestions for Improvement
- Clarify whether per-client connections need to be tracked separately.
- Maybe Provide example Prometheus alert rules to detect when Consul is close to its limit like:
  ```yaml
  - alert: ConsulHighConnections
    expr: tcp_connections{port="8500"} > 180
    for: 5m
    labels:
      severity: warning
    annotations:
      summary: "Consul is approaching connection limit ({{ $value }} connections)"
  ```
