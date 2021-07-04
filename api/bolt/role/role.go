package role

import (
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/internal"

	"github.com/boltdb/bolt"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "roles"
)

// Service represents a service for managing endpoint data.
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

// Role returns a Role by ID
func (service *Service) Role(ID portainer.RoleID) (*portainer.Role, error) {
	var set portainer.Role
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &set)
	if err != nil {
		return nil, err
	}

	return &set, nil
}

// Roles return an array containing all the sets.
func (service *Service) Roles() ([]portainer.Role, error) {
	var sets = make([]portainer.Role, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var set portainer.Role
			err := internal.UnmarshalObject(v, &set)
			if err != nil {
				return err
			}
			sets = append(sets, set)
		}

		return nil
	})

	return sets, err
}

// CreateRole creates a new Role.
func (service *Service) CreateRole(role *portainer.Role) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		if role.ID == 0 {
			id, _ := bucket.NextSequence()
			role.ID = portainer.RoleID(id)
		}

		data, err := internal.MarshalObject(role)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(role.ID)), data)
	})
}

// UpdateRole updates a role.
func (service *Service) UpdateRole(ID portainer.RoleID, role *portainer.Role) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, role)
}
