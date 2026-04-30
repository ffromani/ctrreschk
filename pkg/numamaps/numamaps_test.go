// SPDX-License-Identifier: Apache-2.0

package numamaps

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/ffromani/ctrreschk/pkg/environ"
)

func TestRead(t *testing.T) {
	testCases := []struct {
		name        string
		path        string
		content     string
		expectedErr bool
		checkFn     func(t *testing.T, nm NumaMaps)
	}{
		{
			name:        "non-existent path",
			path:        "/this/path/does/not/exist",
			expectedErr: true,
		},
		{
			name: "single node all local",
			content: "55dc822f1000 default file=/usr/bin/ctrreschk mapped=10 active=5 N0=10 kernelpagesize_kB=4\n" +
				"7fffef2ff000 default stack anon=3 dirty=3 active=1 N0=3 kernelpagesize_kB=4\n",
			checkFn: func(t *testing.T, nm NumaMaps) {
				if len(nm.VMAs) != 2 {
					t.Fatalf("expected 2 VMAs, got %d", len(nm.VMAs))
				}
				nodes := nm.NUMANodes()
				if !reflect.DeepEqual(nodes, []int{0}) {
					t.Fatalf("expected nodes [0], got %v", nodes)
				}
				pages := nm.TotalPagesByNode()
				if pages[0] != 13 {
					t.Fatalf("expected 13 pages on N0, got %d", pages[0])
				}
				bytes := nm.TotalBytesByNode()
				if bytes[0] != 13*4*1024 {
					t.Fatalf("expected %d bytes on N0, got %d", 13*4*1024, bytes[0])
				}
			},
		},
		{
			name: "two nodes split pages",
			content: "55dc822f1000 default file=/usr/bin/sleep mapped=2 active=0 N1=2 kernelpagesize_kB=4\n" +
				"55dc822fb000 default file=/usr/bin/sleep anon=1 dirty=1 active=0 N0=1 kernelpagesize_kB=4\n" +
				"7fb2a62c7000 default file=/usr/lib/x86_64-linux-gnu/libc.so.6 mapped=40 active=0 N1=40 kernelpagesize_kB=4\n" +
				"7fffef2ff000 default stack anon=3 dirty=3 active=1 N0=3 kernelpagesize_kB=4\n",
			checkFn: func(t *testing.T, nm NumaMaps) {
				if len(nm.VMAs) != 4 {
					t.Fatalf("expected 4 VMAs, got %d", len(nm.VMAs))
				}
				nodes := nm.NUMANodes()
				if !reflect.DeepEqual(nodes, []int{0, 1}) {
					t.Fatalf("expected nodes [0 1], got %v", nodes)
				}
				pages := nm.TotalPagesByNode()
				if pages[0] != 4 {
					t.Fatalf("expected 4 pages on N0, got %d", pages[0])
				}
				if pages[1] != 42 {
					t.Fatalf("expected 42 pages on N1, got %d", pages[1])
				}
			},
		},
		{
			name: "empty VMAs no pages",
			content: "7fb2a64be000 default\n" +
				"7fb2a64c2000 default\n",
			checkFn: func(t *testing.T, nm NumaMaps) {
				if len(nm.VMAs) != 2 {
					t.Fatalf("expected 2 VMAs, got %d", len(nm.VMAs))
				}
				nodes := nm.NUMANodes()
				if len(nodes) != 0 {
					t.Fatalf("expected no nodes, got %v", nodes)
				}
			},
		},
		{
			name:    "mixed page sizes",
			content: "400000 default anon=10 N0=10 kernelpagesize_kB=4\n500000 default anon=5 N0=5 kernelpagesize_kB=2048\n",
			checkFn: func(t *testing.T, nm NumaMaps) {
				pages := nm.TotalPagesByNode()
				if pages[0] != 15 {
					t.Fatalf("expected 15 pages on N0, got %d", pages[0])
				}
				bytes := nm.TotalBytesByNode()
				expectedBytes := int64(10*4*1024 + 5*2048*1024)
				if bytes[0] != expectedBytes {
					t.Fatalf("expected %d bytes on N0, got %d", expectedBytes, bytes[0])
				}
			},
		},
		{
			name:    "file path parsed",
			content: "55dc822f1000 default file=/usr/bin/ctrreschk mapped=2 N0=2 kernelpagesize_kB=4\n",
			checkFn: func(t *testing.T, nm NumaMaps) {
				if nm.VMAs[0].FilePath != "/usr/bin/ctrreschk" {
					t.Fatalf("expected file path /usr/bin/ctrreschk, got %q", nm.VMAs[0].FilePath)
				}
			},
		},
		{
			name:    "bind policy",
			content: "400000 bind:0 anon=5 N0=5 kernelpagesize_kB=4\n",
			checkFn: func(t *testing.T, nm NumaMaps) {
				if nm.VMAs[0].Policy != "bind:0" {
					t.Fatalf("expected policy bind:0, got %q", nm.VMAs[0].Policy)
				}
			},
		},
	}
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			var env environ.Environ
			if len(tt.path) > 0 {
				env = environ.Environ{
					Root: environ.FS{
						Proc: tt.path,
					},
					Log: environ.DefaultLog(),
				}
			} else if len(tt.content) > 0 {
				tmpDir := t.TempDir()
				env = environ.Environ{
					Root: environ.FS{
						Proc: tmpDir,
					},
					Log: environ.DefaultLog(),
				}
				tmpPath := NumaMapsPath(&env)
				err := os.MkdirAll(filepath.Dir(tmpPath), os.ModePerm)
				if err != nil {
					t.Fatalf("cannot prepare fake data path at %v: %v", tmpPath, err)
				}
				err = os.WriteFile(tmpPath, []byte(tt.content), 0o644)
				if err != nil {
					t.Fatalf("cannot prepare fake data file at %v: %v", tmpPath, err)
				}
			} else {
				t.Fatalf("neither path or content given; wrong test")
			}

			got, err := Read(&env)
			if tt.expectedErr && err == nil {
				t.Fatalf("expected error, got success")
			}
			if !tt.expectedErr && err != nil {
				t.Fatalf("expected success, got err=%v", err)
			}
			if tt.checkFn != nil {
				tt.checkFn(t, got)
			}
		})
	}
}
