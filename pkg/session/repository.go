package session

// Repository defines the interface for session storage backends.
type Repository interface {
	GetSession(id string) (*Session, error)
	Save(s *Session) error
	Delete(id string)
}
