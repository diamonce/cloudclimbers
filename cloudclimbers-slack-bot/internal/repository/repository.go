package repository

import "cloudclimbers-slack-bot/internal/models"

type EnvironmentRepository interface {
    CreateEnvironment(env *models.Environment) error
    GetEnvironment(id string) (*models.Environment, error)
    DeleteEnvironment(id string) error
}
