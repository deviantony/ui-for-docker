package ssl

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "ssl"
	key        = "SSL"
)

// Service represents a service for managing ssl data.
type Service struct {
	connection *internal.DbConnection
}

// NewService creates a new instance of a service.
func NewService(connection *internal.DbConnection) (*Service, error) {
	err := internal.CreateBucket(connection, BucketName)
	if err != nil {
		return nil, err
	}

	return &Service{
		connection: connection,
	}, nil
}

// Settings retrieve the ssl settings object.
func (service *Service) Settings() (*portainer.SSLSettings, error) {
	var settings portainer.SSLSettings

	err := internal.GetObject(service.connection, BucketName, []byte(key), &settings)
	if err != nil {
		return nil, err
	}

	return &settings, nil
}

// UpdateSettings persists a SSLSettings object.
func (service *Service) UpdateSettings(settings *portainer.SSLSettings) error {
	return internal.UpdateObject(service.connection, BucketName, []byte(key), settings)
}
