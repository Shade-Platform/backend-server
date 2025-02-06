package containers

// ContainerService contains business logic related to containers.
type ContainerService struct {
	ContainerRepo ContainerRepository
}

// NewContainerService creates and returns a new instance of ContainerService.
func NewContainerService(repo ContainerRepository) *ContainerService {
	return &ContainerService{ContainerRepo: repo}
}

// CreateContainer handles the creation of a new user.
func (s *ContainerService) CreateContainer(username, containerTag  string, mappedPort int32) (*Container, error) {
	// Create a new container instance
	container := &Container{
		UserName:     username,
		ContainerTag: containerTag,
		MappedPort: mappedPort,
	}

	// Push the container to the cluster
	createdContainer, err := s.ContainerRepo.Create(container)
	if err != nil {
		return nil, err
	}

	return createdContainer, nil
}
