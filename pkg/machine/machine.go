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
	"github.com/jaypipes/ghw/pkg/topology"
)

type Machine struct {
	Topology *topology.Info
}

func Discover(env *environ.Environ) (Machine, error) {
	info, err := topology.New()
	if err != nil {
		return Machine{}, err
	}
	env.Log.V(2).Info("detected machine", "topology", info)
	return Machine{
		Topology: info,
	}, nil
}
