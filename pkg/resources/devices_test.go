// SPDX-License-Identifier: Apache-2.0

package resources

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/ffromani/ctrreschk/pkg/environ"
)

func TestDiscoverDevicesFromEnv(t *testing.T) {
	testCases := []struct {
		name     string
		environ  []string
		prefixes []string
		sysfs    map[string]string // pciAddress -> numa_node content
		expected []DeviceInfo
	}{
		{
			name:     "no matching prefixes",
			environ:  []string{"FOO_BAR=0000:05:10.2"},
			prefixes: []string{"SRIOVNETWORK_VF_"},
			expected: nil,
		},
		{
			name:     "single device",
			environ:  []string{"SRIOVNETWORK_VF_DEVICE_0000_05_10_2=0000:05:10.2"},
			prefixes: []string{"SRIOVNETWORK_VF_"},
			sysfs:    map[string]string{"0000:05:10.2": "0\n"},
			expected: []DeviceInfo{
				{EnvVar: "SRIOVNETWORK_VF_DEVICE_0000_05_10_2", PCIAddress: "0000:05:10.2", NUMANode: 0},
			},
		},
		{
			name:     "multiple devices comma separated",
			environ:  []string{"PCIDEVICE_IO_NICS=0000:86:00.0,0000:87:00.0"},
			prefixes: []string{"PCIDEVICE_"},
			sysfs:    map[string]string{"0000:86:00.0": "0\n", "0000:87:00.0": "1\n"},
			expected: []DeviceInfo{
				{EnvVar: "PCIDEVICE_IO_NICS", PCIAddress: "0000:86:00.0", NUMANode: 0},
				{EnvVar: "PCIDEVICE_IO_NICS", PCIAddress: "0000:87:00.0", NUMANode: 1},
			},
		},
		{
			name:     "missing sysfs entry returns -1",
			environ:  []string{"SRIOVNETWORK_VF_DEV=0000:99:00.0"},
			prefixes: []string{"SRIOVNETWORK_VF_"},
			sysfs:    map[string]string{},
			expected: []DeviceInfo{
				{EnvVar: "SRIOVNETWORK_VF_DEV", PCIAddress: "0000:99:00.0", NUMANode: -1},
			},
		},
		{
			name:     "numa node -1 from sysfs",
			environ:  []string{"SRIOVNETWORK_VF_DEV=0000:05:10.2"},
			prefixes: []string{"SRIOVNETWORK_VF_"},
			sysfs:    map[string]string{"0000:05:10.2": "-1\n"},
			expected: []DeviceInfo{
				{EnvVar: "SRIOVNETWORK_VF_DEV", PCIAddress: "0000:05:10.2", NUMANode: -1},
			},
		},
		{
			name:     "multiple prefixes",
			environ:  []string{"SRIOVNETWORK_VF_A=0000:05:10.2", "PCIDEVICE_B=0000:06:00.0", "OTHER=ignored"},
			prefixes: []string{"SRIOVNETWORK_VF_", "PCIDEVICE_"},
			sysfs:    map[string]string{"0000:05:10.2": "0\n", "0000:06:00.0": "0\n"},
			expected: []DeviceInfo{
				{EnvVar: "SRIOVNETWORK_VF_A", PCIAddress: "0000:05:10.2", NUMANode: 0},
				{EnvVar: "PCIDEVICE_B", PCIAddress: "0000:06:00.0", NUMANode: 0},
			},
		},
		{
			name:     "empty value skipped",
			environ:  []string{"SRIOVNETWORK_VF_DEV="},
			prefixes: []string{"SRIOVNETWORK_VF_"},
			expected: nil,
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			env := &environ.Environ{
				Root: environ.FS{Sys: tmpDir},
				Log:  environ.DefaultLog(),
			}

			for addr, content := range tt.sysfs {
				devDir := filepath.Join(tmpDir, "bus", "pci", "devices", addr)
				if err := os.MkdirAll(devDir, os.ModePerm); err != nil {
					t.Fatalf("cannot create sysfs dir: %v", err)
				}
				if err := os.WriteFile(filepath.Join(devDir, "numa_node"), []byte(content), 0o644); err != nil {
					t.Fatalf("cannot write numa_node: %v", err)
				}
			}

			got := DiscoverDevicesFromEnv(env, tt.environ, tt.prefixes)

			if len(got) != len(tt.expected) {
				t.Fatalf("expected %d devices, got %d: %+v", len(tt.expected), len(got), got)
			}
			for i, exp := range tt.expected {
				if got[i].EnvVar != exp.EnvVar {
					t.Errorf("device[%d] EnvVar: expected %q, got %q", i, exp.EnvVar, got[i].EnvVar)
				}
				if got[i].PCIAddress != exp.PCIAddress {
					t.Errorf("device[%d] PCIAddress: expected %q, got %q", i, exp.PCIAddress, got[i].PCIAddress)
				}
				if got[i].NUMANode != exp.NUMANode {
					t.Errorf("device[%d] NUMANode: expected %d, got %d", i, exp.NUMANode, got[i].NUMANode)
				}
			}
		})
	}
}
