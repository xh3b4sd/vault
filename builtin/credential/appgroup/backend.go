package appgroup

import (
	"sync"

	"github.com/hashicorp/vault/helper/salt"
	"github.com/hashicorp/vault/logical"
	"github.com/hashicorp/vault/logical/framework"
)

func Factory(conf *logical.BackendConfig) (logical.Backend, error) {
	b, err := Backend(conf)
	if err != nil {
		return nil, err
	}
	return b.Setup(conf)
}

func Backend(conf *logical.BackendConfig) (*framework.Backend, error) {
	// Initialize the salt
	salt, err := salt.NewSalt(conf.StorageView, &salt.Config{
		HashFunc: salt.SHA256Hash,
	})
	if err != nil {
		return nil, err
	}

	// Create a backend object
	b := &backend{
		salt:        salt,
		appLock:     &sync.RWMutex{},
		groupLock:   &sync.RWMutex{},
		genericLock: &sync.RWMutex{},
	}

	// Attach the paths and secrets that are to be handled by the backend
	b.Backend = &framework.Backend{
		Help:      backendHelp,
		AuthRenew: b.pathLoginRenew,
		PathsSpecial: &logical.Paths{
			Unauthenticated: []string{
				"login",
			},
		},
		Paths: framework.PathAppend(
			appPaths(b),
			groupPaths(b),
			genericPaths(b),
			[]*framework.Path{
				pathLogin(b),
			},
		),
	}
	return b.Backend, nil
}

type backend struct {
	*framework.Backend
	salt        *salt.Salt
	appLock     *sync.RWMutex
	groupLock   *sync.RWMutex
	genericLock *sync.RWMutex
	userIDLocks map[string]*sync.RWMutex
}

const backendHelp = `
`
