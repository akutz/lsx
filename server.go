package lsx

import (
	"context"
	"io"
	"sync"
)

// Server is the interface for a server.
//
// When the server's Init function is invoked, the configuration
// map will have a key named "addrs" that is an array of strings
// and will contain the network addresses on which the server is
// supposed to listen when its Serve function is invoked.
type Server interface {
	Module
	io.Closer

	// Serve instructs the server to start listening on the addresses
	// provided via the config with which the server was initialized.
	//
	// This function returns a channel on which errors are received.
	// Reading this channel is also the prescribed manner for clients
	// wishing to block until the server is shutdown as the error
	// channel will be closed when the server is stopped.
	Serve(ctx context.Context) (<-chan error, error)
}

type serverCtor func() Server

var (
	serverCtors    = map[string]serverCtor{}
	serverCtorsRWL = sync.RWMutex{}
)

// RegisterServer registers the name of a new server type
// and the function used to create a new server object.
func RegisterServer(name string, ctor serverCtor) {
	serverCtorsRWL.Lock()
	defer serverCtorsRWL.Unlock()
	serverCtors[name] = ctor
}

// Servers returns a channel on which constructed server objects
// for all registered servers are returned.
func Servers() <-chan Server {
	serverCtorsRWL.RLock()
	defer serverCtorsRWL.RUnlock()
	c := make(chan Server)
	go func() {
		for _, ctor := range serverCtors {
			c <- ctor()
		}
	}()
	return c
}
