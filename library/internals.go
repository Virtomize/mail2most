package mail2most

import (
	"fmt"
	"strings"
)

// New creates a new Mail2Most object
func New(confPath string) (Mail2Most, error) {
	var conf config
	err := parseConfig(confPath, &conf)
	if err != nil {
		return Mail2Most{}, err
	}
	return Mail2Most{Config: conf}, nil
}

func (m Mail2Most) containsFrom(profile int, mail Mail) bool {
	if len(m.Config.Profiles[profile].Filter.From) == 0 {
		return true
	}
	for _, from := range m.Config.Profiles[profile].Filter.From {
		for _, mailfrom := range mail.From {
			test := fmt.Sprintf("%s@%s", mailfrom.MailboxName, mailfrom.HostName)
			if strings.Contains(test, from) {
				return true
			}
		}
	}
	return false
}

func (m Mail2Most) containsTo(profile int, mail Mail) bool {
	if len(m.Config.Profiles[profile].Filter.To) == 0 {
		return true
	}
	for _, to := range m.Config.Profiles[profile].Filter.To {
		for _, mailto := range mail.To {
			test := fmt.Sprintf("%s@%s", mailto.MailboxName, mailto.HostName)
			if strings.Contains(test, to) {
				return true
			}
		}
	}
	return false
}

func (m Mail2Most) containsSubject(profile int, mail Mail) bool {
	if len(m.Config.Profiles[profile].Filter.Subject) == 0 {
		return true
	}
	for _, subj := range m.Config.Profiles[profile].Filter.Subject {
		if strings.Contains(mail.Subject, subj) {
			return true
		}
	}
	return false
}

func (m Mail2Most) checkFilters(profile int, mail Mail) bool {
	if m.containsFrom(profile, mail) && m.containsTo(profile, mail) && m.containsSubject(profile, mail) {
		return true
	}
	return false

}
