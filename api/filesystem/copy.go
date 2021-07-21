package filesystem

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// CopyPath copies file or directory defined by the path to the toDir path
func CopyPath(path string, toDir string) error {
	info, err := os.Stat(path)
	if err != nil && errors.Is(err, os.ErrNotExist) {
		// skip copy if file does not exist
		return nil
	}

	if !info.IsDir() {
		destination := filepath.Join(toDir, info.Name())
		return copyFile(path, destination)
	}

	return CopyDir(path, toDir, true)
}

// CopyDir copies contents of fromDir to toDir.
// When keepParent is true, contents will be copied with their immediate parent dir,
// i.e. given /from/dirA and /to/dirB with keepParent == true, result will be /to/dirB/dirA/<children>
func CopyDir(fromDir, toDir string, keepParent bool) error {
	cleanedSourcePath := filepath.Clean(fromDir)
	parentDirectory := filepath.Dir(cleanedSourcePath)
	err := filepath.Walk(cleanedSourcePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		var destination string
		if keepParent {
			destination = filepath.Join(toDir, strings.TrimPrefix(path, parentDirectory))
		} else {
			destination = filepath.Join(toDir, strings.TrimPrefix(path, cleanedSourcePath))
		}

		if destination == "" {
			return nil
		}

		if info.IsDir() {
			return nil // skip directory creations
		}

		if info.Mode()&os.ModeSymlink != 0 { // entry is a symlink
			return nil // don't copy symlinks
		}

		return copyFile(path, destination)
	})

	return err
}

// copies regular a file from src to dst
func copyFile(src, dst string) error {
	from, err := os.Open(src)
	if err != nil {
		return err
	}
	defer from.Close()

	// has to include 'execute' bit, otherwise fails. MkdirAll follows `mkdir -m` restrictions
	if err := os.MkdirAll(filepath.Dir(dst), 0744); err != nil {
		return err
	}
	to, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer to.Close()

	_, err = io.Copy(to, from)
	return err
}
