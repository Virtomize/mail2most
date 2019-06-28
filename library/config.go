package mail2most

import (
	"os"

	"github.com/BurntSushi/toml"
)

type config struct {
	Profiles []profile `toml:"Profile"`
}

type profile struct {
	Mail       maildata
	Filter     filter
	Mattermost mattermost
}

type maildata struct {
	ImapServer, Username, Password string
	ReadOnly                       bool
}

type filter struct {
	Folders, From, To, Subject []string
	Unseen                     bool
}

type mattermost struct {
	URL, Team, Username, Password string
	Channels                      []string
}

func parseConfig(fileName string, conf *config) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return err
	}
	_, err := toml.DecodeFile(fileName, conf)
	return err
}
