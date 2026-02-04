package service

type Services struct {
	AuthService AuthService
}

func NewServices() *Services {
	return &Services{}
}
