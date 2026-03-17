/*
 * Copyright 2026 Red Hat, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/klog/v2"

	"github.com/ffromani/ctrreschk/pkg/device"
	"github.com/ffromani/ctrreschk/pkg/environ"
)

type PCIEScanOptions struct{}

func NewPCIEScanCommand(env *environ.Environ, opts *Options) *cobra.Command {
	return &cobra.Command{
		Use:   "pciescan",
		Short: "show pcieroot data",
		RunE: func(cmd *cobra.Command, args []string) error {
			sysfs := os.DirFS(env.Root.Sys).(device.SysFS)
			err := runScan(sysfs)
			if err != nil {
				return err
			}
			return MainLoop(opts)
		},
		Args: cobra.NoArgs,
	}
}

func runScan(sysfs device.SysFS) error {
	domains, err := device.PCIEDomainsFromFS(sysfs)
	if err != nil {
		return fmt.Errorf("failed to scan the PCIE domains: %w", err)
	}

	klog.V(3).Infof("found %d PCIE domains", len(domains))
	for _, dom := range domains {
		klog.Infof("PCIE domain: %s", dom.String())
	}

	onlineCPUs, err := device.OnlineCPUs(sysfs)
	if err != nil {
		return fmt.Errorf("failed to get the online CPUs: %w", err)
	}

	orphans := device.FindOrphanedCPUs(domains, onlineCPUs)
	klog.V(3).Infof("found %d orphaned CPUs", orphans.Size())
	return nil
}
