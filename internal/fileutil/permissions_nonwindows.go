//go:build !windows

package fileutil

import "os"

func hardenJSONDirPermissions(path string) error {
	return os.Chmod(path, 0700)
}

func hardenJSONFilePermissions(path string) error {
	return os.Chmod(path, 0600)
}
