package backup

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func listFiles(dir string) []string {
	items := make([]string, 0)
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if path == dir {
			return nil
		}
		items = append(items, path)
		return nil
	})

	return items
}

func contains(t *testing.T, list []string, path string) {
	assert.Contains(t, list, path)
	copyContent, _ := ioutil.ReadFile(path)
	assert.Equal(t, "content\n", string(copyContent))
}

func Test_copyFile_returnsError_whenSourceDoesNotExist(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	err := copyFile("does-not-exist", tmpdir)
	assert.NotNil(t, err)
}

func Test_copyFile_shouldMakeAbackup(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	content := []byte("content")
	ioutil.WriteFile(path.Join(tmpdir, "origin"), content, 0600)

	err := copyFile(path.Join(tmpdir, "origin"), path.Join(tmpdir, "copy"))
	assert.Nil(t, err)

	copyContent, _ := ioutil.ReadFile(path.Join(tmpdir, "copy"))
	assert.Equal(t, content, copyContent)
}

func Test_copyDir_shouldCopyAllFilesAndDirectories(t *testing.T) {
	destination, _ := ioutil.TempDir("", "destination")
	defer os.RemoveAll(destination)
	err := copyDir("./test_assets/copy_test", destination)
	assert.Nil(t, err)

	createdFiles := listFiles(destination)

	contains(t, createdFiles, filepath.Join(destination, "copy_test", "outer"))
	contains(t, createdFiles, filepath.Join(destination, "copy_test", "dir", ".dotfile"))
	contains(t, createdFiles, filepath.Join(destination, "copy_test", "dir", "inner"))
}

func Test_backupPath_shouldSkipWhenNotExist(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	err := copyPath("does-not-exists", tmpdir)
	assert.Nil(t, err)

	assert.Empty(t, listFiles(tmpdir))
}

func Test_backupPath_shouldCopyFile(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	content := []byte("content")
	ioutil.WriteFile(path.Join(tmpdir, "file"), content, 0600)

	os.MkdirAll(path.Join(tmpdir, "backup"), 0700)
	err := copyPath(path.Join(tmpdir, "file"), path.Join(tmpdir, "backup"))
	assert.Nil(t, err)

	copyContent, err := ioutil.ReadFile(path.Join(tmpdir, "backup", "file"))
	assert.Nil(t, err)
	assert.Equal(t, content, copyContent)
}

func Test_backupPath_shouldCopyDir(t *testing.T) {
	destination, _ := ioutil.TempDir("", "destination")
	defer os.RemoveAll(destination)
	err := copyPath("./test_assets/copy_test", destination)
	assert.Nil(t, err)

	createdFiles := listFiles(destination)

	contains(t, createdFiles, filepath.Join(destination, "copy_test", "outer"))
	contains(t, createdFiles, filepath.Join(destination, "copy_test", "dir", ".dotfile"))
	contains(t, createdFiles, filepath.Join(destination, "copy_test", "dir", "inner"))
}
