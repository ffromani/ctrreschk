/*
 * Copyright 2024 Red Hat, Inc.
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

package align

import (
	"fmt"
	"slices"
	"strings"

	"k8s.io/utils/cpuset"

	"github.com/jaypipes/ghw/pkg/topology"

	apiv0 "github.com/ffromani/ctrreschk/api/v0"
	"github.com/ffromani/ctrreschk/pkg/environ"
	"github.com/ffromani/ctrreschk/pkg/machine"
	"github.com/ffromani/ctrreschk/pkg/resources"
)

func Check(env *environ.Environ, container resources.Resources, machine machine.Machine) (apiv0.Allocation, error) {
	rmap := makeRMap(env, machine.Topology)
	env.Log.V(2).Info("reverse mapping", "rmap", rmap)

	if !container.MEMs.IsEmpty() && rmap.totalMemory <= 0 {
		return apiv0.Allocation{}, fmt.Errorf("memory nodes assigned but no memory information available from the machine topology")
	}

	resp := apiv0.Allocation{}

	checkSMT(env, &resp, container.CPUs.Clone(), rmap)
	checkLLC(env, &resp, container.CPUs.Clone(), rmap)
	checkNUMA(env, &resp, container.CPUs.Clone(), rmap)
	checkMemory(env, &resp, container.CPUs.Clone(), container.MEMs.Clone(), rmap)

	return resp, nil
}

func checkSMT(env *environ.Environ, resp *apiv0.Allocation, cores cpuset.CPUSet, rmap rMap) {
	var coreList []int
	for _, coreID := range cores.UnsortedList() {
		phy := rmap.cpuLog2Phy[coreID]
		altCores := rmap.cpuPhy2Log[phy]
		coreList = append(coreList, altCores...)
		env.Log.V(2).Info("check SMT alignment", "vcpuID", coreID, "pcpuID", phy, "altCores", altCores)
	}
	computedCores := cpuset.New(coreList...)

	env.Log.V(2).Info("check SMT alignment", "cores", cores.String(), "computedCores", computedCores.String())

	resp.Alignment.SMT = cores.Equals(computedCores)
	if !resp.Alignment.SMT {
		if resp.Unaligned == nil {
			resp.Unaligned = &apiv0.UnalignedInfo{}
		}
		// by construction, computedCores is always a superset of cores
		resp.Unaligned.SMT.CPUs = computedCores.Difference(cores).List()
	}
}

func checkLLC(env *environ.Environ, resp *apiv0.Allocation, cores cpuset.CPUSet, rmap rMap) {
	for llcID := range rmap.llc {
		if cores.Size() <= 0 {
			break
		}
		llcCores := rmap.llc.CPUSet(llcID)
		thisLLCSubset := cores.Intersection(llcCores)
		if resp.Aligned == nil {
			resp.Aligned = apiv0.NewAlignedInfo()
		}
		dets := resp.Aligned.LLC[llcID]
		if cpus := thisLLCSubset.List(); len(cpus) > 0 {
			dets.CPUs = cpus
			resp.Aligned.LLC[llcID] = dets
		}

		cores = cores.Difference(thisLLCSubset)
	}

	resp.Alignment.LLC = cores.IsEmpty() && (len(resp.Aligned.LLC) == 1)
	if !resp.Alignment.LLC {
		if resp.Unaligned == nil {
			resp.Unaligned = &apiv0.UnalignedInfo{}
		}
		resp.Unaligned.LLC.CPUs = cores.List()
	}
}

func checkNUMA(env *environ.Environ, resp *apiv0.Allocation, cores cpuset.CPUSet, rmap rMap) {
	for numaID := range rmap.numa {
		if cores.Size() <= 0 {
			break
		}
		numaCores := rmap.numa.CPUSet(numaID)
		thisNUMASubset := cores.Intersection(numaCores)
		if resp.Aligned == nil {
			resp.Aligned = apiv0.NewAlignedInfo()
		}
		dets := resp.Aligned.NUMA[numaID]
		if cpus := thisNUMASubset.List(); len(cpus) > 0 {
			dets.CPUs = cpus
			resp.Aligned.NUMA[numaID] = dets
		}

		cores = cores.Difference(thisNUMASubset)
	}

	resp.Alignment.NUMA = cores.IsEmpty() && (len(resp.Aligned.NUMA) == 1)
	if !resp.Alignment.NUMA {
		if resp.Unaligned == nil {
			resp.Unaligned = &apiv0.UnalignedInfo{}
		}
		resp.Unaligned.NUMA.CPUs = cores.List()
	}
}

const bytesPerMiB = 1024 * 1024

func checkMemory(env *environ.Environ, resp *apiv0.Allocation, cpus cpuset.CPUSet, mems cpuset.CPUSet, rmap rMap) {
	if mems.IsEmpty() {
		env.Log.V(1).Info("no memory node info available, skipping memory alignment check")
		return
	}

	// determine which NUMA nodes the container CPUs belong to
	cpuNUMANodes := cpuset.New()
	for numaID := range rmap.numa {
		numaCPUs := rmap.numa.CPUSet(numaID)
		if !cpus.Intersection(numaCPUs).IsEmpty() {
			cpuNUMANodes = cpuNUMANodes.Union(cpuset.New(numaID))
		}
	}

	env.Log.V(2).Info("check memory alignment", "cpuNUMANodes", cpuNUMANodes.String(), "mems", mems.String(), "totalMemory", rmap.totalMemory)

	resp.Alignment.Memory = cpuNUMANodes.Equals(mems)

	// report the common NUMA nodes (have both CPUs and memory)
	commonNodes := cpuNUMANodes.Intersection(mems)
	for _, numaID := range commonNodes.List() {
		if resp.Aligned == nil {
			resp.Aligned = apiv0.NewAlignedInfo()
		}
		dets := resp.Aligned.Memory[numaID]
		numaCPUs := rmap.numa.CPUSet(numaID)
		if containerCPUs := cpus.Intersection(numaCPUs).List(); len(containerCPUs) > 0 {
			dets.CPUs = containerCPUs
		}
		dets.MemoryMiB = rmap.numaMemory[numaID] / bytesPerMiB
		dets.MemoryPercent = float64(rmap.numaMemory[numaID]) / float64(rmap.totalMemory) * 100.0
		resp.Aligned.Memory[numaID] = dets
	}

	if !resp.Alignment.Memory {
		if resp.Unaligned == nil {
			resp.Unaligned = &apiv0.UnalignedInfo{}
		}
		resp.Unaligned.Memory.NUMANodes = mems.List()
		// NUMA nodes in mems but without container CPUs
		extraMem := mems.Difference(cpuNUMANodes)
		if !extraMem.IsEmpty() {
			var extraBytes int64
			for _, numaID := range extraMem.List() {
				extraBytes += rmap.numaMemory[numaID]
			}
			resp.Unaligned.Memory.MemoryMiB = extraBytes / bytesPerMiB
			resp.Unaligned.Memory.MemoryPercent = float64(extraBytes) / float64(rmap.totalMemory) * 100.0
		}
		// container CPUs on NUMA nodes not in mems (no local memory)
		extraCPUNUMA := cpuNUMANodes.Difference(mems)
		for _, numaID := range extraCPUNUMA.List() {
			numaCPUs := rmap.numa.CPUSet(numaID)
			resp.Unaligned.Memory.CPUs = append(resp.Unaligned.Memory.CPUs, cpus.Intersection(numaCPUs).List()...)
		}
	}
}

// Reverse ID MAP (PhysicalID|LLCID|NUMAID) -> LogicalIDs
type ridMap map[int][]int

func (rm ridMap) CPUSet(id int) cpuset.CPUSet {
	return cpuset.New(rm[id]...)
}

func (rm ridMap) String() string {
	var sb strings.Builder
	for key, values := range rm {
		fmt.Fprintf(&sb, " %02d->[%s]", key, cpuset.New(values...).String())
	}
	return sb.String()[1:]
}

// Resource MAPping
type rMap struct {
	cpuLog2Phy  map[int]int
	cpuPhy2Log  ridMap
	llc         ridMap
	numa        ridMap
	numaMemory  map[int]int64 // numaID -> usable bytes
	totalMemory int64         // sum of all NUMA nodes usable bytes
}

func (rm rMap) String() string {
	return fmt.Sprintf("<phys={%s} llc={%s} numa{%s}>", rm.cpuPhy2Log.String(), rm.llc.String(), rm.numa.String())
}

func newRMap() rMap {
	return rMap{
		cpuLog2Phy: make(map[int]int),
		cpuPhy2Log: make(ridMap),
		llc:        make(ridMap),
		numa:       make(ridMap),
		numaMemory: make(map[int]int64),
	}
}

func makeRMap(env *environ.Environ, topo *topology.Info) rMap {
	res := newRMap()
	llcID := 0
	for _, node := range topo.Nodes {
		for _, core := range node.Cores {
			coreID, _ := getUniqueCoreID(core.LogicalProcessors)
			phys := res.cpuPhy2Log[coreID]
			phys = append(phys, core.LogicalProcessors...)
			res.cpuPhy2Log[coreID] = phys
			env.Log.V(2).Info("rmap cpus core -> vpcus", "coreID", coreID, "vcpuIDs", core.LogicalProcessors, "cores", phys)

			for _, lid := range core.LogicalProcessors {
				res.cpuLog2Phy[lid] = coreID
				env.Log.V(2).Info("rmap cpus vcpu -> core", "vcpuID", lid, "coreID", coreID)
			}

			numa := res.numa[node.ID]
			numa = append(numa, core.LogicalProcessors...)
			res.numa[node.ID] = numa
			env.Log.V(4).Info("rmap numa -> vcpus", "numaID", node.ID, "vcpus", numa)
		}
		if node.Memory != nil {
			res.numaMemory[node.ID] = node.Memory.TotalUsableBytes
			res.totalMemory += node.Memory.TotalUsableBytes
			env.Log.V(4).Info("rmap numa -> memory", "numaID", node.ID, "usableBytes", node.Memory.TotalUsableBytes)
		}

		// TODO: yes, we assume LLC=L3.
		for _, cache := range node.Caches {
			if cache.Level < 3 {
				continue
			}

			llc := res.llc[llcID]
			for _, id := range cache.LogicalProcessors {
				llc = append(llc, int(id))
			}
			res.llc[llcID] = llc
			env.Log.V(4).Info("rmap LLC llcid -> vpcuID", "llcID", llcID, "vcpuIDs", llc)

			llcID += 1
		}
	}

	return res
}

// getUniqueCoreID computes coreId as the lowest cpuID
// for a given Threads []int slice. This will assure that coreID's are
// platform unique (opposite to what cAdvisor reports)
func getUniqueCoreID(threads []int) (coreID int, err error) {
	if len(threads) == 0 {
		return 0, fmt.Errorf("no cpus provided")
	}

	if len(threads) != cpuset.New(threads...).Size() {
		return 0, fmt.Errorf("cpus provided are not unique")
	}

	return slices.Min(threads), nil
}
