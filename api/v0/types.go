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

package v0

type ContainerResourcesDetails struct {
	// CPUs are identified by their virtual cpu ID
	CPUs []int `json:"cpus,omitempty"`
	// NUMANodes lists the NUMA node IDs involved in unaligned memory allocation
	NUMANodes []int `json:"numaNodes,omitempty"`
	// MemoryMiB is the usable memory in MiB available on the relevant NUMA node(s)
	MemoryMiB int64 `json:"memoryMiB,omitempty"`
	// MemoryPercent is the percentage of total machine memory on the relevant NUMA node(s)
	MemoryPercent float64 `json:"memoryPercent,omitempty"`
	// Hugepages are anonymous
	Hugepages2Mi int `json:"hugepages2Mi,omitempty"`
	Hugepages1Gi int `json:"hugepages1Gi,omitempty"`
	// Devices are identified by name
	Devices []string `json:"devices,omitempty"`
}

type AlignedInfo struct {
	// vcoreid -> resources
	SMT map[int]ContainerResourcesDetails `json:"smt,omitempty"`
	// llcid -> resources
	LLC map[int]ContainerResourcesDetails `json:"llc,omitempty"`
	// numacellid -> resources
	NUMA map[int]ContainerResourcesDetails `json:"numa,omitempty"`
	// numacellid -> resources (memory NUMA nodes matching CPU NUMA nodes)
	Memory map[int]ContainerResourcesDetails `json:"memory,omitempty"`
}

type UnalignedInfo struct {
	SMT    ContainerResourcesDetails `json:"smt,omitempty"`
	LLC    ContainerResourcesDetails `json:"llc,omitempty"`
	NUMA   ContainerResourcesDetails `json:"numa,omitempty"`
	Memory ContainerResourcesDetails `json:"memory,omitempty"`
}

type Alignment struct {
	SMT    bool `json:"smt"`
	LLC    bool `json:"llc"`
	NUMA   bool `json:"numa"`
	Memory bool `json:"memory"`
}

type Allocation struct {
	Alignment Alignment      `json:"alignment"`
	Aligned   *AlignedInfo   `json:"aligned,omitempty"`
	Unaligned *UnalignedInfo `json:"unaligned,omitempty"`
}
