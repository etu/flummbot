package modules

import (
	"github.com/etu/flummbot/src/irc"
)

type Module interface {
	RegisterCallbacks(*irc.IrcConnection)
}
