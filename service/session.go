package service

import (
	"a21hc3NpZ25tZW50/model"
	repo "a21hc3NpZ25tZW50/repository"
)

type SessionService interface {
	GetSessionByEmail(email string) (model.Session, error)
}

type sessionService struct {
	sessionRepo repo.SessionRepository
}

func NewSessionService(sessionRepo repo.SessionRepository) *sessionService {
	return &sessionService{sessionRepo}
}

func (s *sessionService) GetSessionByEmail(email string) (model.Session, error) {
	session, err := s.sessionRepo.SessionAvailEmail(email)
	if err != nil {
		return model.Session{}, err
	}

	return session, nil
}
