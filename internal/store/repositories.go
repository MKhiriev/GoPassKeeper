package store

type Repositories struct {
	UserRepository     UserRepository
	PrivateDataStorage PrivateDataStorage
}

func NewServices() *Repositories {
	return &Repositories{}
}
