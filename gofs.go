package gofs

import (
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"syscall"
)

// CopyFile copies src file to dst file.
func CopyFile(dst string, src string) error {
	srcInfo, err := os.Stat(src)
	if err != nil {
		return err
	}
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}

	tmp, err := ioutil.TempFile(filepath.Dir(dst), filepath.Base(dst))
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err = io.Copy(tmp, srcFile); err != nil {
		return err
	}
	if err = tmp.Close(); err != nil {
		return err
	}

	if err = os.Chmod(tmp.Name(), srcInfo.Mode()); err != nil {
		return err
	}
	if stat, ok := srcInfo.Sys().(*syscall.Stat_t); ok {
		err = os.Chown(tmp.Name(), int(stat.Uid), int(stat.Gid))
		if err != nil {
			return err
		}
	}

	return os.Rename(tmp.Name(), dst)
}

// EnsurePath returns absolute path and creates directories if they don't exist.
// perm will be passed to os.MkdirAll for creating directories.
func EnsurePath(path string, perm os.FileMode) (string, error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}

	dir := filepath.Dir(absPath)
	_, err = os.Stat(dir)
	if os.IsNotExist(err) {
		if err = os.MkdirAll(dir, perm); err != nil {
			return "", err
		}
	} else if err != nil {
		return "", err
	}

	return absPath, nil
}

// FileExists returns true if named path exists.
func FileExists(path string) (bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return true, err
	}

	return !info.IsDir(), nil
}
