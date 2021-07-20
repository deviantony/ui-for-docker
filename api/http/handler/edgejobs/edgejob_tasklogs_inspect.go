package edgejobs

import (
	"net/http"
	"strconv"

	httperror "github.com/portainer/libhttp/error"
	"github.com/portainer/libhttp/request"
	"github.com/portainer/libhttp/response"
)

type fileResponse struct {
	FileContent string `json:"FileContent"`
}

// @id EdgeJobTaskLogsInspect
// @summary Fetch the log for a specifc task on an EdgeJob
// @description
// @tags edge_jobs
// @security jwt
// @accept json
// @produce json
// @param id path string true "EdgeJob Id"
// @param taskID path string true "Task Id"
// @success 200 {object} fileResponse
// @failure 500
// @failure 400
// @failure 503 "Edge compute features are disabled"
// @router /edge_jobs/{id}/tasks/{taskID}/logs [get]
func (handler *Handler) edgeJobTaskLogsInspect(w http.ResponseWriter, r *http.Request) *httperror.HandlerError {
	edgeJobID, err := request.RetrieveNumericRouteVariableValue(r, "id")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Edge job identifier route variable", err}
	}

	taskID, err := request.RetrieveNumericRouteVariableValue(r, "taskID")
	if err != nil {
		return &httperror.HandlerError{http.StatusBadRequest, "Invalid Task identifier route variable", err}
	}

	logFileContent, err := handler.FileService.GetEdgeJobTaskLogFileContent(strconv.Itoa(edgeJobID), strconv.Itoa(taskID))
	if err != nil {
		return &httperror.HandlerError{http.StatusInternalServerError, "Unable to retrieve log file from disk", err}
	}

	return response.JSON(w, &fileResponse{FileContent: string(logFileContent)})
}

// fmt.Sprintf("/tmp/edge_jobs/%s/logs_%s", edgeJobID, taskID)
