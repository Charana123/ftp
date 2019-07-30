package server

// UserContext encapsulates information about the connecting user
type UserContext struct {
	User string // Username of connecting user
	CWD  string // Current working directory of connecting user
}

// newUserContext creates a default UserContext
func newUserContext() *UserContext {
	return &UserContext{
		CWD: globalServerSettings.PublicDirectory,
	}
}
