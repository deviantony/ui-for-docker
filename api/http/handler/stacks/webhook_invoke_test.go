package stacks

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"

	portainer "github.com/portainer/portainer/api"

	"github.com/portainer/portainer/api/bolt/bolttest"
)

func TestHandler_webhookInvoke(t *testing.T) {
	store, teardown := bolttest.MustNewTestStore(true)
	defer teardown()

	webhookID := newGuidString(t)
	store.StackService.CreateStack(&portainer.Stack{
		AutoUpdate: &portainer.StackAutoUpdate{
			Webhook: webhookID,
		},
	})

	h := NewHandler(nil)
	h.DataStore = store

	t.Run("invalid uuid results in http.StatusBadRequest", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest("notuuid")
		h.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
	t.Run("registered webhook ID in http.StatusNoContent", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(webhookID)
		h.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNoContent, w.Code)
	})
	t.Run("unregistered webhook ID in http.StatusNotFound", func(t *testing.T) {
		w := httptest.NewRecorder()
		req := newRequest(newGuidString(t))
		h.Router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}

func newGuidString(t *testing.T) string {
	uuid, err := uuid.NewV4()
	assert.NoError(t, err)

	return uuid.String()
}

func newRequest(webhookID string) *http.Request {
	return httptest.NewRequest(http.MethodPost, "/stacks/webhook/"+webhookID, nil)
}
