package stackutils

import (
	"fmt"
	"path"

	portainer "github.com/portainer/portainer/api"
)

// ResourceControlID returns the stack resource control id
func ResourceControlID(endpointID portainer.EndpointID, name string) string {
	return fmt.Sprintf("%d_%s", endpointID, name)
}

// GetStackFilePaths returns a list of file paths based on stack project path
func GetStackFilePaths(stack *portainer.Stack) []string {
	var filePaths []string
	for _, file := range append([]string{stack.EntryPoint}, stack.AdditionalFiles...) {
		filePaths = append(filePaths, path.Join(stack.ProjectPath, file))
	}
	return filePaths
}
