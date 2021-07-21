package stack

import (
	"strings"

	portainer "github.com/portainer/portainer/api"
	"github.com/portainer/portainer/api/bolt/errors"
	"github.com/portainer/portainer/api/bolt/internal"

	"github.com/boltdb/bolt"
	pkgerrors "github.com/pkg/errors"
)

const (
	// BucketName represents the name of the bucket where this service stores data.
	BucketName = "stacks"
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

// Stack returns a stack object by ID.
func (service *Service) Stack(ID portainer.StackID) (*portainer.Stack, error) {
	var stack portainer.Stack
	identifier := internal.Itob(int(ID))

	err := internal.GetObject(service.connection, BucketName, identifier, &stack)
	if err != nil {
		return nil, err
	}

	return &stack, nil
}

// StackByName returns a stack object by name.
func (service *Service) StackByName(name string) (*portainer.Stack, error) {
	var stack *portainer.Stack

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var t portainer.Stack
			err := internal.UnmarshalObject(v, &t)
			if err != nil {
				return err
			}

			if t.Name == name {
				stack = &t
				break
			}
		}

		if stack == nil {
			return errors.ErrObjectNotFound
		}

		return nil
	})

	return stack, err
}

// Stacks returns an array containing all the stacks.
func (service *Service) Stacks() ([]portainer.Stack, error) {
	var stacks = make([]portainer.Stack, 0)

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		cursor := bucket.Cursor()
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var stack portainer.Stack
			err := internal.UnmarshalObject(v, &stack)
			if err != nil {
				return err
			}
			stacks = append(stacks, stack)
		}

		return nil
	})

	return stacks, err
}

// GetNextIdentifier returns the next identifier for a stack.
func (service *Service) GetNextIdentifier() int {
	return internal.GetNextIdentifier(service.connection, BucketName)
}

// CreateStack creates a new stack.
func (service *Service) CreateStack(stack *portainer.Stack) error {
	return service.connection.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))

		// We manually manage sequences for stacks
		err := bucket.SetSequence(uint64(stack.ID))
		if err != nil {
			return err
		}

		data, err := internal.MarshalObject(stack)
		if err != nil {
			return err
		}

		return bucket.Put(internal.Itob(int(stack.ID)), data)
	})
}

// UpdateStack updates a stack.
func (service *Service) UpdateStack(ID portainer.StackID, stack *portainer.Stack) error {
	identifier := internal.Itob(int(ID))
	return internal.UpdateObject(service.connection, BucketName, identifier, stack)
}

// DeleteStack deletes a stack.
func (service *Service) DeleteStack(ID portainer.StackID) error {
	identifier := internal.Itob(int(ID))
	return internal.DeleteObject(service.connection, BucketName, identifier)
}

// StackByWebhookID returns a pointer to a stack object by webhook ID.
// It returns nil, errors.ErrObjectNotFound if there's no stack associated with the webhook ID.
func (service *Service) StackByWebhookID(id string) (*portainer.Stack, error) {
	if id == "" {
		return nil, pkgerrors.New("webhook ID can't be empty string")
	}
	var stack portainer.Stack
	found := false

	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()

		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			var t struct {
				AutoUpdate *struct {
					WebhookID string `json:"Webhook"`
				} `json:"AutoUpdate"`
			}

			err := internal.UnmarshalObject(v, &t)
			if err != nil {
				return err
			}

			if t.AutoUpdate != nil && strings.EqualFold(t.AutoUpdate.WebhookID, id) {
				found = true
				err := internal.UnmarshalObject(v, &stack)
				if err != nil {
					return err
				}
				break
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}
	if !found {
		return nil, errors.ErrObjectNotFound
	}

	return &stack, nil
}

// RefreshableStacks returns stacks that are configured for a periodic update
func (service *Service) RefreshableStacks() ([]portainer.Stack, error) {
	stacks := make([]portainer.Stack, 0)
	err := service.connection.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(BucketName))
		cursor := bucket.Cursor()

		var stack portainer.Stack
		for k, v := cursor.First(); k != nil; k, v = cursor.Next() {
			err := internal.UnmarshalObject(v, &stack)
			if err != nil {
				return err
			}

			if stack.AutoUpdate != nil && stack.AutoUpdate.Interval != "" {
				stacks = append(stacks, stack)
			}
		}

		return nil
	})

	return stacks, err
}
