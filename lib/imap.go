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

		seqset := new(imap.SeqSet)
		if m.Config.Profiles[profile].Filter.Unseen {
			criteria := imap.NewSearchCriteria()
			criteria.WithoutFlags = []string{imap.SeenFlag}
			ids, err := c.Search(criteria)
			if len(ids) == 0 {
				continue
			}
			m.Info("unseen mails", map[string]interface{}{"ids": ids})
			if err != nil {
				return []Mail{}, err
			}
			seqset.AddNum(ids...)
		} else {
			seqset.AddRange(uint32(1), mbox.Messages)
		}

		// nothing to do here
		if seqset.Empty() {
			continue
		}

		m.Info("processing mails", map[string]interface{}{
			"folder": folder,
		})
		messages := make(chan *imap.Message, 100)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, "BODY[]", imap.FetchUid}, messages)
		}()

		for msg := range messages {
			r := msg.GetBody(&imap.BodySectionName{})

			mr, err := m.read(r)
			if err != nil {
				m.Error("Read Error", map[string]interface{}{"Error": err})
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
					"subject": msg.Envelope.Subject, "message-id": email.ID,
				})
				mails = append(mails, email)
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
