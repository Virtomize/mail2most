package mail2most

import (
	"os"

	"github.com/BurntSushi/toml"
)

type config struct {
	General        general
	Logging        logging
	Profiles       []profile `toml:"Profile"`
	DefaultProfile profile
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
	IgnoreDefaults bool
	Mail           maildata
	Mattermost     mattermost
	Filter         filter
}

type maildata struct {
	ImapServer, Username, Password string
	ReadOnly                       bool
	ImapTLS                        bool
	VerifyTLS                      bool
	Limit                          uint32
}

type filter struct {
	Folders, From, To, Subject []string
	Unseen                     bool
	TimeRange                  string
}

type mattermost struct {
	URL, Team, Username, Password, AccessToken string
	Channels                                   []string
	Users                                      []string
	Broadcast                                  []string
	SubjectOnly                                bool
	StripHTML                                  bool
	HideFrom                                   bool
	MailAttachments                            bool
}

func parseConfig(fileName string, conf *config) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return err
	}
	_, err := toml.DecodeFile(fileName, conf)
	return err
}
