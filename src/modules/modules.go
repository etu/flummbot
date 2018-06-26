package modules

type Module interface {
	DbSetup()
	RegisterCallbacks()
}
