package flummbot

type Config struct {
	Connection struct {
		Channel          string
		Nick             string
		Server           string
		NickservIdentify string
	}
}
