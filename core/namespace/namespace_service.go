package namespace

type NamespaceService struct {
	NamespaceRepo NamespaceRepository
}

func NewNamespaceService(repo NamespaceRepository) *NamespaceService {
	return &NamespaceService{NamespaceRepo: repo}
}

func (s NamespaceService) CreateNamespace(name string) error {
	return s.NamespaceRepo.CreateNamespace(name)
}

func (s NamespaceService) Exists(name string) bool {
	exists, err := s.NamespaceRepo.Exists(name)
	if err != nil {
		return false
	}
	return exists
}
