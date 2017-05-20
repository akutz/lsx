package lsx

// ContextKey is a key used to store context values in a Go context.
type ContextKey uint

const (
	// ConfigKey is the context key used to store and retrieve a
	// Config object in and from a Go context.
	ConfigKey ContextKey = iota
)
