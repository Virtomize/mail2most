package mail2most

import (
	"crypto/tls"
	"strings"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func (m Mail2Most) connect(profile int) (*client.Client, error) {
	var (
		c   *client.Client
		err error
	)
	if m.Config.Profiles[profile].Mail.ImapTLS {
		var tlsconf tls.Config

		if !m.Config.Profiles[profile].Mail.VerifyTLS {
			tlsconf.InsecureSkipVerify = true
		}
		c, err = client.DialTLS(m.Config.Profiles[profile].Mail.ImapServer, &tlsconf)
	} else {
		c, err = client.Dial(m.Config.Profiles[profile].Mail.ImapServer)
	}
	if err != nil {
		return nil, err
	}

	err = c.Login(m.Config.Profiles[profile].Mail.Username, m.Config.Profiles[profile].Mail.Password)
	if err != nil {
		return nil, err
	}
	m.Debug("mailserver", map[string]interface{}{
		"status": "connected",
		"server": m.Config.Profiles[profile].Mail.ImapServer,
	})

	return c, nil
}

// GetMail returns emails filter by profile id
func (m Mail2Most) GetMail(profile int) ([]Mail, error) {

	// Connect to server
	c, err := m.connect(profile)
	if err != nil {
		return []Mail{}, err
	}
	defer c.Logout()

	// Select Folder
	folders := []string{"INBOX"}
	if len(m.Config.Profiles[profile].Filter.Folders) > 0 {
		folders = m.Config.Profiles[profile].Filter.Folders
	}
	m.Debug("checking folders", map[string]interface{}{
		"folders": folders,
	})

	var mails []Mail

	for _, folder := range folders {
		mbox, err := c.Select(folder, m.Config.Profiles[profile].Mail.ReadOnly)
		if err != nil {
			return []Mail{}, err
		}

		m.Info("processing mails", map[string]interface{}{
			"folder": folder,
		})

		limit := m.Config.Profiles[profile].Mail.Limit
		seqset := new(imap.SeqSet)
		if m.Config.Profiles[profile].Filter.Unseen {
			m.Debug("searching unseen", map[string]interface{}{"unseen": m.Config.Profiles[profile].Filter.Unseen})
			criteria := imap.NewSearchCriteria()
			criteria.WithoutFlags = []string{imap.SeenFlag}
			ids, err := c.Search(criteria)
			if len(ids) == 0 {
				m.Debug("no mails found", nil)
				continue
			}
			if err != nil {
				return []Mail{}, err
			}

			// Avoid bucket overflows on ids[0:limit]
			if limit > uint32(len(ids)) {
				limit = uint32(len(ids))
			}

			if limit > 0 {
				m.Info("unseen mails limit found", map[string]interface{}{"ids": ids[0:limit], "limit": limit + 1})
				seqset.AddNum(ids[0:limit]...)
			} else {
				m.Info("unseen mails", map[string]interface{}{"ids": ids})
				seqset.AddNum(ids...)
			}
		} else {
			from := uint32(1)
			if limit > 0 {
				if mbox.Messages > limit {
					from = mbox.Messages - limit
				}
				seqset.AddRange(from, mbox.Messages)
				m.Info("new mails", map[string]interface{}{"from": from, "to": mbox.Messages, "limit": limit + 1})
			} else {
				seqset.AddRange(uint32(1), mbox.Messages)
				m.Info("unseen mails", map[string]interface{}{"from": uint32(1), "to": mbox.Messages, "count": mbox.Messages})
			}
		}

		// nothing to do here
		if seqset.Empty() {
			m.Debug("no mails found", map[string]interface{}{"folder": folder})
			continue
		}

		messages := make(chan *imap.Message, 10000)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, "BODY[]", imap.FetchUid}, messages)
		}()

		for msg := range messages {
			m.Debug("processing message", map[string]interface{}{"uid": msg.Uid, "subject": msg.Envelope.Subject})
			r := msg.GetBody(&imap.BodySectionName{})

			mr, err := m.read(r)
			if err != nil {
				m.Error("Read Error", map[string]interface{}{"Error": err, "function": "Mail2Most.GetMail"})
				return []Mail{}, err
			}

			if mr == nil {
				continue
			}

			body, attachments, err := m.processReader(mr, profile)
			if err != nil {
				m.Error("Read Processing Error", map[string]interface{}{"Error": err})
				return []Mail{}, err
			}

			// Skip empty messages.
			if len(strings.TrimSpace(body)) < 1 && len(attachments) < 1 {
				m.Info("blank message", map[string]interface{}{
					"subject": msg.Envelope.Subject, "uid": msg.Uid,
				})
				continue
			}

			// Skip mailserver error notifications.
			if strings.HasPrefix(msg.Envelope.Subject, "Delivery Status Notification") {
				if m.Config.Profiles[profile].Filter.IgnoreMailErrorNotifications {
					m.Info("skipping mailserver error", map[string]interface{}{
						"subject": msg.Envelope.Subject, "uid": msg.Uid,
					})
					continue
				}
			}

			email := Mail{
				ID:          msg.Uid,
				From:        msg.Envelope.From,
				To:          msg.Envelope.To,
				Subject:     msg.Envelope.Subject,
				Body:        strings.TrimSuffix(body, "\n"),
				Date:        msg.Envelope.Date,
				Attachments: attachments,
			}

			test, err := m.checkFilters(profile, email)
			if err != nil {
				return []Mail{}, err
			}

			if test {
				m.Info("found mail", map[string]interface{}{
					"subject": msg.Envelope.Subject, "uid": msg.Uid,
				})
				mails = append(mails, email)
			} else {
				m.Debug("message not passing the filter", map[string]interface{}{"subject": msg.Envelope.Subject, "uid": msg.Uid})
			}
		}

		if err := <-done; err != nil {
			return []Mail{}, err
		}
	}

	return mails, nil
}

// ListMailBoxes lists all available mailboxes
func (m Mail2Most) ListMailBoxes(profile int) ([]string, error) {

	// Connect to server
	c, err := m.connect(profile)
	if err != nil {
		return []string{}, err
	}
	defer c.Logout()

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	var mboxes []string
	for m := range mailboxes {
		mboxes = append(mboxes, m.Name)
	}

	if err := <-done; err != nil {
		return []string{}, err
	}
	return mboxes, nil
}

// ListFlags lists all flags for profile
func (m Mail2Most) ListFlags(profile int) ([]string, error) {

	// Connect to server
	c, err := m.connect(profile)
	if err != nil {
		return []string{}, err
	}
	defer c.Logout()

	// Select Folder
	folders := []string{"INBOX"}
	if len(m.Config.Profiles[profile].Filter.Folders) > 0 {
		folders = m.Config.Profiles[profile].Filter.Folders
	}
	var flags []string
	for _, folder := range folders {
		mbox, err := c.Select(folder, false)
		if err != nil {
			return []string{}, err
		}

		flags = append(flags, mbox.Flags...)
	}
	return flags, nil
}
