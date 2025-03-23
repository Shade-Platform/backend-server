package containers

import "fmt"

// ContainerService contains business logic related to containers.
type ContainerService struct {
	ContainerRepo ContainerRepository
}

// NewContainerService creates and returns a new instance of ContainerService.
func NewContainerService(repo ContainerRepository) *ContainerService {
	return &ContainerService{ContainerRepo: repo}
}

// CreateContainer handles the creation of a new user.
func (s *ContainerService) CreateContainer(
	id,
	username,
	containerTag string,
	replicas,
	mappedPort int32,
) (*Container, error) {

	// Ensure that container does not end in -service
	if len(id) >= 8 && id[len(id)-8:] == "-service" {
		return nil, fmt.Errorf("container name cannot end in -service")
	}

	// Ensure that id length greater than 3 characters and is less than 64 - 8 = 56 characters
	// (so that service name can be less than 64 characters)
	if len(id) < 3 {
		return nil, fmt.Errorf("container name must be at least 3 characters")
	}
	if len(id) > 56 {
		return nil, fmt.Errorf("container name cannot be longer than 56 characters")
	}

	// Create a new container instance
	container := &Container{
		Name:       id,
		Owner:      username,
		ImageTag:   containerTag,
		Replicas:   replicas,
		MappedPort: mappedPort,
	}

	// Push the container to the cluster
	createdContainer, err := s.ContainerRepo.Create(container)
	if err != nil {
		return nil, err
	}

	return createdContainer, nil
}

func (s *ContainerService) GetContainerStatus(user, name string) (*Container, error) {
	container, err := s.ContainerRepo.GetByName(user, name)
	if err != nil {
		return nil, err
	}

	return container, nil
}

func (s *ContainerService) DeleteContainer(user, name string) error {
	return s.ContainerRepo.Delete(user, name)
}
