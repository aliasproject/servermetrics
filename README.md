# ServerMetrics

A Go package for collecting server and container metrics from Linux systems.

## Features

- **CPU Statistics**: Get detailed CPU usage statistics from `/proc/stat`
- **Memory Statistics**: Get memory usage information from `/proc/meminfo`
- **Disk Statistics**: Get disk space usage for any filesystem path
- **Docker Container Statistics**: Get comprehensive stats for all running Docker containers
- **CPU Usage Calculation**: Calculate CPU usage between two time snapshots

## Installation

```bash
go get github.com/aliasproject/servermetrics
```

## Usage

### Basic System Metrics

```go
package main

import (
    "fmt"
    "log"
    "github.com/aliasproject/servermetrics"
)

func main() {
    // Get CPU stats
    cpuStats, err := servermetrics.GetCPUStats()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("CPU Usage: %.2f%%\n", cpuStats.UsedPct)

    // Get memory stats
    memStats, err := servermetrics.GetMemoryStats()
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Memory Usage: %.2f%% (%d KB used of %d KB total)\n",
        memStats.UsedPct, memStats.Used, memStats.Total)

    // Get disk stats for root partition
    diskStats, err := servermetrics.GetDiskStats("/")
    if err != nil {
        log.Fatal(err)
    }
    fmt.Printf("Disk Usage: %.2f%% (%d bytes used of %d bytes total)\n",
        diskStats.UsedPct, diskStats.Used, diskStats.Total)
}
```

### Docker Container Metrics

```go
package main

import (
    "fmt"
    "log"
    "github.com/aliasproject/servermetrics"
)

func main() {
    // Get Docker container stats
    containers, err := servermetrics.GetContainerStats()
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Found %d running containers:\n\n", len(containers))

    for _, container := range containers {
        fmt.Printf("Container: %s (%s)\n", container.ContainerName, container.ContainerID[:12])
        fmt.Printf("  CPU: %.2f%%\n", container.CPUPct)
        fmt.Printf("  Memory: %s / %s (%.2f%%)\n", container.MemUsage, container.MemLimit, container.MemPct)
        fmt.Printf("  Network I/O: %s\n", container.NetIO)
        fmt.Printf("  Block I/O: %s\n", container.BlockIO)
        fmt.Printf("  PIDs: %d\n\n", container.PIDs)
    }
}
```

### CPU Usage Over Time

```go
package main

import (
    "fmt"
    "log"
    "time"
    "github.com/aliasproject/servermetrics"
)

func main() {
    // Get initial CPU stats
    prevStats, err := servermetrics.GetCPUStats()
    if err != nil {
        log.Fatal(err)
    }

    // Wait a second
    time.Sleep(1 * time.Second)

    // Get current CPU stats
    currentStats, err := servermetrics.GetCPUStats()
    if err != nil {
        log.Fatal(err)
    }

    // Calculate CPU usage for the interval
    usage := servermetrics.CalculateCPUUsage(prevStats, currentStats)
    fmt.Printf("CPU Usage over last second: %.2f%%\n", usage.UsedPct)
}
```

## API Reference

### Types

#### CPUStats

Contains detailed CPU timing information:

- `User`, `Nice`, `System`, `Idle`, `Iowait`, `Irq`, `Softirq`, `Steal`, `Guest`, `GuestNice`: Raw CPU time values
- `ActiveTime`, `IdleTime`, `TotalTime`: Calculated time totals
- `UsedPct`: CPU usage percentage
- `UsedPctSinceBoot`: CPU usage percentage since boot (only set by `CalculateCPUUsage`)

#### MemoryStats

Contains memory usage information:

- `Total`: Total system memory in KB
- `Available`: Available memory in KB
- `Used`: Used memory in KB
- `UsedPct`: Memory usage percentage
- `SwapTotal`: Total swap space in KB (`0` if no swap is configured)
- `SwapUsed`: Swap space in use, in KB
- `SwapUsedPct`: Swap usage percentage (`0` when `SwapTotal` is `0`, not a divide-by-zero)

#### DiskStats

Contains disk space information:

- `Total`: Total disk space in bytes
- `Free`: Free disk space in bytes
- `Used`: Used disk space in bytes
- `UsedPct`: Disk usage percentage

#### ContainerStats

Contains Docker container statistics:

- `ContainerID`: Docker container ID
- `ContainerName`: Container name
- `CPUPct`: CPU usage percentage
- `MemUsage`: Memory usage (e.g., "1.5GiB")
- `MemLimit`: Memory limit (e.g., "2GiB")
- `MemPct`: Memory usage percentage
- `NetIO`: Network I/O statistics
- `BlockIO`: Block I/O statistics
- `PIDs`: Number of processes/threads in the container

### Functions

#### `GetCPUStats() (CPUStats, error)`

Reads CPU statistics from `/proc/stat`.

#### `GetMemoryStats() (MemoryStats, error)`

Reads memory statistics from `/proc/meminfo`.

#### `GetDiskStats(path string) (DiskStats, error)`

Gets disk usage statistics for the specified filesystem path.

#### `GetContainerStats() ([]ContainerStats, error)`

Gets statistics for all running Docker containers. Requires Docker to be installed and accessible.

#### `CalculateCPUUsage(prevStats, currentStats CPUStats) CPUStats`

Calculates CPU usage between two CPU statistics snapshots.

## Requirements

- Linux operating system (uses `/proc` filesystem)
- For container stats: Docker installed and accessible via the `docker` command
- Go 1.23.0 or later

## Notes

- All memory values from system calls are in kilobytes (KB)
- All disk space values are in bytes
- Container stats require the Docker daemon to be running
- The package uses direct system calls and file reads for efficiency
- CPU percentages are calculated based on time spent in different CPU states

## License

MIT — see [LICENSE](LICENSE).
