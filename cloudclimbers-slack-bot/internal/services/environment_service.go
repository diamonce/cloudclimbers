package services

import (
    "cloudclimbers-slack-bot/internal/models"
    "cloudclimbers-slack-bot/internal/repository"
)

type EnvironmentService struct {
    repo repository.EnvironmentRepository
}

func NewEnvironmentService(repo repository.EnvironmentRepository) *EnvironmentService {
    return &EnvironmentService{repo: repo}
}

func (s *EnvironmentService) CreateEnvironment(env *models.Environment) error {
    return s.repo.CreateEnvironment(env)
}

func (s *EnvironmentService) GetEnvironment(id string) (*models.Environment, error) {
    return s.repo.GetEnvironment(id)
}

func (s *EnvironmentService) DeleteEnvironment(id string) error {
    return s.repo.DeleteEnvironment(id)
}
