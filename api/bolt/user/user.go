package user

import (
	"strings"

	"github.com/boltdb/bolt"
	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/bolt/internal"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "users"
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

// User returns a user by ID
func (service *Service) User(ID portainer.UserID) (*portainer.User, error) {
	var user portainer.User
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// UserByUsername returns a user by username.
func (service *Service) UserByUsername(username string) (*portainer.User, error) {
	var user *portainer.User

	username = strings.ToLower(username)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var u portainer.User
			err := internal.UnmarshalObject(v, &u)
			if err != nil {
				return err
			}

			if strings.EqualFold(u.Username, username) {
				user = &u
				break
			}
		}

		if user == nil {
			return errors.ErrObjectNotFound
		}
		return nil
	})

	return user, err
}

// Users return an array containing all the users.
func (service *Service) Users() ([]portainer.User, error) {
	var users = make([]portainer.User, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var user portainer.User
			err := internal.UnmarshalObject(v, &user)
			if err != nil {
				return err
			}
			users = append(users, user)
		}

		return nil
	})

	return users, err
}

// UsersByRole return an array containing all the users with the specified role.
func (service *Service) UsersByRole(role portainer.UserRole) ([]portainer.User, error) {
	var users = make([]portainer.User, 0)
	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var user portainer.User
			err := internal.UnmarshalObject(v, &user)
			if err != nil {
				return err
			}

			if user.Role == role {
				users = append(users, user)
			}
		}
		return nil
	})

	return users, err
}

// UpdateUser saves a user.
func (service *Service) UpdateUser(ID portainer.UserID, user *portainer.User) error {
	identifier := internal.Itob(int(ID))
	user.Username = strings.ToLower(user.Username)
	return internal.UpdateObject(service.connection, BucketName, identifier, user)
}

// CreateUser creates a new user.
func (service *Service) CreateUser(user *portainer.User) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		id, _ := bucket.NextSequence()
		user.ID = portainer.UserID(id)
		user.Username = strings.ToLower(user.Username)

		data, err := internal.MarshalObject(user)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(user.ID)), data)
	})
}

// DeleteUser deletes a user.
func (service *Service) DeleteUser(ID portainer.UserID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}
