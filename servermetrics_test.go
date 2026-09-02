package servermetrics

import (
	"os"
	"strings"
	"testing"
)

func TestCPUStatsStruct(t *testing.T) {
	stats := CPUStats{
		User:             1000,
		Nice:             100,
		System:           500,
		Idle:             9000,
		Iowait:           200,
		Irq:              50,
		Softirq:          30,
		Steal:            10,
		Guest:            5,
		GuestNice:        2,
		ActiveTime:       1697,
		IdleTime:         9200,
		TotalTime:        10897,
		UsedPct:          15.57,
		UsedPctSinceBoot: nil,
		VCPUs:            4,
	}

	if stats.User != 1000 {
		t.Errorf("Expected User 1000, got %d", stats.User)
	}
	if stats.UsedPct != 15.57 {
		t.Errorf("Expected UsedPct 15.57, got %f", stats.UsedPct)
	}
	if stats.VCPUs != 4 {
		t.Errorf("Expected VCPUs 4, got %d", stats.VCPUs)
	}
}

func TestMemoryStatsStruct(t *testing.T) {
	stats := MemoryStats{
		Total:       8388608,
		Available:   4194304,
		Used:        4194304,
		UsedPct:     50.0,
		SwapTotal:   2097152,
		SwapUsed:    1048576,
		SwapUsedPct: 50.0,
	}

	if stats.Total != 8388608 {
		t.Errorf("Expected Total 8388608, got %d", stats.Total)
	}
	if stats.UsedPct != 50.0 {
		t.Errorf("Expected UsedPct 50.0, got %f", stats.UsedPct)
	}
	if stats.SwapTotal != 2097152 {
		t.Errorf("Expected SwapTotal 2097152, got %d", stats.SwapTotal)
	}
	if stats.SwapUsedPct != 50.0 {
		t.Errorf("Expected SwapUsedPct 50.0, got %f", stats.SwapUsedPct)
	}
}

func TestDiskStatsStruct(t *testing.T) {
	stats := DiskStats{
		Total:   1000000000,
		Free:    400000000,
		Used:    600000000,
		UsedPct: 60.0,
	}

	if stats.Total != 1000000000 {
		t.Errorf("Expected Total 1000000000, got %d", stats.Total)
	}
	if stats.UsedPct != 60.0 {
		t.Errorf("Expected UsedPct 60.0, got %f", stats.UsedPct)
	}
}

func TestContainerStatsStruct(t *testing.T) {
	stats := ContainerStats{
		ContainerID:   "abc123def456",
		ContainerName: "test-container",
		CPUPct:        25.5,
		MemUsage:      "1.5GiB",
		MemLimit:      "2GiB",
		MemPct:        75.0,
		NetIO:         "1.2MB / 800kB",
		BlockIO:       "100MB / 50MB",
		PIDs:          15,
	}

	if stats.ContainerID != "abc123def456" {
		t.Errorf("Expected ContainerID 'abc123def456', got '%s'", stats.ContainerID)
	}
	if stats.CPUPct != 25.5 {
		t.Errorf("Expected CPUPct 25.5, got %f", stats.CPUPct)
	}
	if stats.PIDs != 15 {
		t.Errorf("Expected PIDs 15, got %d", stats.PIDs)
	}
}

func TestContainerInfoStruct(t *testing.T) {
	info := ContainerInfo{
		ContainerID:   "abc123def456",
		ContainerName: "test-container",
		Image:         "ghcr.io/aliasproject/php:8.4-frankenphp",
		State:         "running",
		Status:        "Up 3 hours",
		Ports:         "0.0.0.0:8080->80/tcp",
	}

	if info.ContainerID != "abc123def456" {
		t.Errorf("Expected ContainerID 'abc123def456', got '%s'", info.ContainerID)
	}
	if info.Image != "ghcr.io/aliasproject/php:8.4-frankenphp" {
		t.Errorf("Expected Image 'ghcr.io/aliasproject/php:8.4-frankenphp', got '%s'", info.Image)
	}
	if info.State != "running" {
		t.Errorf("Expected State 'running', got '%s'", info.State)
	}
}

func TestGetCPUStats(t *testing.T) {
	stats, err := GetCPUStats()

	// On non-Linux systems, this might fail
	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			t.Skip("Skipping CPU stats test - /proc/stat not available (non-Linux system)")
		}
		t.Fatalf("GetCPUStats failed: %v", err)
	}

	// Verify struct has reasonable values
	if stats.TotalTime == 0 {
		t.Error("Expected non-zero TotalTime")
	}
	if stats.UsedPct < 0 || stats.UsedPct > 100 {
		t.Errorf("UsedPct should be 0-100, got %f", stats.UsedPct)
	}
	if stats.ActiveTime+stats.IdleTime != stats.TotalTime {
		t.Error("ActiveTime + IdleTime should equal TotalTime")
	}

	// Test that all fields are properly set
	if stats.User == 0 && stats.System == 0 && stats.Idle == 0 {
		t.Error("Expected some CPU time values to be non-zero")
	}

	if stats.VCPUs < 1 {
		t.Errorf("Expected VCPUs >= 1, got %d", stats.VCPUs)
	}
}

func TestGetMemoryStats(t *testing.T) {
	stats, err := GetMemoryStats()

	if err != nil {
		if strings.Contains(err.Error(), "no such file or directory") {
			t.Skip("Skipping memory stats test - /proc/meminfo not available (non-Linux system)")
		}
		t.Fatalf("GetMemoryStats failed: %v", err)
	}

	// Verify struct has reasonable values
	if stats.Total == 0 {
		t.Error("Expected non-zero Total memory")
	}
	if stats.UsedPct < 0 || stats.UsedPct > 100 {
		t.Errorf("UsedPct should be 0-100, got %f", stats.UsedPct)
	}
	if stats.Used+stats.Available != stats.Total {
		t.Error("Used + Available should equal Total")
	}
	if stats.Used == 0 {
		t.Error("Expected non-zero Used memory")
	}

	// SwapTotal is legitimately 0 on a host with no swap configured, so it
	// isn't asserted non-zero -- only that the derived fields stay
	// internally consistent whatever the host's actual swap setup is.
	if stats.SwapUsed > stats.SwapTotal {
		t.Errorf("SwapUsed (%d) should not exceed SwapTotal (%d)", stats.SwapUsed, stats.SwapTotal)
	}
	if stats.SwapUsedPct < 0 || stats.SwapUsedPct > 100 {
		t.Errorf("SwapUsedPct should be 0-100, got %f", stats.SwapUsedPct)
	}
	if stats.SwapTotal == 0 && stats.SwapUsedPct != 0 {
		t.Error("Expected SwapUsedPct 0 when SwapTotal is 0")
	}
}

func TestGetDiskStats(t *testing.T) {
	// Test with root directory
	stats, err := GetDiskStats("/")

	if err != nil {
		t.Fatalf("GetDiskStats failed: %v", err)
	}

	// Verify struct has reasonable values
	if stats.Total == 0 {
		t.Error("Expected non-zero Total disk space")
	}
	if stats.UsedPct < 0 || stats.UsedPct > 100 {
		t.Errorf("UsedPct should be 0-100, got %f", stats.UsedPct)
	}
	if stats.Used+stats.Free != stats.Total {
		t.Error("Used + Free should equal Total")
	}
}

func TestGetDiskStatsInvalidPath(t *testing.T) {
	_, err := GetDiskStats("/nonexistent/path/that/should/not/exist")

	if err == nil {
		t.Error("Expected error for invalid path")
	}
}

func TestGetContainerStats(t *testing.T) {
	containers, err := GetContainerStats()

	// It's OK if Docker isn't available
	if err != nil {
		if strings.Contains(err.Error(), "docker command not found") ||
			strings.Contains(err.Error(), "failed to execute docker stats") {
			t.Logf("Docker not available: %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If we get here, Docker is available
	t.Logf("Found %d containers", len(containers))

	// Verify container structure if any containers exist
	for _, container := range containers {
		if container.ContainerID == "" {
			t.Error("Container should have non-empty ID")
		}
		if container.ContainerName == "" {
			t.Error("Container should have non-empty name")
		}
		if container.MemPct < 0 || container.MemPct > 100 {
			t.Errorf("Memory percentage should be 0-100, got %f", container.MemPct)
		}
		if container.CPUPct < 0 {
			t.Errorf("CPU percentage should be non-negative, got %f", container.CPUPct)
		}
		// PIDs can be 0, so just check it's not negative
		// NetIO, BlockIO, MemUsage, MemLimit can be any string format
	}
}

func TestGetContainerList(t *testing.T) {
	containers, err := GetContainerList()

	// It's OK if Docker isn't available
	if err != nil {
		if strings.Contains(err.Error(), "docker command not found") ||
			strings.Contains(err.Error(), "failed to execute docker ps") {
			t.Logf("Docker not available: %v", err)
			return
		}
		t.Fatalf("Unexpected error: %v", err)
	}

	// If we get here, Docker is available
	t.Logf("Found %d containers", len(containers))

	// Verify container structure if any containers exist
	for _, container := range containers {
		if container.ContainerID == "" {
			t.Error("Container should have non-empty ID")
		}
		if container.ContainerName == "" {
			t.Error("Container should have non-empty name")
		}
		if container.State == "" {
			t.Error("Container should have non-empty state")
		}
		// Image, Status, Ports can be any string format (Ports is often empty)
	}
}

func TestCalculateCPUUsage(t *testing.T) {
	// Create mock CPU stats for testing
	prevStats := CPUStats{
		User:       1000,
		Nice:       100,
		System:     500,
		Idle:       8000,
		Iowait:     200,
		Irq:        50,
		Softirq:    30,
		Steal:      10,
		Guest:      5,
		GuestNice:  2,
		ActiveTime: 1697,
		IdleTime:   8200,
		TotalTime:  9897,
		UsedPct:    17.15,
		VCPUs:      8, // must not affect the delta calculated below
	}

	currentStats := CPUStats{
		User:       1200,
		Nice:       120,
		System:     600,
		Idle:       8500,
		Iowait:     250,
		Irq:        60,
		Softirq:    40,
		Steal:      15,
		Guest:      8,
		GuestNice:  3,
		ActiveTime: 2046,
		IdleTime:   8750,
		TotalTime:  10796,
		UsedPct:    18.95,
		VCPUs:      4,
	}

	usage := CalculateCPUUsage(prevStats, currentStats)

	// Verify deltas are calculated correctly
	expectedUserDelta := uint64(200) // 1200 - 1000
	if usage.User != expectedUserDelta {
		t.Errorf("Expected User delta %d, got %d", expectedUserDelta, usage.User)
	}

	expectedSystemDelta := uint64(100) // 600 - 500
	if usage.System != expectedSystemDelta {
		t.Errorf("Expected System delta %d, got %d", expectedSystemDelta, usage.System)
	}

	// Verify percentage calculation
	expectedActiveDelta := uint64(349) // 2046 - 1697
	expectedTotalDelta := uint64(899)  // 10796 - 9897
	expectedUsedPct := float64(expectedActiveDelta) / float64(expectedTotalDelta) * 100.0

	if usage.ActiveTime != expectedActiveDelta {
		t.Errorf("Expected ActiveTime delta %d, got %d", expectedActiveDelta, usage.ActiveTime)
	}
	if usage.TotalTime != expectedTotalDelta {
		t.Errorf("Expected TotalTime delta %d, got %d", expectedTotalDelta, usage.TotalTime)
	}

	// Allow small floating point differences
	if usage.UsedPct < expectedUsedPct-0.01 || usage.UsedPct > expectedUsedPct+0.01 {
		t.Errorf("Expected UsedPct ~%.2f, got %.2f", expectedUsedPct, usage.UsedPct)
	}

	// Verify UsedPctSinceBoot is set
	if usage.UsedPctSinceBoot == nil {
		t.Error("UsedPctSinceBoot should be set")
	} else if *usage.UsedPctSinceBoot != currentStats.UsedPct {
		t.Errorf("Expected UsedPctSinceBoot %.2f, got %.2f", currentStats.UsedPct, *usage.UsedPctSinceBoot)
	}

	// VCPUs is a point-in-time count, not a delta -- it should carry
	// through from currentStats unchanged, not be subtracted from prevStats.
	if usage.VCPUs != currentStats.VCPUs {
		t.Errorf("Expected VCPUs to carry through as %d, got %d", currentStats.VCPUs, usage.VCPUs)
	}
}

func TestCalculateCPUUsageZeroValues(t *testing.T) {
	// Test edge case with zero values
	prevStats := CPUStats{}
	currentStats := CPUStats{
		User:       100,
		System:     50,
		Idle:       1000,
		ActiveTime: 150,
		IdleTime:   1000,
		TotalTime:  1150,
		UsedPct:    13.04,
	}

	usage := CalculateCPUUsage(prevStats, currentStats)

	if usage.User != 100 {
		t.Errorf("Expected User 100, got %d", usage.User)
	}
	if usage.UsedPct < 13.0 || usage.UsedPct > 13.1 {
		t.Errorf("Expected UsedPct ~13.04, got %.2f", usage.UsedPct)
	}
}

// Benchmark tests
func BenchmarkGetCPUStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetCPUStats()
	}
}

func BenchmarkGetMemoryStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetMemoryStats()
	}
}

func BenchmarkGetDiskStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetDiskStats("/")
	}
}

func BenchmarkCalculateCPUUsage(b *testing.B) {
	prevStats := CPUStats{
		User: 1000, System: 500, Idle: 8000,
		ActiveTime: 1500, IdleTime: 8000, TotalTime: 9500, UsedPct: 15.8,
	}
	currentStats := CPUStats{
		User: 1100, System: 600, Idle: 8200,
		ActiveTime: 1700, IdleTime: 8200, TotalTime: 9900, UsedPct: 17.2,
	}

	for i := 0; i < b.N; i++ {
		CalculateCPUUsage(prevStats, currentStats)
	}
}

func BenchmarkGetContainerStats(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetContainerStats()
	}
}

func BenchmarkGetContainerList(b *testing.B) {
	for i := 0; i < b.N; i++ {
		GetContainerList()
	}
}

// Test file operations that might be used on different systems
func TestFileOperationsCompatibility(t *testing.T) {
	// Test that we handle missing files gracefully
	_, err := os.ReadFile("/proc/nonexistent")
	if err == nil {
		t.Skip("Unexpected success reading nonexistent file")
	}

	// This tests the same error handling path as GetCPUStats and GetMemoryStats
	if !strings.Contains(err.Error(), "no such file or directory") &&
		!strings.Contains(err.Error(), "cannot find the file") { // Windows
		t.Logf("File error format: %v", err)
	}
}
