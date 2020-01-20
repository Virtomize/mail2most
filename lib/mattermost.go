package mail2most

import (
	"encoding/base64"
	"fmt"
	"strings"

	"github.com/k3a/html2text"
	"github.com/mattermost/mattermost-server/model"
)

func (m Mail2Most) mlogin(profile int) (*model.Client4, error) {
	c := model.NewAPIv4Client(m.Config.Profiles[profile].Mattermost.URL)

	if m.Config.Profiles[profile].Mattermost.Username != "" && m.Config.Profiles[profile].Mattermost.Password != "" {
		_, resp := c.Login(m.Config.Profiles[profile].Mattermost.Username, m.Config.Profiles[profile].Mattermost.Password)
		if resp.Error != nil {
			return nil, resp.Error
		}
	} else if m.Config.Profiles[profile].Mattermost.AccessToken != "" {
		c.AuthToken = m.Config.Profiles[profile].Mattermost.AccessToken
		c.AuthType = "BEARER"
		r, err := c.DoApiGet("/users/me", "")
		if err != nil {
			return nil, err
		}
		u := model.UserFromJson(r.Body)
		m.Config.Profiles[profile].Mattermost.Username = u.Email
	} else {
		return nil, fmt.Errorf("no username, password or token is set")
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

	// check if body is base64 encoded
	var body string
	bb, err := base64.StdEncoding.DecodeString(mail.Body)
	if err != nil {
		body = mail.Body
	} else {
		body = string(bb)
	}

	if m.Config.Profiles[profile].Mattermost.StripHTML {
		body = html2text.HTML2Text(body)
		mail.Subject = html2text.HTML2Text(mail.Subject)
		mail.From[0].PersonalName = html2text.HTML2Text(mail.From[0].PersonalName)
		mail.From[0].MailboxName = html2text.HTML2Text(mail.From[0].MailboxName)
		mail.From[0].HostName = html2text.HTML2Text(mail.From[0].HostName)
	}

	msg := ":email: "

	if !m.Config.Profiles[profile].Mattermost.HideFrom {
		msg += fmt.Sprintf("_From: **<%s> %s@%s**_",
			mail.From[0].PersonalName,
			mail.From[0].MailboxName,
			mail.From[0].HostName,
		)
	}

	if m.Config.Profiles[profile].Mattermost.SubjectOnly {
		msg += fmt.Sprintf(
			"\n>_%s_\n\n",
			mail.Subject,
		)
	} else {
		msg += fmt.Sprintf(
			"\n>_%s_\n\n```\n%s```\n",
			mail.Subject,
			body,
		)
	}

	for _, b := range m.Config.Profiles[profile].Mattermost.Broadcast {
		msg = b + " " + msg
	}
	// max message length is about 16383
	// https://docs.mattermost.com/administration/important-upgrade-notes.html
	if len(msg) > 16383 {
		msg = msg[0:16382]
	}

	fallback := fmt.Sprintf(
		":email: _From: **<%s> %s@%s**_\n>_%s_\n\n",
		mail.From[0].PersonalName,
		mail.From[0].MailboxName,
		mail.From[0].HostName,
		mail.Subject,
	)

	for _, channel := range m.Config.Profiles[profile].Mattermost.Channels {

		channelName := strings.ReplaceAll(channel, "#", "")
		channelName = strings.ReplaceAll(channelName, "@", "")

		ch, resp := c.GetChannelByNameForTeamName(channelName, m.Config.Profiles[profile].Mattermost.Team, "")
		if resp.Error != nil {
			return resp.Error
		}

		post := &model.Post{ChannelId: ch.Id, Message: msg}
		_, resp = c.CreatePost(post)
		if resp.Error != nil {
			m.Error("Mattermost Post Error", map[string]interface{}{"Error": resp.Error, "status": "fallback send only subject"})
			post := &model.Post{ChannelId: ch.Id, Message: fallback}
			_, resp = c.CreatePost(post)
			if resp.Error != nil {
				m.Error("Mattermost Post Error", map[string]interface{}{"Error": resp.Error, "status": "fallback not working"})
				return resp.Error
			}
		}
	}

	if len(m.Config.Profiles[profile].Mattermost.Users) > 0 {

		var (
			me   *model.User
			resp *model.Response
		)

		// who am i
		// user is defined by its email address
		if strings.Contains(m.Config.Profiles[profile].Mattermost.Username, "@") {
			me, resp = c.GetUserByEmail(m.Config.Profiles[profile].Mattermost.Username, "")
			if resp.Error != nil {
				return resp.Error
			}
		} else {
			me, resp = c.GetUserByEmail(m.Config.Profiles[profile].Mattermost.Username, "")
			if resp.Error != nil {
				return resp.Error
			}
		}
		myid := me.Id

		for _, user := range m.Config.Profiles[profile].Mattermost.Users {
			var (
				u *model.User
			)
			// user is defined by its email address
			if strings.Contains(user, "@") {
				u, resp = c.GetUserByEmail(user, "")
				if resp.Error != nil {
					return resp.Error
				}
			} else {
				u, resp = c.GetUserByUsername(user, "")
				if resp.Error != nil {
					return resp.Error
				}
			}

			ch, resp := c.CreateDirectChannel(myid, u.Id)
			if resp.Error != nil {
				return resp.Error
			}

			post := &model.Post{ChannelId: ch.Id, Message: msg}
			_, resp = c.CreatePost(post)
			if resp.Error != nil {
				m.Error("Mattermost Post Error", map[string]interface{}{"Error": err, "status": "fallback send only subject"})
				post := &model.Post{ChannelId: ch.Id, Message: fallback}
				_, resp = c.CreatePost(post)
				if resp.Error != nil {
					m.Error("Mattermost Post Error", map[string]interface{}{"Error": err, "status": "fallback not working"})
					return resp.Error
				}
			}
		}
	}

	return nil
}
