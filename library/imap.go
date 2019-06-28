package mail2most

import (
	"bytes"
	"log"
	"net/mail"
	"strings"

	imap "github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

func (m Mail2Most) connect(profile int) (*client.Client, error) {
	c, err := client.DialTLS(m.Config.Profiles[profile].Mail.ImapServer, nil)
	if err != nil {
		return nil, err
	}

	err = c.Login(m.Config.Profiles[profile].Mail.Username, m.Config.Profiles[profile].Mail.Password)
	if err != nil {
		return nil, err
	}

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
			log.Println("found unseen mail ids", ids)
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

		log.Println("processing mails in", folder)
		messages := make(chan *imap.Message, 100)
		done := make(chan error, 1)
		go func() {
			done <- c.Fetch(seqset, []imap.FetchItem{imap.FetchEnvelope, "BODY[]"}, messages)
		}()

		for msg := range messages {
			r := msg.GetBody(&imap.BodySectionName{})
			body, err := mail.ReadMessage(r)
			if err != nil {
				return []Mail{}, err
			}

			buf := new(bytes.Buffer)
			buf.ReadFrom(body.Body)

			mail := Mail{
				ID:      msg.SeqNum,
				From:    msg.Envelope.From,
				To:      msg.Envelope.To,
				Subject: msg.Envelope.Subject,
				Body:    strings.TrimSuffix(buf.String(), "\n"),
				Date:    msg.Envelope.Date,
			}

			test, err := m.checkFilters(profile, mail)
			if err != nil {
				return []Mail{}, err
			}
			if test {
				log.Println("found mail", msg.Envelope.Subject)
				mails = append(mails, mail)
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
