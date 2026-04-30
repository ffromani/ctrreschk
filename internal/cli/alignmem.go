// SPDX-License-Identifier: Apache-2.0

package cli

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	"k8s.io/utils/cpuset"

	apiv0 "github.com/ffromani/ctrreschk/api/v0"
	"github.com/ffromani/ctrreschk/pkg/cgroups"
	"github.com/ffromani/ctrreschk/pkg/environ"
	"github.com/ffromani/ctrreschk/pkg/machine"
	"github.com/ffromani/ctrreschk/pkg/numamaps"
)

func NewAlignMemCommand(env *environ.Environ, opts *Options) *cobra.Command {
	alignMemCmd := &cobra.Command{
		Use:   "alignmem",
		Short: "verify actual memory NUMA placement via numa_maps",
		RunE: func(cmd *cobra.Command, args []string) error {
			cpus, err := cgroups.Cpuset(env)
			if err != nil {
				return err
			}

			mach, err := machine.Discover(env)
			if err != nil {
				return err
			}

			nm, err := numamaps.Read(env)
			if err != nil {
				return err
			}

			cpuNUMANodes := cpuNUMANodesFromTopology(cpus, mach)
			env.Log.V(2).Info("alignmem", "cpuNUMANodes", cpuNUMANodes.String())

			result := buildNUMAMapsInfo(nm, cpuNUMANodes)

			err = json.NewEncoder(os.Stdout).Encode(result)
			if err != nil {
				return err
			}
			return MainLoop(opts)
		},
		Args: cobra.NoArgs,
	}

	alignMemCmd.PersistentFlags().StringVarP(&env.DataPath, "machinedata", "M", "", "read fake machine data from path, don't read real data from the system")

	return alignMemCmd
}

func cpuNUMANodesFromTopology(cpus cpuset.CPUSet, mach machine.Machine) cpuset.CPUSet {
	result := cpuset.New()
	for _, node := range mach.Topology.Nodes {
		for _, core := range node.Cores {
			coreCPUs := cpuset.New(core.LogicalProcessors...)
			if !cpus.Intersection(coreCPUs).IsEmpty() {
				result = result.Union(cpuset.New(node.ID))
				break
			}
		}
	}
	return result
}

func buildNUMAMapsInfo(nm numamaps.NumaMaps, cpuNUMANodes cpuset.CPUSet) apiv0.NUMAMapsInfo {
	pagesByNode := nm.TotalPagesByNode()
	bytesByNode := nm.TotalBytesByNode()

	info := apiv0.NUMAMapsInfo{
		Nodes: make(map[int]apiv0.NUMAMapsNodeInfo),
	}

	for nodeID, pages := range pagesByNode {
		info.Nodes[nodeID] = apiv0.NUMAMapsNodeInfo{
			Pages:   pages,
			SizeKiB: bytesByNode[nodeID] / 1024,
		}
		if cpuNUMANodes.Contains(nodeID) {
			info.LocalPages += pages
		} else {
			info.RemotePages += pages
		}
	}

	info.Local = info.RemotePages == 0 && info.LocalPages > 0

	return info
}
