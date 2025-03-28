package service

import (
	"fmt"
	
	"lqkhoi-go-http-api/internal/dto"
	"lqkhoi-go-http-api/internal/models"
	"lqkhoi-go-http-api/internal/repository"
)

type ProjectService interface {
	CreateProject(project *models.Project) (*models.Project, error)
	ListProjects(filter dto.ProjectFilter) ([]models.Project, error)
}

type projectService struct {
	projectRepository repository.ProjectRepository
}

func NewProjectService(projectRepository repository.ProjectRepository) ProjectService {
	return &projectService{
		projectRepository: projectRepository,
	}
}

func (s *projectService) CreateProject(project *models.Project) (*models.Project, error) {
	return s.projectRepository.Create(project)
}

func (s *projectService) ListProjects(filter dto.ProjectFilter) ([]models.Project, error) {
	projects, err := s.projectRepository.Find(filter)
	if err != nil {
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	return projects, nil
}
