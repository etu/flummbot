package main

type Config struct {
	Connection struct {
		Channels         []string
		Nick             string
		Server           string
		NickservIdentify string
		TLS              bool
		Message          string
	}

	Database struct {
		File string
	}

	Tells struct {
		Command         string
		AllowedChannels []string
	}

	Quotes struct {
		Command         string
		AllowedChannels []string
	}

	Invite struct {
		Message   string
		Whitelist []string
	}

	Karma struct {
		PlusOperator    string
		MinusOperator   string
		Command         string
		AllowedChannels []string
	}
}
