package stackutils

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/stretchr/testify/assert"
)

func Test_GetStackFilePaths(t *testing.T) {
	stack := &portainer.Stack{
		ProjectPath: "/tmp/stack/1",
		EntryPoint:  "file-one.yml",
	}

	t.Run("stack doesn't have additional files", func(t *testing.T) {
		expected := []string{"/tmp/stack/1/file-one.yml"}
		assert.ElementsMatch(t, expected, GetStackFilePaths(stack))
	})

	t.Run("stack has additional files", func(t *testing.T) {
		stack.AdditionalFiles = []string{"file-two.yml", "file-three.yml"}
		expected := []string{"/tmp/stack/1/file-one.yml", "/tmp/stack/1/file-two.yml", "/tmp/stack/1/file-three.yml"}
		assert.ElementsMatch(t, expected, GetStackFilePaths(stack))
	})
}
