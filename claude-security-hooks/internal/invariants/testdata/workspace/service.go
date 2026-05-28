package workspace

// Service holds the application state.
type Service struct {
	Name string
}

// New creates a new Service instance.
func New(name string) *Service {
	return &Service{Name: name}
}
