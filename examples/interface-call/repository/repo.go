package repository

type MemoryUserRepository struct{}

func NewMemoryUserRepository() *MemoryUserRepository {
	return &MemoryUserRepository{}
}

func (r *MemoryUserRepository) Save(name string) error {
	_ = name
	return nil
}
