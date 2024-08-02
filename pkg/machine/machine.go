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

package machine

import (
	"github.com/ffromani/ctrreschk/pkg/environ"
	"github.com/jaypipes/ghw/pkg/cpu"
	"github.com/jaypipes/ghw/pkg/topology"
)

type Machine struct {
	CPU      *cpu.Info      `json:"cpu"`
	Topology *topology.Info `json:"topology"`
}

func Discover(env *environ.Environ) (Machine, error) {
	mc := Machine{}

	cpu, err := cpu.New()
	if err != nil {
		return mc, err
	}
	mc.CPU = cpu
	env.Log.V(2).Info("detected machine", "CPU", cpu)

	topo, err := topology.New()
	if err != nil {
		return mc, err
	}
	mc.Topology = topo
	env.Log.V(2).Info("detected machine", "topology", topo)

	return mc, nil
}
