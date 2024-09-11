package servermetrics

import (
	"fmt"
	"os"
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
}

// DiskStats represents the disk space usage stats at a point in time
type DiskStats struct {
	Total   uint64  `json:"total"`
	Free    uint64  `json:"free"`
	Used    uint64  `json:"used"`
	UsedPct float64 `json:"usedpct"`
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

	// Calculate the used percentage
	usedPct := (float64(used) / float64(total)) * 100.0

	return MemoryStats{
		Total:     total,
		Available: available,
		Used:      used,
		UsedPct:   usedPct,
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
