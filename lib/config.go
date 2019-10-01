package mail2most

import (
	"os"

	"github.com/BurntSushi/toml"
)

type config struct {
	General  general
	Logging  logging
	Profiles []profile `toml:"Profile"`
}
type general struct {
	File         string
	TimeInterval uint
}

type logging struct {
	Loglevel string
	Logtype  string
	Logfile  string
	Output   string
}

type profile struct {
	Mail       maildata
	Mattermost mattermost
	Filter     filter
}

type maildata struct {
	ImapServer, Username, Password string
	ReadOnly                       bool
	ImapTLS                        bool
}

type filter struct {
	Folders, From, To, Subject []string
	Unseen                     bool
	TimeRange                  string
}

type mattermost struct {
	URL, Team, Username, Password string
	Channels                      []string
	Broadcast                     []string
	SubjectOnly                   bool
	StripHTML                     bool
}

func parseConfig(fileName string, conf *config) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return err
	}
	_, err := toml.DecodeFile(fileName, conf)
	return err
}
