package edgestacks

import (
	"net/http"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/response"
)

// @id EdgeStackList
// @summary Fetches the list of EdgeStacks
// @description
// @tags edge_stacks
// @security jwt
// @accept json
// @produce json
// @success 200 {array} portainer.EdgeStack
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_stacks [get]
func (handler *Handler) edgeStackList(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeStacks, err := handler.DataStore.EdgeStack().EdgeStacks()
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve edge stacks from the database", err}
	}

	return response.JSON(w, edgeStacks)
}
