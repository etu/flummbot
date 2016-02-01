package flummbot

type Config struct {
	Connection struct {
		Channel          string
		Nick             string
		Server           string
		NickservIdentify string
	}

	Tells struct {
		Command string
	}

	Quotes struct {
		Command string
	}
}
