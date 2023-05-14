package session

import "sync"

type InMemory struct {
	sessions map[string]*Session
	mu       sync.RWMutex
}

func NewInMemory() *InMemory {
	return &InMemory{
		sessions: map[string]*Session{},
	}
}

func (im *InMemory) AddSelection(s string) {

}

func (im *InMemory) GetSession(id string) (*Session, error) {

	for _, s := range im.sessions {
		if s.GetID() == id {
			return s, nil
		}
	}

	return NewSession(id), nil

}

func (im *InMemory) Save(s *Session) error {
	im.mu.Lock()
	defer im.mu.Unlock()
	im.sessions[s.GetID()] = s
	return nil
}

func (im *InMemory) Delete(id string) {
	im.mu.Lock()
	defer im.mu.Unlock()
	delete(im.sessions, id)
}
