package session

type InMemory struct {
	sessions map[string]*Session
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
	im.sessions[s.GetID()] = s
	return nil
}

func (im *InMemory) Delete(id string) {
	delete(im.sessions, id)
}
