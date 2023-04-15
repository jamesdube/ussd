package session

import (
	"github.com/jamesdube/ussd/internal/utils"
	"go.uber.org/zap"
)

type Session struct {
	Id         string            `json:"id"`
	Attributes map[string]string `json:"attributes"`
	Selections []string          `json:"selections"`
	Active     bool              `json:"active"`
}

func NewSession(id string) *Session {
	utils.Logger.Info("creating new session", zap.String("sessionId", id))
	return &Session{
		Id:         id,
		Attributes: map[string]string{},
	}
}

func (s *Session) AddSelection(m string) {
	s.Selections = append(s.Selections, m)
}

func (s *Session) RemoveLastSelection() {
	if len(s.Selections) > 0 {
		s.Selections = s.Selections[:len(s.Selections)-1]
	}
}

func (s *Session) GetSelections() []string {
	return s.Selections
}

func (s *Session) GetID() string {
	return s.Id
}
