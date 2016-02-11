package flummbot

type Config struct {
	Connection struct {
		Channel          string
		Nick             string
		Server           string
		NickservIdentify string
		TLS              bool
	}

	Database struct {
		File string
	}

	Tells struct {
		Command string
	}

	Quotes struct {
		Command string
	}

	Invite struct {
		Message   string
		Whitelist []string
	}
}
