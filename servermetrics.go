package servermetrics

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
)

// CPUStats represents the CPU stats at a point in time
type CPUStats struct {
	User             uint64   `json:"user"`
	Nice             uint64   `json:"nice"`
	System           uint64   `json:"system"`
	Idle             uint64   `json:"idle"`
	Iowait           uint64   `json:"iowait"`
	Irq              uint64   `json:"irq"`
	Softirq          uint64   `json:"softirq"`
	Steal            uint64   `json:"steal"`
	Guest            uint64   `json:"guest"`
	GuestNice        uint64   `json:"guest_nice"`
	ActiveTime       uint64   `json:"active_time"`
	IdleTime         uint64   `json:"idle_time"`
	TotalTime        uint64   `json:"total_time"`
	UsedPct          float64  `json:"usedpct"`
	UsedPctSinceBoot *float64 `json:"usedpct_since_boot,omitempty"`
}

// MemoryStats represents the memory usage stats at a point in time
type MemoryStats struct {
	Total     uint64  `json:"total"`
	Available uint64  `json:"available"`
	Used      uint64  `json:"used"`
	UsedPct   float64 `json:"usedpct"`
	// SwapTotal is 0 on a host with no swap configured at all -- callers
	// distinguish "no swap" from "swap present but unused" by checking this
	// rather than SwapUsedPct, which is 0 in both cases.
	SwapTotal   uint64  `json:"swap_total"`
	SwapUsed    uint64  `json:"swap_used"`
	SwapUsedPct float64 `json:"swap_usedpct"`
}

// DiskStats represents the disk space usage stats at a point in time
type DiskStats struct {
	Total   uint64  `json:"total"`
	Free    uint64  `json:"free"`
	Used    uint64  `json:"used"`
	UsedPct float64 `json:"usedpct"`
}

// ContainerStats represents the Docker container stats at a point in time
type ContainerStats struct {
	ContainerID   string  `json:"container_id"`
	ContainerName string  `json:"container_name"`
	CPUPct        float64 `json:"cpu_pct"`
	MemUsage      string  `json:"mem_usage"`
	MemLimit      string  `json:"mem_limit"`
	MemPct        float64 `json:"mem_pct"`
	NetIO         string  `json:"net_io"`
	BlockIO       string  `json:"block_io"`
	PIDs          uint64  `json:"pids"`
}

// ContainerInfo represents a Docker container's identity and lifecycle state,
// independent of whether it is currently running. Unlike ContainerStats
// (populated by `docker stats`, running containers only), this comes from
// `docker ps -a` and is the only source that ever reports a stopped/exited
// container.
type ContainerInfo struct {
	ContainerID   string `json:"container_id"`
	ContainerName string `json:"container_name"`
	Image         string `json:"image"`
	State         string `json:"state"`  // "running", "exited", "created", "paused", ...
	Status        string `json:"status"` // human string, e.g. "Up 3 hours" / "Exited (0) 2 days ago"
	Ports         string `json:"ports"`  // raw docker ps ports column, e.g. "0.0.0.0:8080->80/tcp"
}

// GetCPUStats reads CPU stats from /proc/stat and returns a CPUStats struct
func GetCPUStats() (CPUStats, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return CPUStats{}, err
	}

	lines := strings.Split(string(data), "\n")
	cpuLine := strings.Fields(lines[0])

	if len(cpuLine) < 11 {
		return CPUStats{}, fmt.Errorf("unexpected format in /proc/stat")
	}

	user, err := strconv.ParseUint(cpuLine[1], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	nice, err := strconv.ParseUint(cpuLine[2], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	system, err := strconv.ParseUint(cpuLine[3], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	idle, err := strconv.ParseUint(cpuLine[4], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	iowait, err := strconv.ParseUint(cpuLine[5], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	irq, err := strconv.ParseUint(cpuLine[6], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	softirq, err := strconv.ParseUint(cpuLine[7], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	steal, err := strconv.ParseUint(cpuLine[8], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	guest, err := strconv.ParseUint(cpuLine[9], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}
	guestNice, err := strconv.ParseUint(cpuLine[10], 10, 64)
	if err != nil {
		return CPUStats{}, err
	}

	activeTime := user + nice + system + irq + softirq + steal + guest + guestNice
	idleTime := idle + iowait
	totalTime := activeTime + idleTime
	usedPct := float64(activeTime) / float64(totalTime) * 100.0

	return CPUStats{
		User:       user,
		Nice:       nice,
		System:     system,
		Idle:       idle,
		Iowait:     iowait,
		Irq:        irq,
		Softirq:    softirq,
		Steal:      steal,
		Guest:      guest,
		GuestNice:  guestNice,
		ActiveTime: activeTime,
		IdleTime:   idleTime,
		TotalTime:  totalTime,
		UsedPct:    usedPct,
	}, nil
}

// GetMemoryStats reads memory stats from /proc/meminfo and returns a MemoryStats struct
func GetMemoryStats() (MemoryStats, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return MemoryStats{}, err
	}

	var total, available, swapTotal, swapFree uint64

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		key := fields[0][:len(fields[0])-1] // Remove trailing ':'
		value, err := strconv.ParseUint(fields[1], 10, 64)
		if err != nil {
			return MemoryStats{}, err
		}

		switch key {
		case "MemTotal":
			total = value
		case "MemAvailable":
			available = value
		case "SwapTotal":
			swapTotal = value
		case "SwapFree":
			swapFree = value
		}
	}

	if total == 0 {
		return MemoryStats{}, fmt.Errorf("could not determine total memory")
	}

	used := total - available

	// Calculate the used percentage
	usedPct := (float64(used) / float64(total)) * 100.0

	swapUsed := swapTotal - swapFree
	var swapUsedPct float64
	if swapTotal > 0 {
		swapUsedPct = (float64(swapUsed) / float64(swapTotal)) * 100.0
	}

	return MemoryStats{
		Total:       total,
		Available:   available,
		Used:        used,
		UsedPct:     usedPct,
		SwapTotal:   swapTotal,
		SwapUsed:    swapUsed,
		SwapUsedPct: swapUsedPct,
	}, nil
}

// GetDiskStats reads disk usage stats for the given path and returns a DiskStats struct
func GetDiskStats(path string) (DiskStats, error) {
	var stat syscall.Statfs_t

	err := syscall.Statfs(path, &stat)
	if err != nil {
		return DiskStats{}, err
	}

	// Total blocks multiplied by block size
	total := stat.Blocks * uint64(stat.Bsize)

	// Free blocks multiplied by block size
	free := stat.Bfree * uint64(stat.Bsize)

	// Used space is total minus free
	used := total - free

	// Calculate the used percentage
	usedPct := (float64(used) / float64(total)) * 100.0

	return DiskStats{
		Total:   total,
		Free:    free,
		Used:    used,
		UsedPct: usedPct,
	}, nil
}

// GetContainerStats gets Docker container statistics for all running containers
func GetContainerStats() ([]ContainerStats, error) {
	// Check if docker command is available
	if _, err := exec.LookPath("docker"); err != nil {
		return []ContainerStats{}, nil
	}

	// Execute docker stats command with --no-stream to get a single snapshot
	cmd := exec.Command("docker", "stats", "--no-stream", "--format", "table {{.Container}}\t{{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.MemPerc}}\t{{.NetIO}}\t{{.BlockIO}}\t{{.PIDs}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker stats: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return []ContainerStats{}, nil
	}

	var containers []ContainerStats

	// Skip the header line and process container data
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 8 {
			continue
		}

		containerID := strings.TrimSpace(fields[0])
		containerName := strings.TrimSpace(fields[1])
		cpuPercStr := strings.TrimSpace(fields[2])
		memUsage := strings.TrimSpace(fields[3])
		memPercStr := strings.TrimSpace(fields[4])
		netIO := strings.TrimSpace(fields[5])
		blockIO := strings.TrimSpace(fields[6])
		pidsStr := strings.TrimSpace(fields[7])

		// Parse CPU percentage
		cpuPercStr = strings.TrimSuffix(cpuPercStr, "%")
		cpuPct, err := strconv.ParseFloat(cpuPercStr, 64)
		if err != nil {
			continue
		}

		// Parse memory percentage
		memPercStr = strings.TrimSuffix(memPercStr, "%")
		memPct, err := strconv.ParseFloat(memPercStr, 64)
		if err != nil {
			continue
		}

		// Parse memory usage and limit
		var memUsagePart, memLimitPart string
		if slashIndex := strings.Index(memUsage, "/"); slashIndex != -1 {
			memUsagePart = strings.TrimSpace(memUsage[:slashIndex])
			memLimitPart = strings.TrimSpace(memUsage[slashIndex+1:])
		} else {
			memUsagePart = memUsage
			memLimitPart = ""
		}

		// Parse PIDs
		pids, err := strconv.ParseUint(pidsStr, 10, 64)
		if err != nil {
			pids = 0
		}

		containers = append(containers, ContainerStats{
			ContainerID:   containerID,
			ContainerName: containerName,
			CPUPct:        cpuPct,
			MemUsage:      memUsagePart,
			MemLimit:      memLimitPart,
			MemPct:        memPct,
			NetIO:         netIO,
			BlockIO:       blockIO,
			PIDs:          pids,
		})
	}

	return containers, nil
}

// GetContainerList gets Docker container identity/lifecycle-state information
// for every container on the host, running or not. Unlike GetContainerStats
// (`docker stats`, running containers only), this uses `docker ps -a` so a
// stopped/exited container is still reported.
func GetContainerList() ([]ContainerInfo, error) {
	// Check if docker command is available
	if _, err := exec.LookPath("docker"); err != nil {
		return []ContainerInfo{}, nil
	}

	// Execute docker ps -a to get every container regardless of state
	cmd := exec.Command("docker", "ps", "-a", "--format", "table {{.ID}}\t{{.Names}}\t{{.Image}}\t{{.State}}\t{{.Status}}\t{{.Ports}}")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to execute docker ps: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	if len(lines) < 2 {
		return []ContainerInfo{}, nil
	}

	var containers []ContainerInfo

	// Skip the header line and process container data
	for i := 1; i < len(lines); i++ {
		line := strings.TrimSpace(lines[i])
		if line == "" {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) < 5 {
			continue
		}

		// Ports is often empty (a container with no exposed ports), which
		// Split still accounts for as an empty trailing field -- but only
		// when the format string actually emitted the column, hence the
		// >= 5 (not == 6) check above and the defensive index below.
		ports := ""
		if len(fields) >= 6 {
			ports = strings.TrimSpace(fields[5])
		}

		containers = append(containers, ContainerInfo{
			ContainerID:   strings.TrimSpace(fields[0]),
			ContainerName: strings.TrimSpace(fields[1]),
			Image:         strings.TrimSpace(fields[2]),
			State:         strings.TrimSpace(fields[3]),
			Status:        strings.TrimSpace(fields[4]),
			Ports:         ports,
		})
	}

	return containers, nil
}

// CalculateCPUUsage calculates the CPU usage between two snapshots of CPUStats
func CalculateCPUUsage(prevStats, currentStats CPUStats) CPUStats {
	calcDelta := func(current, prev uint64) uint64 {
		return current - prev
	}

	return CPUStats{
		User:             calcDelta(currentStats.User, prevStats.User),
		Nice:             calcDelta(currentStats.Nice, prevStats.Nice),
		System:           calcDelta(currentStats.System, prevStats.System),
		Idle:             calcDelta(currentStats.Idle, prevStats.Idle),
		Iowait:           calcDelta(currentStats.Iowait, prevStats.Iowait),
		Irq:              calcDelta(currentStats.Irq, prevStats.Irq),
		Softirq:          calcDelta(currentStats.Softirq, prevStats.Softirq),
		Steal:            calcDelta(currentStats.Steal, prevStats.Steal),
		Guest:            calcDelta(currentStats.Guest, prevStats.Guest),
		GuestNice:        calcDelta(currentStats.GuestNice, prevStats.GuestNice),
		ActiveTime:       calcDelta(currentStats.ActiveTime, prevStats.ActiveTime),
		IdleTime:         calcDelta(currentStats.IdleTime, prevStats.IdleTime),
		TotalTime:        calcDelta(currentStats.TotalTime, prevStats.TotalTime),
		UsedPct:          float64(calcDelta(currentStats.ActiveTime, prevStats.ActiveTime)) / float64(calcDelta(currentStats.TotalTime, prevStats.TotalTime)) * 100.0,
		UsedPctSinceBoot: &currentStats.UsedPct,
	}
}
