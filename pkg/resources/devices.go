// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ffromani/ctrreschk/pkg/environ"
)

func DiscoverDevicesFromEnv(env *environ.Environ, osEnviron []string, prefixes []string) []DeviceInfo {
	var devices []DeviceInfo
	for _, entry := range osEnviron {
		name, value, ok := strings.Cut(entry, "=")
		if !ok || value == "" {
			continue
		}
		if !matchesAnyPrefix(name, prefixes) {
			continue
		}
		for _, addr := range strings.Split(value, ",") {
			addr = strings.TrimSpace(addr)
			if addr == "" {
				continue
			}
			numaNode := readDeviceNUMANode(env, addr)
			devices = append(devices, DeviceInfo{
				EnvVar:     name,
				PCIAddress: addr,
				NUMANode:   numaNode,
			})
			env.Log.V(2).Info("discovered device", "envVar", name, "pciAddress", addr, "numaNode", numaNode)
		}
	}
	return devices
}

func matchesAnyPrefix(name string, prefixes []string) bool {
	for _, prefix := range prefixes {
		if strings.HasPrefix(name, prefix) {
			return true
		}
	}
	return false
}

func readDeviceNUMANode(env *environ.Environ, pciAddress string) int {
	numaPath := filepath.Join(env.Root.Sys, "bus", "pci", "devices", pciAddress, "numa_node")
	data, err := os.ReadFile(numaPath)
	if err != nil {
		env.Log.V(1).Info("cannot read device NUMA node, skipping", "pciAddress", pciAddress, "error", err)
		return -1
	}
	node, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		env.Log.V(1).Info("cannot parse device NUMA node, skipping", "pciAddress", pciAddress, "error", err)
		return -1
	}
	return node
}
