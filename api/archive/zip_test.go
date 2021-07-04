package archive

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUnzipFile(t *testing.T) {
	dir, err := ioutil.TempDir("", "unzip-test-")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)
	/*
		Archive structure.
		├── 0
		│	├── 1
		│	│	└── 2.txt
		│	└── 1.txt
		└── 0.txt
	*/

	err = UnzipFile("./testdata/sample_archive.zip", dir)

	assert.NoError(t, err)
	archiveDir := dir + "/sample_archive"
	assert.FileExists(t, filepath.Join(archiveDir, "0.txt"))
	assert.FileExists(t, filepath.Join(archiveDir, "0", "1.txt"))
	assert.FileExists(t, filepath.Join(archiveDir, "0", "1", "2.txt"))

}
