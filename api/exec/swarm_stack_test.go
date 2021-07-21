package exec

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfigFilePaths(t *testing.T) {
	args := []string{"stack", "deploy", "--with-registry-auth"}
	filePaths := []string{"dir/file", "dir/file-two", "dir/file-three"}
	expected := []string{"stack", "deploy", "--with-registry-auth", "--compose-file", "dir/file", "--compose-file", "dir/file-two", "--compose-file", "dir/file-three"}
	output := configureFilePaths(args, filePaths)
	assert.ElementsMatch(t, expected, output, "wrong output file paths")
}
