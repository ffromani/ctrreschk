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

package cli

import (
	"encoding/json"
	"os"

	"github.com/jaypipes/ghw/pkg/topology"
	"github.com/spf13/cobra"

	"github.com/ffromani/ctrreschk/pkg/environ"
	"github.com/ffromani/ctrreschk/pkg/machine"
)

type InfoOptions struct{}

func NewInfoCommand(env *environ.Environ, opts *Options) *cobra.Command {
	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "show machine properties",
		RunE: func(cmd *cobra.Command, args []string) error {
			machine, err := machine.Discover(env)
			if err != nil {
				return err
			}
			// fixup ghw quirks
			machine.Topology.Architecture = topology.ARCHITECTURE_NUMA
			err = json.NewEncoder(os.Stdout).Encode(machine)
			if err != nil {
				return err
			}
			return MainLoop(opts)
		},
		Args: cobra.NoArgs,
	}
	return infoCmd
}
