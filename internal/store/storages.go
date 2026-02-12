package store

type Storages struct {
	UserRepository     UserRepository
	PrivateDataStorage PrivateDataStorage
}

func NewServices() *Storages {
	return &Storages{}
}
