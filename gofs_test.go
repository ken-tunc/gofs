package gofs

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyFile(t *testing.T) {
	var tmpRoot string

	joinPath := func(args ...string) string {
		return filepath.Join(append([]string{tmpRoot}, args...)...)
	}

	makeTmpDir := func(perm os.FileMode, dir string) {
		if err := os.Mkdir(filepath.Join(tmpRoot, dir), perm); err != nil {
			t.Fatal(err)
		}
	}

	makeTmpFile := func(perm os.FileMode, contents string, path ...string) {
		if err := ioutil.WriteFile(joinPath(path...), []byte(contents), perm); err != nil {
			t.Fatal(err)
		}
	}

	tests := []struct {
		name     string
		setup    func(t *testing.T)
		dst, src string // automatically joined to tmpRoot
		check    func(t *testing.T, src, dst string, err error)
	}{{
		name: "dst file doesn't exist",
		setup: func(*testing.T) {
			makeTmpDir(0755, "a")
			makeTmpFile(0644, "file1", "a", "file1")
		},
		dst: "a/file2",
		src: "a/file1",
	}, {
		name: "dst file exists",
		setup: func(*testing.T) {
			makeTmpDir(0755, "a")
			makeTmpFile(0644, "file1", "a", "file1")
			makeTmpFile(0644, "file2", "a", "file2")
		},
		dst: "a/file2",
		src: "a/file1",
	}}

	root, err := ioutil.TempDir("", "gofs_test_copy_file")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(root)

	for i, tt := range tests {
		tmpRoot, err = ioutil.TempDir(root, fmt.Sprintf("test-%d", i))
		if err != nil {
			t.Fatal(err)
		}
		tt.setup(t)
		src := joinPath(filepath.FromSlash(tt.src))
		dst := joinPath(filepath.FromSlash(tt.dst))

		err := CopyFile(dst, src)

		if err != nil {
			t.Errorf("TestCopyFile() error = %v", err)
		}
	}
}

func TestEnsurePath(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "gofs_test_ensure_path")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpDir)

	tmpFile, err := ioutil.TempFile(tmpDir, "tmpFile")
	if err != nil {
		t.Fatal(err)
	}
	unCreatedFile := filepath.Join(tmpDir, "newDir", "testFile.txt")

	type args struct {
		path string
		perm os.FileMode
	}
	tests := []struct {
		name  string
		args  args
		want  string
		check func(absPath string, t *testing.T)
	}{
		{
			name: "File exists",
			args: args{path: tmpFile.Name(), perm: 0755},
			want: filepath.Join(tmpDir, filepath.Base(tmpFile.Name())),
			check: func(absPath string, t *testing.T) {
				fp, err := os.Open(absPath)
				if err != nil {
					t.Errorf("EnsurePath() cannot open file. path = %s, error = %v", absPath, err)
				}
				fp.Close()
			},
		},
		{
			name: "File doesn't exist",
			args: args{path: unCreatedFile, perm: 0755},
			want: unCreatedFile,
			check: func(absPath string, t *testing.T) {
				fp, err := os.Create(absPath)
				if err != nil {
					t.Errorf("EnsurePath() cannot open file. path = %s, error = %v", absPath, err)
				}
				fp.Close()
			},
		},
	}

	for _, tt := range tests {
		got, err := EnsurePath(tt.args.path, tt.args.perm)
		if err != nil {
			t.Errorf("EnsurePath() error = %v", err)
			return
		}
		if got != tt.want {
			t.Errorf("EnsurePath() got = %v, want %v", got, tt.want)
		}
		tt.check(got, t)
	}
}

func TestFileExists(t *testing.T) {
	tmp, err := ioutil.TempFile("", "gofs_test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmp.Name())

	type args struct {
		path string
	}
	tests := []struct {
		name    string
		args    args
		want    bool
		wantErr bool
	}{
		{
			name:    "File exists",
			args:    args{path: tmp.Name()},
			want:    true,
			wantErr: false,
		},
		{
			name:    "File doesn't exist",
			args:    args{path: tmp.Name() + "invalid_path"},
			want:    false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := FileExists(tt.args.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("FileExists() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FileExists() got = %v, want %v", got, tt.want)
			}
		})
	}
}
