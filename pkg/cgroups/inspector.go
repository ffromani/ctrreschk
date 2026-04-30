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

package cgroups

import (
	"os"
	"path/filepath"
	"strings"

	"k8s.io/utils/cpuset"

	"github.com/ffromani/ctrreschk/pkg/environ"
)

const (
	CgroupPath = "fs/cgroup"
	CpusetFile = "cpuset.cpus.effective"
	MemsetFile = "cpuset.mems.effective"
)

func CpusetPath(env *environ.Environ) string {
	return filepath.Join(env.Root.Sys, CgroupPath, CpusetFile)
}

func MemsetPath(env *environ.Environ) string {
	return filepath.Join(env.Root.Sys, CgroupPath, MemsetFile)
}

func Cpuset(env *environ.Environ) (cpuset.CPUSet, error) {
	cpusetPath := CpusetPath(env)
	env.Log.V(2).Info("reading cpuset", "path", cpusetPath)
	data, err := os.ReadFile(cpusetPath)
	if err != nil {
		env.Log.V(1).Info("failed to read cpuset", "path", cpusetPath, "error", err)
		return cpuset.New(), err
	}
	cpus, err := cpuset.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		env.Log.V(1).Info("failed to parse cpuset", "path", cpusetPath, "error", err)
		return cpuset.New(), err
	}
	env.Log.V(2).Info("parsed cpuset", "path", cpusetPath, "cpus", cpus.String())
	return cpus, nil
}

// Memset reads the NUMA memory nodes allowed for the container from cpuset.mems.effective.
// The format is the same range-list notation used for cpuset.cpus.effective.
func Memset(env *environ.Environ) (cpuset.CPUSet, error) {
	memsetPath := MemsetPath(env)
	env.Log.V(2).Info("reading memset", "path", memsetPath)
	data, err := os.ReadFile(memsetPath)
	if err != nil {
		env.Log.V(1).Info("failed to read memset", "path", memsetPath, "error", err)
		return cpuset.New(), err
	}
	mems, err := cpuset.Parse(strings.TrimSpace(string(data)))
	if err != nil {
		env.Log.V(1).Info("failed to parse memset", "path", memsetPath, "error", err)
		return cpuset.New(), err
	}
	env.Log.V(2).Info("parsed memset", "path", memsetPath, "mems", mems.String())
	return mems, nil
}
