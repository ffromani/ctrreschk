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

	apiv1 "github.com/ffromani/ctrreschk/api/v1"
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

type rMap struct {
	phys ridMap
	llc  ridMap
	numa ridMap
}

func (rm rMap) String() string {
	return fmt.Sprintf("<phys={%s} llc={%s} numa{%s}>", rm.phys.String(), rm.llc.String(), rm.numa.String())
}

func newRMap() rMap {
	return rMap{
		phys: make(ridMap),
		llc:  make(ridMap),
		numa: make(ridMap),
	}
}

func makeRMap(topo *topology.Info) rMap {
	res := newRMap()
	llcID := 0
	for _, node := range topo.Nodes {
		for _, core := range node.Cores {
			phys := res.phys[core.ID]
			phys = append(phys, core.LogicalProcessors...)
			res.phys[core.ID] = phys

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

func Check(env *environ.Environ, container resources.Resources, machine machine.Machine) (apiv1.Allocation, error) {
	rmap := makeRMap(machine.Topology)
	env.Log.V(2).Info("reverse mapping", "rmap", rmap)

	resp := apiv1.Allocation{}

	// TODO: SMT
	checkLLC(env, &resp, container.CPUs.Clone(), rmap)
	checkNUMA(env, &resp, container.CPUs.Clone(), rmap)

	return resp, nil
}

func checkLLC(env *environ.Environ, resp *apiv1.Allocation, cores cpuset.CPUSet, rmap rMap) {
	for llcID := range rmap.llc {
		if cores.Size() <= 0 {
			break
		}
		llcCores := rmap.llc.CPUSet(llcID)
		thisLLCSubset := cores.Intersection(llcCores)
		if resp.Aligned == nil {
			resp.Aligned = apiv1.NewAlignedInfo()
		}
		dets := resp.Aligned.LLC[llcID]
		dets.CPUs = thisLLCSubset.List()
		resp.Aligned.LLC[llcID] = dets

		cores = cores.Difference(thisLLCSubset)
	}

	resp.Alignment.LLC = cores.IsEmpty()
	if !resp.Alignment.LLC {
		if resp.Unaligned == nil {
			resp.Unaligned = &apiv1.UnalignedInfo{}
		}
		resp.Unaligned.LLC.CPUs = cores.List()
	}
}

func checkNUMA(env *environ.Environ, resp *apiv1.Allocation, cores cpuset.CPUSet, rmap rMap) {
	for numaID := range rmap.numa {
		if cores.Size() <= 0 {
			break
		}
		numaCores := rmap.numa.CPUSet(numaID)
		thisNUMASubset := cores.Intersection(numaCores)
		if resp.Aligned == nil {
			resp.Aligned = apiv1.NewAlignedInfo()
		}
		dets := resp.Aligned.NUMA[numaID]
		dets.CPUs = thisNUMASubset.List()
		resp.Aligned.NUMA[numaID] = dets

		cores = cores.Difference(thisNUMASubset)
	}

	resp.Alignment.NUMA = cores.IsEmpty()
	if !resp.Alignment.NUMA {
		if resp.Unaligned == nil {
			resp.Unaligned = &apiv1.UnalignedInfo{}
		}
		resp.Unaligned.NUMA.CPUs = cores.List()
	}
}
