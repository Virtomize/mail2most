package mail2most

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"

	"github.com/mattermost/mattermost-server/model"
	"github.com/microcosm-cc/bluemonday"
)

func (m Mail2Most) mlogin(profile int) (*model.Client4, error) {
	c := model.NewAPIv4Client(m.Config.Profiles[profile].Mattermost.URL)

	_, resp := c.Login(m.Config.Profiles[profile].Mattermost.Username, m.Config.Profiles[profile].Mattermost.Password)
	if resp.Error != nil {
		return nil, resp.Error
	}

	return c, nil
}

// PostMattermost posts a msg to mattermost
func (m Mail2Most) PostMattermost(profile int, mail Mail) error {
	c, err := m.mlogin(profile)
	if err != nil {
		return err
	}
	defer c.Logout()

	for _, channel := range m.Config.Profiles[profile].Mattermost.Channels {

		channelName := strings.ReplaceAll(channel, "#", "")
		channelName = strings.ReplaceAll(channelName, "@", "")

		ch, resp := c.GetChannelByNameForTeamName(channelName, m.Config.Profiles[profile].Mattermost.Team, "")
		if resp.Error != nil {
			return resp.Error
		}

		// check if body is base64 encoded
		var body string
		bb, err := base64.StdEncoding.DecodeString(mail.Body)
		if err != nil {
			body = mail.Body
		} else {
			body = string(bb)
		}

		if m.Config.Profiles[profile].Mattermost.StripHTML {
			blm := bluemonday.StrictPolicy()
			body = blm.Sanitize(body)
			space := regexp.MustCompile(`[\s\p{Zs}]{2,}`)
			body = space.ReplaceAllString(body, " ")

			mail.Subject = blm.Sanitize(mail.Subject)
			// just to be sure
			mail.From[0].PersonalName = blm.Sanitize(mail.From[0].PersonalName)
			mail.From[0].MailboxName = blm.Sanitize(mail.From[0].MailboxName)
			mail.From[0].HostName = blm.Sanitize(mail.From[0].HostName)
		}

		msg := fmt.Sprintf(
			":email: _From: **<%s> %s@%s**_\n>_%s_\n\n```\n%s```\n",
			mail.From[0].PersonalName,
			mail.From[0].MailboxName,
			mail.From[0].HostName,
			mail.Subject,
			body,
		)
		if m.Config.Profiles[profile].Mattermost.SubjectOnly {
			msg = fmt.Sprintf(
				":email: _From: **<%s> %s@%s**_\n>_%s_\n\n",
				mail.From[0].PersonalName,
				mail.From[0].MailboxName,
				mail.From[0].HostName,
				mail.Subject,
			)
		}

		// max message length is about 16383
		// https://docs.mattermost.com/administration/important-upgrade-notes.html
		if len(msg) > 16383 {
			msg = msg[0:16382]
		}

		post := &model.Post{ChannelId: ch.Id, Message: msg}
		_, resp = c.CreatePost(post)
		if resp.Error != nil {
			m.Error("Mattermost Post Error", map[string]interface{}{"Error": err, "status": "fallback send only subject"})
			msg = fmt.Sprintf(
				":email: _From: **<%s> %s@%s**_\n>_%s_\n\n",
				mail.From[0].PersonalName,
				mail.From[0].MailboxName,
				mail.From[0].HostName,
				mail.Subject,
			)
			post := &model.Post{ChannelId: ch.Id, Message: msg}
			_, resp = c.CreatePost(post)
			if resp.Error != nil {
				m.Error("Mattermost Post Error", map[string]interface{}{"Error": err, "status": "fallback not working"})
				return resp.Error
			}
		}
	}

	return nil
}
