package filesystem

import (
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_copyFile_returnsError_whenSourceDoesNotExist(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	err := copyFile("does-not-exist", tmpdir)
	assert.Error(t, err)
}

func Test_copyFile_shouldMakeAbackup(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	content := []byte("content")
	ioutil.WriteFile(path.Join(tmpdir, "origin"), content, 0600)

	err := copyFile(path.Join(tmpdir, "origin"), path.Join(tmpdir, "copy"))
	assert.NoError(t, err)

	copyContent, _ := ioutil.ReadFile(path.Join(tmpdir, "copy"))
	assert.Equal(t, content, copyContent)
}

func Test_CopyDir_shouldCopyAllFilesAndDirectories(t *testing.T) {
	destination, _ := ioutil.TempDir("", "destination")
	defer os.RemoveAll(destination)
	err := CopyDir("./testdata/copy_test", destination, true)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(destination, "copy_test", "outer"))
	assert.FileExists(t, filepath.Join(destination, "copy_test", "dir", ".dotfile"))
	assert.FileExists(t, filepath.Join(destination, "copy_test", "dir", "inner"))
}

func Test_CopyDir_shouldCopyOnlyDirContents(t *testing.T) {
	destination, _ := ioutil.TempDir("", "destination")
	defer os.RemoveAll(destination)
	err := CopyDir("./testdata/copy_test", destination, false)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(destination, "outer"))
	assert.FileExists(t, filepath.Join(destination, "dir", ".dotfile"))
	assert.FileExists(t, filepath.Join(destination, "dir", "inner"))
}

func Test_CopyPath_shouldSkipWhenNotExist(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	err := CopyPath("does-not-exists", tmpdir)
	assert.NoError(t, err)

	assert.NoFileExists(t, tmpdir)
}

func Test_CopyPath_shouldCopyFile(t *testing.T) {
	tmpdir, _ := ioutil.TempDir("", "backup")
	defer os.RemoveAll(tmpdir)

	content := []byte("content")
	ioutil.WriteFile(path.Join(tmpdir, "file"), content, 0600)

	os.MkdirAll(path.Join(tmpdir, "backup"), 0700)
	err := CopyPath(path.Join(tmpdir, "file"), path.Join(tmpdir, "backup"))
	assert.NoError(t, err)

	copyContent, err := ioutil.ReadFile(path.Join(tmpdir, "backup", "file"))
	assert.NoError(t, err)
	assert.Equal(t, content, copyContent)
}

func Test_CopyPath_shouldCopyDir(t *testing.T) {
	destination, _ := ioutil.TempDir("", "destination")
	defer os.RemoveAll(destination)
	err := CopyPath("./testdata/copy_test", destination)
	assert.NoError(t, err)

	assert.FileExists(t, filepath.Join(destination, "copy_test", "outer"))
	assert.FileExists(t, filepath.Join(destination, "copy_test", "dir", ".dotfile"))
	assert.FileExists(t, filepath.Join(destination, "copy_test", "dir", "inner"))
}
