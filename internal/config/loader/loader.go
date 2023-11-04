package loader

import (
	"context"
	"fmt"
	"net/url"
	"sync"
)

var (
	mu      sync.RWMutex
	loaders = make(map[string]Loader)
)

type Loader interface {
	LoadWithContext(context.Context, *url.URL) ([]byte, error)
}

// Register a loader.
func Register(name string, loader Loader) {
	mu.Lock()
	defer mu.Unlock()
	if loader == nil {
		panic("The loader to be registered is nil.")
	}
	if _, ok := getRegisteredLoader(name); ok {
		panic(fmt.Sprintf("The loader '%s' is already registered.", name))
	}
	loaders[name] = loader
}

// LoadWithContext returns the contents of the file loaded from the loader as a byte array.
func LoadWithContext(ctx context.Context, u *url.URL) ([]byte, error) {
	name := u.Scheme
	if name == "" {
		name = "file"
	}
	loader, ok := getRegisteredLoader(name)
	if !ok {
		return nil, fmt.Errorf("There is no loader registered for the '%s' schema.", name)
	}
	return loader.LoadWithContext(ctx, u)
}

func getRegisteredLoader(name string) (Loader, bool) {
	loader, ok := loaders[name]
	return loader, ok
}
