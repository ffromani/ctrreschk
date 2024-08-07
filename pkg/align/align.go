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
	"strings"

	"k8s.io/utils/cpuset"

	"github.com/jaypipes/ghw/pkg/topology"

	apiv0 "github.com/ffromani/ctrreschk/api/v0"
	"github.com/ffromani/ctrreschk/pkg/environ"
	"github.com/ffromani/ctrreschk/pkg/machine"
	"github.com/ffromani/ctrreschk/pkg/resources"
)

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
	cpuLog2Phy map[int]int
	cpuPhy2Log ridMap
	llc        ridMap
	numa       ridMap
}

func (rm rMap) String() string {
	return fmt.Sprintf("<phys=%{s} llc={%s} numa{%s}>", rm.cpuPhy2Log.String(), rm.llc.String(), rm.numa.String())
}

func newRMap() rMap {
	return rMap{
		cpuLog2Phy: make(map[int]int),
		cpuPhy2Log: make(ridMap),
		llc:        make(ridMap),
		numa:       make(ridMap),
	}
}

func makeRMap(topo *topology.Info) rMap {
	res := newRMap()
	llcID := 0
	for _, node := range topo.Nodes {
		for _, core := range node.Cores {
			phys := res.cpuPhy2Log[core.ID]
			phys = append(phys, core.LogicalProcessors...)
			res.cpuPhy2Log[core.ID] = phys
			for _, lid := range core.LogicalProcessors {
				res.cpuLog2Phy[lid] = core.ID
			}

			numa := res.numa[node.ID]
			numa = append(numa, core.LogicalProcessors...)
			res.numa[node.ID] = numa
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

			llcID += 1
		}
	}

	return res
}

func Check(env *environ.Environ, container resources.Resources, machine machine.Machine) (apiv0.Allocation, error) {
	rmap := makeRMap(machine.Topology)
	env.Log.V(2).Info("reverse mapping", "rmap", rmap)

	resp := apiv0.Allocation{}

	checkSMT(env, &resp, container.CPUs.Clone(), rmap)
	checkLLC(env, &resp, container.CPUs.Clone(), rmap)
	checkNUMA(env, &resp, container.CPUs.Clone(), rmap)

	return resp, nil
}

func checkSMT(env *environ.Environ, resp *apiv0.Allocation, cores cpuset.CPUSet, rmap rMap) {
	var coreList []int
	for _, coreID := range cores.UnsortedList() {
		phy := rmap.cpuLog2Phy[coreID]
		coreList = append(coreList, rmap.cpuPhy2Log[phy]...)
	}
	computedCores := cpuset.New(coreList...)

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
		dets.CPUs = thisLLCSubset.List()
		resp.Aligned.LLC[llcID] = dets

		cores = cores.Difference(thisLLCSubset)
	}

	resp.Alignment.LLC = cores.IsEmpty()
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
		dets.CPUs = thisNUMASubset.List()
		resp.Aligned.NUMA[numaID] = dets

		cores = cores.Difference(thisNUMASubset)
	}

	resp.Alignment.NUMA = cores.IsEmpty()
	if !resp.Alignment.NUMA {
		if resp.Unaligned == nil {
			resp.Unaligned = &apiv0.UnalignedInfo{}
		}
		resp.Unaligned.NUMA.CPUs = cores.List()
	}
}
