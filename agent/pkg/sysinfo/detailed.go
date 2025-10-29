// Package sysinfo provides detailed system information collection functionality.
//
// Author: mmwei3 (2025-10-28)
// Wethers: cloudWays
package sysinfo

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
)

// GetDetailedCPUInfo returns detailed CPU information
func GetDetailedCPUInfo() map[string]interface{} {
	info := map[string]interface{}{
		"model":         "Unknown",
		"vendor":        "Unknown",
		"family":        "Unknown",
		"model_number": "Unknown",
		"stepping":      "Unknown",
		"microcode":     "Unknown",
		"cpus":          runtime.NumCPU(),
		"cache":         map[string]string{},
		"flags":         []string{},
		"freq_base":     "Unknown",
		"freq_max":      "Unknown",
		"freq_min":      "Unknown",
	}

	if runtime.GOOS == "linux" {
		// Get CPU info from /proc/cpuinfo
		if out, err := exec.Command("cat", "/proc/cpuinfo").Output(); err == nil {
			cpuinfo := parseCPUInfo(string(out))
			for k, v := range cpuinfo {
				info[k] = v
			}
		}

		// Get CPU frequency
		if out, err := exec.Command("lscpu").Output(); err == nil {
			lines := strings.Split(string(out), "\n")
			for _, line := range lines {
				if strings.HasPrefix(line, "Model name:") {
					info["model"] = strings.TrimSpace(strings.Split(line, ":")[1])
				}
				if strings.HasPrefix(line, "CPU max MHz:") {
					info["freq_max"] = strings.TrimSpace(strings.Split(line, ":")[1]) + " MHz"
				}
				if strings.HasPrefix(line, "CPU min MHz:") {
					info["freq_min"] = strings.TrimSpace(strings.Split(line, ":")[1]) + " MHz"
				}
			}
		}
	}

	return info
}

// GetDetailedMemoryInfo returns detailed memory information including DIMM info
func GetDetailedMemoryInfo() []map[string]interface{} {
	var dimms []map[string]interface{}

	if runtime.GOOS == "linux" {
		// Get memory devices from dmidecode
		if out, err := exec.Command("dmidecode", "-t", "memory").Output(); err == nil {
			dimms = parseMemoryDevices(string(out))
		}
	}

	return dimms
}

// GetDetailedDiskInfo returns detailed disk information
func GetDetailedDiskInfo() []map[string]interface{} {
	var disks []map[string]interface{}

	if runtime.GOOS == "linux" {
		// Get disk info from lsblk
		if out, err := exec.Command("lsblk", "-b", "-d", "-o", "NAME,SIZE,TYPE,MODEL,ROTA").Output(); err == nil {
			disks = parseDiskInfo(string(out))
		}

		// Get filesystem info from df
		if out, err := exec.Command("df", "-h").Output(); err == nil {
			filesystems := parseFilesystemInfo(string(out))
			for i, disk := range disks {
				device := disk["name"].(string)
				for _, fs := range filesystems {
					if strings.Contains(fs["device"].(string), device) {
						disks[i]["mountpoint"] = fs["mountpoint"]
						disks[i]["filesystem"] = fs["filesystem"]
						disks[i]["used"] = fs["used"]
						disks[i]["available"] = fs["available"]
						disks[i]["usage_percent"] = fs["usage_percent"]
						break
					}
				}
			}
		}

		// Get SMART info if available
		for i, disk := range disks {
			device := disk["name"].(string)
			if smartInfo := getSMARTInfo(device); smartInfo != nil {
				disks[i]["smart"] = smartInfo
			}
		}
	}

	return disks
}

// GetDetailedGPUInfo returns detailed GPU information
func GetDetailedGPUInfo() []map[string]interface{} {
	var gpus []map[string]interface{}

	if runtime.GOOS == "linux" {
		// Check NVIDIA GPUs
		if out, err := exec.Command("nvidia-smi", "--query-gpu=index,name,memory.total,driver_version,temperature.gpu,power.draw", "--format=csv,noheader,nounits").Output(); err == nil {
			lines := strings.Split(strings.TrimSpace(string(out)), "\n")
			for i, line := range lines {
				fields := strings.Split(line, ", ")
				if len(fields) >= 5 {
					gpus = append(gpus, map[string]interface{}{
						"index":        i,
						"name":         strings.TrimSpace(fields[1]),
						"memory_total": strings.TrimSpace(fields[2]) + " MB",
						"driver":       strings.TrimSpace(fields[3]),
						"temperature":  strings.TrimSpace(fields[4]),
						"power":        strings.TrimSpace(fields[5]),
						"vendor":       "NVIDIA",
					})
				}
			}
		}

		// Check AMD GPUs
		if len(gpus) == 0 {
			if _, err := exec.Command("radeontop", "-d", "-l", "1").Output(); err == nil {
				// Parse AMD GPU info
				// Implementation depends on radeontop output format
			}
		}
	}

	return gpus
}

// GetDetailedNetworkInfo returns detailed network interface information
func GetDetailedNetworkInfo() []map[string]interface{} {
	var interfaces []map[string]interface{}

	if runtime.GOOS == "linux" {
		// Get interface info from ip command
		if out, err := exec.Command("ip", "-j", "link", "show").Output(); err == nil {
			var links []map[string]interface{}
			if err := json.Unmarshal(out, &links); err == nil {
				for _, link := range links {
					ifname := link["ifname"].(string)
					info := map[string]interface{}{
						"name":       ifname,
						"type":       link["link_type"],
						"state":      link["operstate"],
						"mtu":        link["mtu"],
						"mac":        link["address"],
						"addresses":  []string{},
						"tx_bytes":   0,
						"rx_bytes":   0,
						"tx_packets": 0,
						"rx_packets": 0,
					}

					// Get IP addresses
					if out, err := exec.Command("ip", "-j", "addr", "show", ifname).Output(); err == nil {
						var addrs []map[string]interface{}
						if err := json.Unmarshal(out, &addrs); err == nil {
							addresses := []string{}
							for _, addr := range addrs {
								for _, addrInfo := range addr["addr_info"].([]interface{}) {
									if info := addrInfo.(map[string]interface{}); info["family"].(string) == "inet" {
										addresses = append(addresses, fmt.Sprintf("%s/%d", info["local"], int(info["prefixlen"].(float64))))
									}
								}
							}
							info["addresses"] = addresses
						}
					}

					// Get statistics from /sys/class/net
					if tx, err := readFileInt64(fmt.Sprintf("/sys/class/net/%s/statistics/tx_bytes", ifname)); err == nil {
						info["tx_bytes"] = tx
					}
					if rx, err := readFileInt64(fmt.Sprintf("/sys/class/net/%s/statistics/rx_bytes", ifname)); err == nil {
						info["rx_bytes"] = rx
					}

					interfaces = append(interfaces, info)
				}
			}
		}
	}

	return interfaces
}

// Helper functions

func parseCPUInfo(cpuinfo string) map[string]interface{} {
	info := make(map[string]interface{})
	flags := []string{}

	lines := strings.Split(cpuinfo, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "vendor_id") {
			info["vendor"] = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "cpu family") {
			info["family"] = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "model\t") {
			info["model_number"] = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "stepping") {
			info["stepping"] = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "microcode") {
			info["microcode"] = strings.TrimSpace(strings.Split(line, ":")[1])
		}
		if strings.HasPrefix(line, "flags") {
			flagsStr := strings.TrimSpace(strings.Split(line, ":")[1])
			flags = strings.Fields(flagsStr)
		}
	}

	info["flags"] = flags
	return info
}

func parseMemoryDevices(dmidecode string) []map[string]interface{} {
	var dimms []map[string]interface{}
	
	// Simple parser for dmidecode output
	// This is a simplified version, full implementation would need more parsing
	re := regexp.MustCompile(`Size:\s+(\d+)\s+MB`)
	matches := re.FindAllStringSubmatch(dmidecode, -1)
	
	for _, match := range matches {
		if len(match) > 1 {
			size, _ := strconv.Atoi(match[1])
			if size > 0 {
				dimms = append(dimms, map[string]interface{}{
					"size": fmt.Sprintf("%d MB", size),
				})
			}
		}
	}
	
	return dimms
}

func parseDiskInfo(lsblk string) []map[string]interface{} {
	var disks []map[string]interface{}
	
	lines := strings.Split(lsblk, "\n")
	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 5 {
			disks = append(disks, map[string]interface{}{
				"name":  fields[0],
				"size":  fields[1],
				"type":  fields[2],
				"model": fields[3],
				"rota":  fields[4],
			})
		}
	}
	
	return disks
}

func parseFilesystemInfo(df string) []map[string]interface{} {
	var filesystems []map[string]interface{}
	
	lines := strings.Split(df, "\n")
	for i, line := range lines {
		if i == 0 {
			continue // skip header
		}
		
		fields := strings.Fields(line)
		if len(fields) >= 6 {
			filesystems = append(filesystems, map[string]interface{}{
				"filesystem":    fields[0],
				"mountpoint":    fields[5],
				"used":          fields[2],
				"available":     fields[3],
				"usage_percent": strings.TrimRight(fields[4], "%"),
			})
		}
	}
	
	return filesystems
}

func getSMARTInfo(device string) map[string]interface{} {
	smartInfo := make(map[string]interface{})
	
	// Check if device is /dev/sdX
	if !strings.HasPrefix(device, "/dev/") {
		device = "/dev/" + device
	}
	
	// Check if smartmontools is available
	if _, err := exec.LookPath("smartctl"); err != nil {
		return nil
	}
	
	// Get SMART attributes
	if out, err := exec.Command("smartctl", "-A", device).Output(); err == nil {
		// Parse SMART attributes
		// This is simplified, full implementation would parse all attributes
		smartInfo["available"] = true
		smartInfo["attributes"] = parseSMARTAtrributes(string(out))
	}
	
	return smartInfo
}

func parseSMARTAtrributes(smart string) map[string]interface{} {
	attributes := make(map[string]interface{})
	
	lines := strings.Split(smart, "\n")
	for _, line := range lines {
		fields := strings.Fields(line)
		if len(fields) >= 10 {
			attrName := fields[1]
			attrValue := fields[3]
			attributes[attrName] = attrValue
		}
	}
	
	return attributes
}

func readFileInt64(filepath string) (int64, error) {
	data, err := os.ReadFile(filepath)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(strings.TrimSpace(string(data)), 10, 64)
}

