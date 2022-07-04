package internal

type UselessAddon interface {
	Name() string
	GetMessage(oldMessage string) (string, error)
}
