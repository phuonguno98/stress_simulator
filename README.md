# Stress Simulator 🚀

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/phuonguno98/stress_simulator)](https://goreportcard.com/report/github.com/phuonguno98/stress_simulator)
[![Go Version](https://img.shields.io/badge/Go-1.26.2-blue.svg)](https://golang.org/)

**Stress Simulator** is a professional system load simulation tool designed to generate continuous, unpredictable, and realistic resource utilization patterns. It mimics the behavior of a production service, making it ideal for testing monitoring systems, alerting rules, and performance analysis tools without needing actual user traffic.

---

## 🌟 Key Features

- **Comprehensive Resource Simulation**:
  - 💻 **CPU**: High utilization and **True OS-level I/O wait** simulation (using `O_SYNC` blocking writes).
  - 🧠 **Memory**: Dynamic memory consumption.
  - 💾 **Disk**: Realistic Disk I/O patterns bypassing page caches for true hardware load.
  - 🌐 **Network**: Accurate, chunk-based, bi-directional traffic (In/Out) simulation enforcing strict MB/s rates.
- **Realistic Load Patterns**:
  - **Smooth Random Walk**: Avoids abrupt jumps; resource graphs move naturally.
  - **Seasonality**: Built-in sine-wave patterns based on time-of-day and day-of-week. Implements a sharp drop outside business hours for realistic off-peak behavior.
  - **Unpredictable Anomalies**: Bursty spikes and dips based on Exponential Distribution, with strict time overlap prevention.
- **Production-Ready & Safe Design**:
  - **Zero-Allocation Payload Pool**: Pre-allocated 1MB buffer to prevent CPU spikes during network/disk stress.
  - **Crash-Resistant Disk Guard**: Automatic pre-startup cleanup, deterministic file handling, and aggressive 10-second Garbage Collection to guarantee zero disk exhaustion or temp file leakage, even on `SIGKILL`.
  - **Auto-Configuration**: Automatically adapts to system memory, network ports, and environment.
- **Continuous Operation**: Designed to run indefinitely (`-forever` mode) until manually terminated.
- **Modern Architecture**: Optimized for minimal binary size and supports both `amd64` and `arm64` (Apple Silicon/Graviton) deployments.

## 🛠 Installation

### Prerequisites
- **OS**: Ubuntu 24.04 (Recommended), Windows, macOS
- **Language**: Go 1.26.2

### Setup
```bash
# Clone the repository
git clone https://github.com/phuonguno98/stress_simulator.git
cd stress_simulator

# Build the binary
go build -trimpath -ldflags="-s -w" -o stress_simulator
```

## 🚀 Usage

Run the simulator with the following command:

```bash
./stress_simulator -cpu 50.0 -iowait 20.0 -memory 50.0 -disk 50.0 -network 10.0 -target 10.82.14.32 -forever
```

### Configuration Flags

| Flag | Description | Range/Example |
| :--- | :--- | :--- |
| `-cpu` | Average CPU utilization percentage | `0.0` - `100.0` |
| `-iowait` | Average CPU I/O wait percentage | `0.0` - `100.0` |
| `-memory` | Average Memory utilization percentage | `0.0` - `100.0` |
| `-disk` | Average Disk utilization percentage | `0.0` - `100.0` |
| `-network` | Average network throughput (MB/s) | e.g., `10.0` |
| `-target` | Target IP address(es) for network stress | `10.82.14.31,10.82.14.32` |
| `-forever` | Run indefinitely with seasonal patterns | Boolean flag |
| `-version` | Print version information and exit | Boolean flag |

> **Note**: The default listener port for incoming network traffic is `8080`.

## 🔬 Advanced Load Mechanisms

1. **Smooth Random Walk**: Instead of random jumps, indices use a random walk algorithm to change gradually, creating natural-looking telemetry.
2. **Sine Wave Seasonality**: Uses a sine function mapped to the 24-hour clock. Load peaks aggressively during business hours (07:00-11:00, 13:00-17:00) and experiences a sharp baseline drop during evenings and weekends.
3. **Exponential Distribution Anomalies**: Spikes and dips are governed by an exponential distribution, ensuring "bursty" behavior that is impossible to predict. Normal seasonal updates are paused during an active anomaly.
4. **Bi-directional Network Traffic**: Opens a listener port (default `8080`) to consume data from other nodes while simultaneously sending highly precise byte-chunked traffic, simulating real inter-node communication.
5. **Pre-allocated Payload Pool**: To avoid "Double-Stress" (where network/disk activity causes CPU spikes due to memory allocation), a 1MB shared buffer is generated at startup.
6. **Disk Limit Guard**: A highly aggressive 10-second GC interval continuously monitors the temporary data directory, capping storage at 500MB and utilizing fixed file pointers to prevent OS temp directory leaks on hard crashes.

## 🧪 Unit Tests

The project includes comprehensive unit tests for all modules:

To run all tests with race detection and coverage:
```bash
go test -v -race -cover ./...
```

## 🧹 Linting

The project uses `golangci-lint` for strict code quality checks.

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.11.4

# Run linter
golangci-lint run
```

Configuration is stored in [.golangci.yml](.golangci.yml).

## ⚠️ Important Notes

- **Environment**: Only run this tool in isolated testing environments. It is designed to consume significant system resources.
- **Connectivity**: Ensure the target IP addresses are reachable.
- **Termination**: When using `-forever`, the process will run until it is killed (e.g., `Ctrl+C` or `kill -9`).

## 🤝 Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

**Maintained by [Nguyễn Thanh Phương](https://github.com/phuonguno98)**  
**Contact**: [t.me/phuonguno](https://t.me/phuonguno)
