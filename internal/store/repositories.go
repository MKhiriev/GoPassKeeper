package store

type Repositories struct {
	UserRepository UserRepository
}

func NewServices() *Repositories {
	return &Repositories{}
}
