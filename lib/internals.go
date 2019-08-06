package mail2most

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	gomessage "github.com/emersion/go-message"
	"github.com/emersion/go-message/charset"
	gomail "github.com/emersion/go-message/mail"
)

// New creates a new Mail2Most object
func New(confPath string) (Mail2Most, error) {
	var conf config
	err := parseConfig(confPath, &conf)
	if err != nil {
		return Mail2Most{}, err
	}
	m := Mail2Most{Config: conf}
	err = m.initLogger()
	if err != nil {
		return Mail2Most{}, err
	}

	return m, nil
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

func (m Mail2Most) filterByTimeRange(profile int, mail Mail) (bool, error) {
	if m.Config.Profiles[profile].Filter.TimeRange == "" {
		return true, nil
	}
	d, err := time.ParseDuration(m.Config.Profiles[profile].Filter.TimeRange)
	if err != nil {
		return false, err
	}
	now := time.Now()
	notBefore := now.Add(-d)
	if mail.Date.Before(notBefore) {
		return false, nil
	}
	return true, nil
}

func (m Mail2Most) checkFilters(profile int, mail Mail) (bool, error) {
	if m.containsFrom(profile, mail) && m.containsTo(profile, mail) && m.containsSubject(profile, mail) {
		test, err := m.filterByTimeRange(profile, mail)
		if err != nil {
			return false, err
		}
		if test {
			return true, nil
		}
		return false, nil
	}
	return false, nil

}

func writeToFile(data [][]uint32, filename string) error {
	file, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(filename, file, 0644)
	if err != nil {
		return err
	}
	return nil
}

// read returns a mail.Reader if the charset is correct or convertable
// else it will return an error if something goes very wrong
// if the charset is not convertable it just returns nil, nil
func (m Mail2Most) read(r io.Reader) (*gomail.Reader, error) {
	if r == nil {
		return nil, fmt.Errorf("nil reader")
	}
	// fix charset errors
	var charSetError bool
	e, err := gomessage.Read(r)
	if gomessage.IsUnknownCharset(err) {
		m.Debug("Charset Error", map[string]interface{}{"Error": err, "status": "trying to convert"})
		charSetError = true
	} else if err != nil {
		m.Error("Read Error", map[string]interface{}{"Error": err})
		if err != nil {
			return nil, err
		}
	}
	mr := gomail.NewReader(e)
	if charSetError {
		_, params, err := mr.Header.ContentType()
		if err != nil {
			return nil, err
		}
		newr, err := charset.Reader(params["charset"], r)
		if err != nil {
			m.Error("Charset Error", map[string]interface{}{"Error": err, "status": "could not convert"})
			return nil, nil
		}
		e, err = gomessage.Read(newr)
		if err != nil {
			return nil, err
		}
		mr = gomail.NewReader(e)
	}
	return mr, nil
}

// processReader processes a mail.Reader and returns the body or an error
func (m Mail2Most) processReader(mr *gomail.Reader) (string, error) {
	if mr == nil {
		return "", fmt.Errorf("nil reader")
	}
	var body string
	// Process each message's part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			if err != nil {
				return "", err
			}
		}

		switch h := p.Header.(type) {
		case *gomail.InlineHeader:
			// This is the message's text (can be plain-text or HTML)
			b, err := ioutil.ReadAll(p.Body)
			if err != nil {
				return "", err
			}
			body = string(b)
		case *gomail.AttachmentHeader:
			// This is an attachment
			filename, err := h.Filename()
			if err != nil {
				return "", err
			}
			if filename != "" {
				m.Debug("attachments found", map[string]interface{}{"filename": filename})
			}
		}
	}
	return body, nil
}
