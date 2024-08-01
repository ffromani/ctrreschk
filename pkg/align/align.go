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
	apiv1 "github.com/ffromani/ctrreschk/api/v1"
	"github.com/ffromani/ctrreschk/pkg/machine"
	"github.com/ffromani/ctrreschk/pkg/resources"
)

func Check(container resources.Resources, machine machine.Machine) (apiv1.Allocation, error) {
	return apiv1.Allocation{}, nil
}
