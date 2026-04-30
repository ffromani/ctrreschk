// SPDX-License-Identifier: Apache-2.0

package numamaps

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/ffromani/ctrreschk/pkg/environ"
)

const (
	ProcSelfPath = "self"
	NumaMapsFile = "numa_maps"
)

var numaPageRe = regexp.MustCompile(`^N(\d+)=(\d+)$`)

type VMA struct {
	Address        uint64
	Policy         string
	FilePath       string
	NUMAPages      map[int]int64
	KernPageSizeKB int64
}

type NumaMaps struct {
	VMAs []VMA
}

func NumaMapsPath(env *environ.Environ) string {
	return filepath.Join(env.Root.Proc, ProcSelfPath, NumaMapsFile)
}

func Read(env *environ.Environ) (NumaMaps, error) {
	path := NumaMapsPath(env)
	env.Log.V(2).Info("reading numa_maps", "path", path)

	f, err := os.Open(path)
	if err != nil {
		return NumaMaps{}, err
	}
	defer f.Close()

	return parse(f)
}

func parse(f *os.File) (NumaMaps, error) {
	var result NumaMaps
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		vma, err := parseLine(line)
		if err != nil {
			return NumaMaps{}, fmt.Errorf("parsing line %q: %w", line, err)
		}
		result.VMAs = append(result.VMAs, vma)
	}
	if err := scanner.Err(); err != nil {
		return NumaMaps{}, err
	}
	return result, nil
}

func parseLine(line string) (VMA, error) {
	fields := strings.Fields(line)
	if len(fields) < 2 {
		return VMA{}, fmt.Errorf("too few fields")
	}

	addr, err := strconv.ParseUint(fields[0], 16, 64)
	if err != nil {
		return VMA{}, fmt.Errorf("invalid address %q: %w", fields[0], err)
	}

	vma := VMA{
		Address:        addr,
		Policy:         fields[1],
		NUMAPages:      make(map[int]int64),
		KernPageSizeKB: 4, // default
	}

	for _, field := range fields[2:] {
		if m := numaPageRe.FindStringSubmatch(field); m != nil {
			nodeID, _ := strconv.Atoi(m[1])
			pages, _ := strconv.ParseInt(m[2], 10, 64)
			vma.NUMAPages[nodeID] = pages
		} else if rest, ok := strings.CutPrefix(field, "file="); ok {
			vma.FilePath = rest
		} else if rest, ok := strings.CutPrefix(field, "kernelpagesize_kB="); ok {
			if v, err := strconv.ParseInt(rest, 10, 64); err == nil {
				vma.KernPageSizeKB = v
			}
		}
	}

	return vma, nil
}

func (nm NumaMaps) NUMANodes() []int {
	seen := make(map[int]struct{})
	for _, vma := range nm.VMAs {
		for nodeID := range vma.NUMAPages {
			seen[nodeID] = struct{}{}
		}
	}
	nodes := make([]int, 0, len(seen))
	for nodeID := range seen {
		nodes = append(nodes, nodeID)
	}
	sort.Ints(nodes)
	return nodes
}

func (nm NumaMaps) TotalPagesByNode() map[int]int64 {
	totals := make(map[int]int64)
	for _, vma := range nm.VMAs {
		for nodeID, pages := range vma.NUMAPages {
			totals[nodeID] += pages
		}
	}
	return totals
}

func (nm NumaMaps) TotalBytesByNode() map[int]int64 {
	totals := make(map[int]int64)
	for _, vma := range nm.VMAs {
		for nodeID, pages := range vma.NUMAPages {
			totals[nodeID] += pages * vma.KernPageSizeKB * 1024
		}
	}
	return totals
}
