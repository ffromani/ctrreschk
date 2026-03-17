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

	"github.com/go-logr/logr"
	"github.com/spf13/cobra"

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
			err := runScan(env.Log, sysfs)
			if err != nil {
				return err
			}
			return MainLoop(opts)
		},
		Args: cobra.NoArgs,
	}
}

func runScan(lh logr.Logger, sysfs device.SysFS) error {
	domains, err := device.PCIEDomainsFromFS(lh, sysfs)
	if err != nil {
		return fmt.Errorf("failed to scan the PCIE domains: %w", err)
	}
	lh.V(4).Info("found PCIE domains", "count", len(domains))
	for _, dom := range domains {
		lh.Info("PCIE domain", dom.Loggable()...)
	}

	onlineCPUs, err := device.OnlineCPUs(lh, sysfs)
	if err != nil {
		return fmt.Errorf("failed to get the online CPUs: %w", err)
	}

	orphans := device.FindOrphanedCPUs(domains, onlineCPUs)
	lh.V(2).Info("found orphaned CPUs", "count", orphans.Size())
	return nil
}
