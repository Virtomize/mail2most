package mail2most

import (
	"encoding/json"
	"fmt"
	"image"
	"io"
	"os"
	"reflect"
	"strings"
	"time"

	// image extensions
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	gomessage "github.com/emersion/go-message"
	"github.com/emersion/go-message/charset"
	gomail "github.com/emersion/go-message/mail"
)

func init() {
	// already downloaded attachments dont need to be replaced
	seenAttachments = make(map[[32]byte]string)
	// additional charsets are defined in charsets.go
	for name, chst := range charsets {
		charset.RegisterEncoding(name, chst)
	}
}

// New creates a new Mail2Most object
func New(confPath string) (Mail2Most, error) {
	var conf config
	err := parseConfig(confPath, &conf)
	if err != nil {
		return Mail2Most{}, err
	}

	for k, p := range conf.Profiles {
		if !p.IgnoreDefaults {
			// create a default profile and overwrite what is defined in the profile
			prof := conf.DefaultProfile
			voft := reflect.ValueOf(&prof).Elem()
			vof := reflect.ValueOf(p)
			for i := 0; i < voft.NumField(); i++ {
				for j := 0; j < vof.NumField(); j++ {
					if vof.Field(j).Type().Kind() == reflect.Struct {
						if voft.Type().Field(i).Name == vof.Type().Field(j).Name {
							for k := 0; k < vof.Field(j).NumField(); k++ {
								if !vof.Field(j).Field(k).IsZero() {
									voft.Field(i).FieldByName(vof.Field(j).Type().Field(k).Name).Set(vof.Field(j).Field(k))
								}
							}
						}
					}
				}
			}
			conf.Profiles[k] = prof
		}
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
	err = os.WriteFile(filename, file, 0600)
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
		m.Error("Read Error", map[string]interface{}{"Error": err, "function": "Mail2Most.read"})
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

// processReader processes a mail.Reader and returns the body and a list of attachment filename paths or an error
func (m Mail2Most) processReader(mr *gomail.Reader, profile int) (string, []Attachment, error) {
	if mr == nil {
		return "", []Attachment{}, fmt.Errorf("nil reader")
	}
	var (
		body        string
		html        string
		text        string
		attachments []Attachment
	)
	// Process each message's part
	for {
		p, err := mr.NextPart()
		if err == io.EOF {
			break
		} else if err != nil {
			if err != nil {
				continue
			}
		}

		m.Debug("InlineHeader type is", map[string]interface{}{"type": p.Header.Get("Content-Type")})
		switch h := p.Header.(type) {
		case *gomail.InlineHeader:

			// Parse HTML e-mails
			if strings.HasPrefix(p.Header.Get("Content-Type"), "text/html") {

				// This is the message's text (can be plain-text or HTML)
				b, err := io.ReadAll(p.Body)
				if err != nil {
					continue
				}

				b, err = m.parseHTML(b, profile)
				if err != nil {
					m.Error("Parse Error", map[string]interface{}{"error": err, "function": "Mail2Most.paresHTML", "stage": "parse html"})
					continue
				}

				html += string(b)

				// Parse plaintext e-mails
			} else if strings.HasPrefix(p.Header.Get("Content-Type"), "text/plain") {
				// only parse if no html is found
				if len(html) < 1 {
					b, err := io.ReadAll(p.Body)
					if err != nil {
						m.Error("Read Error", map[string]interface{}{"error": err, "function": "ioutil.ReadAll", "stage": "parse plain text"})
						continue
					}
					_, _, err = image.Decode(strings.NewReader(string(b)))
					// images will be ignored
					if err != nil {
						b, err = m.parseText(b)
						if err != nil {
							m.Error("Parse Text Error", map[string]interface{}{"error": err, "function": "Mail2Most.parseText", "stage": "parse plain text"})
							continue
						}
						text += string(b)
					}
				}

				// Parse images
			} else if strings.HasPrefix(p.Header.Get("Content-Type"), "image/") {
				if m.Config.Profiles[profile].Mattermost.MailAttachments {

					b, err := io.ReadAll(p.Body)
					if err != nil {
						m.Error("Read Error", map[string]interface{}{"error": err, "function": "io.ReadAll", "stage": "parse images"})
						continue
					}

					attachment, err := m.parseAttachment(b, p.Header.Get("Content-Type"))
					if err != nil {
						m.Error("Parse Attachment Error", map[string]interface{}{"error": err, "function": "Mail2Most.parseAttachment", "stage": "parse images"})
						continue
					}
					attachments = append(attachments, attachment)
				}

			} else {
				m.Debug("InlineHeader Unknown", map[string]interface{}{"type": p.Header.Get("Content-Type")})
			}
		case *gomail.AttachmentHeader:
			// This is an attachment
			if m.Config.Profiles[profile].Mattermost.MailAttachments {
				filename, err := h.Filename()
				if err != nil {
					continue
				}

				if filename != "" {
					m.Debug("attachments found", map[string]interface{}{"filename": filename})
				}

				b, err := io.ReadAll(p.Body)
				if err != nil {
					// Skip this attachment and hope things aren't boned.
					m.Error("io returned an error", map[string]interface{}{"error": err})
					continue
				}

				attachment, err := m.parseAttachment(b, fmt.Sprintf("name=\"%s\"", filename))
				if err != nil {
					m.Error("Parse Attachment Error", map[string]interface{}{"error": err, "function": "Mail2Most.parseAttachment", "stage": "parse attachment"})
				}
				attachments = append(attachments, attachment)
			}
		}
		if len(html) > 0 {
			body = html
		} else if len(text) > 0 {
			body = text
		} else {
			body = ""
		}
	}
	return body, attachments, nil
}
