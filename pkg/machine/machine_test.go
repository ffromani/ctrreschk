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
	"testing"

	"github.com/ffromani/ctrreschk/pkg/environ"
)

func TestDiscoverFundamentals(t *testing.T) {
	env := environ.New()
	got, err := Discover(env)
	if err != nil {
		t.Fatalf("discover error against real machine: %v", err)
	}
	if got.CPU == nil || got.Topology == nil {
		t.Fatalf("missing expected data in machine info CPU=%v Topology=%v", got.CPU, got.Topology)
	}
}
