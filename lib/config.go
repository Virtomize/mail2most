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
	RunAsService bool
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
	StartTLS                       bool
	VerifyTLS                      bool
	Limit                          uint32
	GenerateLocalUIDs              bool
}

type filter struct {
	Folders, From, To, Subject   []string
	Unseen                       bool
	TimeRange                    string
	IgnoreMailErrorNotifications bool
}

type mattermost struct {
	URL, Team, Username, Password, AccessToken, UserId string
	Channels                                   []string
	Users                                      []string
	Broadcast                                  []string
	SubjectOnly                                bool
	BodyOnly                                   bool
	SkipEmptyMessages                          bool
	StripHTML                                  bool
	ConvertToMarkdown                          bool
	HideFrom                                   bool
	HideFromEmail                              bool
	MailAttachments                            bool
	BodyPrefix, BodySuffix                     string
}

func parseConfig(fileName string, conf *config) error {
	if _, err := os.Stat(fileName); os.IsNotExist(err) {
		return err
	}
	_, err := toml.DecodeFile(fileName, conf)
	return err
}
