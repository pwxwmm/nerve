// Package sysinfo provides system information collection functionality.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package sysinfo

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Hostname returns the system hostname
func Hostname() string {
	hostname, err := exec.Command("hostname").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(string(hostname))
}

// OS returns the operating system information
func OS() string {
	return fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)
}

// GetCPUData returns CPU type and logical cores
func GetCPUData() (string, int) {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("lscpu").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			var model string
			cores := runtime.NumCPU()
			
			for _, line := range lines {
				if strings.HasPrefix(line, "Model name:") {
					model = strings.TrimSpace(strings.Split(line, ":")[1])
				}
			}
			return model, cores
		}
	}
	
	return "Unknown", runtime.NumCPU()
}

// GetMemory returns total memory in KB and formatted string
func GetMemory() (int64, string) {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("free", "-k").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			if len(lines) > 1 {
				fields := strings.Fields(lines[1])
				if len(fields) > 1 {
					var total int64
					fmt.Sscanf(fields[1], "%d", &total)
					return total, formatSize(total * 1024)
				}
			}
		}
	}
	return 0, "Unknown"
}

// GetSN returns the system serial number
func GetSN() string {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("dmidecode", "-s", "system-serial-number").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return "Unknown"
}

// GetProduct returns the product name
func GetProduct() string {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("dmidecode", "-s", "system-product-name").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return "Unknown"
}

// GetBrand returns the manufacturer
func GetBrand() string {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("dmidecode", "-s", "system-manufacturer").Output()
		if err == nil {
			return strings.TrimSpace(string(out))
		}
	}
	return "Unknown"
}

// GetNetcard returns list of network interfaces
func GetNetcard() []string {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("ls", "/sys/class/net").Output()
		if err == nil {
			interfaces := strings.Fields(string(out))
			return interfaces
		}
	}
	return []string{}
}

// Basearch returns the base architecture
func Basearch() string {
	return runtime.GOARCH
}

// Disk returns disk information
func Disk() map[string]interface{} {
	return map[string]interface{}{
		"total": "Unknown",
		"usage": "Unknown",
	}
}

// Raid returns RAID controller information
func Raid() string {
	if runtime.GOOS == "linux" {
		if _, err := exec.Command("which", "megacli").Output(); err == nil {
			return "MegaRAID"
		}
		if _, err := exec.Command("which", "mdadm").Output(); err == nil {
			return "mdadm"
		}
	}
	return "None"
}

// IPMI returns IPMI IP address
func IPMI() string {
	if runtime.GOOS == "linux" {
		out, err := exec.Command("ipmitool", "lan", "print", "1").Output()
		if err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.Contains(line, "IP Address") {
					fields := strings.Fields(line)
					if len(fields) > 3 {
						return fields[3]
					}
				}
			}
		}
	}
	return ""
}

// ManagerIP returns management IP
func ManagerIP() string {
	// TODO: Implement management IP detection
	return ""
}

// StorageIP returns storage IP
func StorageIP(cluster string) string {
	// TODO: Implement storage IP detection
	return ""
}

// ParamIP returns parameter IP
func ParamIP() string {
	// TODO: Implement parameter IP detection
	return ""
}

// GPUInfo returns GPU information
func GPUInfo() map[string]interface{} {
	result := map[string]interface{}{
		"count":    0,
		"type":     "",
		"vendors":  []string{},
		"info":     []map[string]interface{}{},
	}

	if runtime.GOOS == "linux" {
		// Check NVIDIA
		if out, err := exec.Command("nvidia-smi", "--list-gpus").Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			result["count"] = len(lines)
			result["type"] = "NVIDIA"
			result["vendors"] = []string{"NVIDIA"}
		}
	}

	return result
}

// GetDiskInfo returns detailed disk information
func GetDiskInfo() []map[string]interface{} {
	return []map[string]interface{}{}
}

// GetMemoryInfo returns detailed memory information
func GetMemoryInfo() []map[string]interface{} {
	return []map[string]interface{}{}
}

// GetCPUInfo returns detailed CPU information
func GetCPUInfo() map[string]interface{} {
	cpuType, cpuCores := GetCPUData()
	return map[string]interface{}{
		"model":  cpuType,
		"cores":  cpuCores,
		"arch":   runtime.GOARCH,
		"vendor": "Unknown",
	}
}

// GetGPUInfos returns detailed GPU information
func GetGPUInfos() []map[string]interface{} {
	return []map[string]interface{}{}
}

// GetNetworkInfo returns detailed network information
func GetNetworkInfo() []map[string]interface{} {
	return []map[string]interface{}{}
}

// formatSize formats bytes to human-readable format
func formatSize(bytes int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB"}
	size := float64(bytes)
	unit := 0
	
	for size >= 1024 && unit < len(units)-1 {
		size /= 1024
		unit++
	}
	
	return fmt.Sprintf("%.2f %s", size, units[unit])
}

