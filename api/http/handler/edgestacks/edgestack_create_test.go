package edgestacks

import (
	"testing"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/internal/testhelpers"
	"github.com/stretchr/testify/assert"
)

func Test_updateEndpointRelation_successfulRuns(t *testing.T) {
	edgeStackID := portainer.EdgeStackID(5)
	endpointRelations := []portainer.EndpointRelation{
		{EndpointID: 1, EdgeStacks: map[portainer.EdgeStackID]bool{}},
		{EndpointID: 2, EdgeStacks: map[portainer.EdgeStackID]bool{}},
		{EndpointID: 3, EdgeStacks: map[portainer.EdgeStackID]bool{}},
		{EndpointID: 4, EdgeStacks: map[portainer.EdgeStackID]bool{}},
		{EndpointID: 5, EdgeStacks: map[portainer.EdgeStackID]bool{}},
	}

	relatedIds := []portainer.EndpointID{2, 3}

	dataStore := testhelpers.NewDatastore(testhelpers.WithEndpointRelations(endpointRelations))

	err := updateEndpointRelations(dataStore.EndpointRelation(), edgeStackID, relatedIds)

	assert.NoError(t, err, "updateEndpointRelations should not fail")

	relatedSet := map[portainer.EndpointID]bool{}
	for _, relationID := range relatedIds {
		relatedSet[relationID] = true
	}

	for _, relation := range endpointRelations {
		shouldBeRelated := relatedSet[relation.EndpointID]
		assert.Equal(t, shouldBeRelated, relation.EdgeStacks[edgeStackID])
	}
}
