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
	User       uint64
	Nice       uint64
	System     uint64
	Idle       uint64
	Iowait     uint64
	Irq        uint64
	Softirq    uint64
	Steal      uint64
	Guest      uint64
	GuestNice  uint64
	ActiveTime uint64
	IdleTime   uint64
	TotalTime  uint64
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
func CalculateCPUUsage(prevStats, currentStats CPUStats) map[string]uint64 {
	usage := make(map[string]uint64)

	calcDelta := func(current, prev uint64) uint64 {
		return current - prev
	}

	usage["user"] = calcDelta(currentStats.User, prevStats.User)
	usage["nice"] = calcDelta(currentStats.Nice, prevStats.Nice)
	usage["system"] = calcDelta(currentStats.System, prevStats.System)
	usage["idle"] = calcDelta(currentStats.Idle, prevStats.Idle)
	usage["iowait"] = calcDelta(currentStats.Iowait, prevStats.Iowait)
	usage["irq"] = calcDelta(currentStats.Irq, prevStats.Irq)
	usage["softirq"] = calcDelta(currentStats.Softirq, prevStats.Softirq)
	usage["steal"] = calcDelta(currentStats.Steal, prevStats.Steal)
	usage["guest"] = calcDelta(currentStats.Guest, prevStats.Guest)
	usage["guestNice"] = calcDelta(currentStats.GuestNice, prevStats.GuestNice)

	usage["active"] = calcDelta(currentStats.ActiveTime, prevStats.ActiveTime)
	usage["idle"] = calcDelta(currentStats.IdleTime, prevStats.IdleTime)
	usage["total"] = calcDelta(currentStats.TotalTime, prevStats.TotalTime)

	return usage
}
