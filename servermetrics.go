package cpustats

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"syscall"
)

// CPUStats represents the CPU stats at a point in time
type CPUStats struct {
	User   uint64
	Nice   uint64
	System uint64
	Idle   uint64
	Iowait uint64
	Total  uint64
}

// MemoryStats represents the memory usage stats at a point in time
type MemoryStats struct {
	Total     uint64
	Available uint64
	Used      uint64
}

// DiskStats represents the disk space usage stats at a point in time
type DiskStats struct {
	Total   uint64
	Free    uint64
	Used    uint64
	UsedPct float64
}

// GetCPUStats reads CPU stats from /proc/stat and returns a CPUStats struct
func GetCPUStats() (CPUStats, error) {
	data, err := os.ReadFile("/proc/stat")
	if err != nil {
		return CPUStats{}, err
	}

	lines := strings.Split(string(data), "\n")
	cpuLine := strings.Fields(lines[0])

	if len(cpuLine) < 8 {
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

	total := user + nice + system + idle + iowait

	return CPUStats{
		User:   user,
		Nice:   nice,
		System: system,
		Idle:   idle,
		Iowait: iowait,
		Total:  total,
	}, nil
}

// GetMemoryStats reads memory stats from /proc/meminfo and returns a MemoryStats struct
func GetMemoryStats() (MemoryStats, error) {
	data, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return MemoryStats{}, err
	}

	var total, available uint64

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
		}
	}

	if total == 0 {
		return MemoryStats{}, fmt.Errorf("could not determine total memory")
	}

	used := total - available

	return MemoryStats{
		Total:     total,
		Available: available,
		Used:      used,
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

// CalculateCPUUsage calculates the CPU usage between two snapshots of CPUStats
func CalculateCPUUsage(prevStats, currentStats CPUStats) float64 {
	totalDelta := currentStats.Total - prevStats.Total
	idleDelta := currentStats.Idle - prevStats.Idle

	if totalDelta == 0 {
		return 0.0
	}

	return 100.0 * (1.0 - float64(idleDelta)/float64(totalDelta))
}
