package storage

import (
	"path/filepath"
	"testing"
)

func TestSanitizePath(t *testing.T) {
	base := "/srv/data"

	tests := []struct {
		name    string
		base    string
		path    string
		wantErr bool
		wantOut string
	}{
		{
			name:    "simple valid filename",
			base:    base,
			path:    "hello.txt",
			wantErr: false,
			wantOut: filepath.Join(base, "hello.txt"),
		},
		{
			name:    "nested valid path",
			base:    base,
			path:    "subdir/file.png",
			wantErr: false,
			wantOut: filepath.Join(base, "subdir/file.png"),
		},
		{
			name:    "dot-dot traversal rejected",
			base:    base,
			path:    "../secret.txt",
			wantErr: true,
		},
		{
			name:    "absolute traversal through subdirectory rejected",
			base:    base,
			path:    "subdir/../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "triple dot-dot rejected",
			base:    base,
			path:    "../../etc/hosts",
			wantErr: true,
		},
		{
			name:    "current directory dot",
			base:    base,
			path:    "./file.txt",
			wantErr: false,
			wantOut: filepath.Join(base, "file.txt"),
		},
		{
			name:    "deeply nested valid path",
			base:    base,
			path:    "a/b/c/d/file.bin",
			wantErr: false,
			wantOut: filepath.Join(base, "a/b/c/d/file.bin"),
		},
		{
			name:    "filename with spaces",
			base:    base,
			path:    "my file.txt",
			wantErr: false,
			wantOut: filepath.Join(base, "my file.txt"),
		},
		{
			name:    "dot-dot in middle of path",
			base:    base,
			path:    "foo/../../../etc/shadow",
			wantErr: true,
		},
		{
			name:    "bare filename no extension",
			base:    base,
			path:    "myfile",
			wantErr: false,
			wantOut: filepath.Join(base, "myfile"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := SanitizePath(tc.base, tc.path)
			if tc.wantErr {
				if err == nil {
					t.Errorf("expected error for path %q, got nil (result=%q)", tc.path, got)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error for path %q: %v", tc.path, err)
				return
			}
			if got != tc.wantOut {
				t.Errorf("SanitizePath(%q, %q) = %q; want %q", tc.base, tc.path, got, tc.wantOut)
			}
		})
	}
}
