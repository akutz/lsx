package lsx

// Service is the interface for a storage service.
type Service interface {
	Module

	// Driver returns the name of the storage driver used by the service.
	Driver() string
}
