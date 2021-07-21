package bolt

import (
	"io/ioutil"
	"log"
	"os"

	"github.com/pkg/errors"
	"github.com/portainer/portainer/api/filesystem"
)

var errTempDir = errors.New("can't create a temp dir")

func MustNewTestStore(init bool) (*Store, func()) {
	store, teardown, err := NewTestStore(init)
	if err != nil {
		if !errors.Is(err, errTempDir) {
			teardown()
		}
		log.Fatal(err)
	}

	return store, teardown
}

func NewTestStore(init bool) (*Store, func(), error) {
	// Creates unique temp directory in a concurrency friendly manner.
	dataStorePath, err := ioutil.TempDir("", "boltdb")
	if err != nil {
		return nil, nil, errors.Wrap(errTempDir, err.Error())
	}

	fileService, err := filesystem.NewService(dataStorePath, "")
	if err != nil {
		return nil, nil, err
	}

	store, err := NewStore(dataStorePath, fileService)
	if err != nil {
		return nil, nil, err
	}

	err = store.Open()
	if err != nil {
		return nil, nil, err
	}

	if init {
		err = store.Init()
		if err != nil {
			return nil, nil, err
		}
	}

	teardown := func() {
		teardown(store, dataStorePath)
	}

	return store, teardown, nil
}

func teardown(store *Store, dataStorePath string) {
	err := store.Close()
	if err != nil {
		log.Fatalln(err)
	}

	err = os.RemoveAll(dataStorePath)
	if err != nil {
		log.Fatalln(err)
	}
}
