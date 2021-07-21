package auth

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func TestIsLDAPAdmin_Match(t *testing.T) {

	h := Handler{
		LDAPService: testhelpers.NewLDAPService(),
	}
	mockLDAPSettings := &portainer.LDAPSettings{
		AdminGroups: []string{"manager", "operator"},
	}

	matched, err := h.isLDAPAdmin("username", mockLDAPSettings)
	assert.NoError(t, err)
	assert.Equal(t, true, matched)
}

func TestIsLDAPAdmin_NotMatch(t *testing.T) {

	h := Handler{
		LDAPService: testhelpers.NewLDAPService(),
	}
	mockLDAPSettings := &portainer.LDAPSettings{
		AdminGroups: []string{"admin", "operator"},
	}

	matched, err := h.isLDAPAdmin("username", mockLDAPSettings)
	assert.NoError(t, err)
	assert.Equal(t, false, matched)
}
