package mail2most

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/k3a/html2text"
	"github.com/mattermost/mattermost-server/model"
	"github.com/mattn/godown"
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

func (m Mail2Most) getFromLine(profile int, userName string, email string) string {
	// nothing to do here
	if len(userName) < 1 && len(email) < 1 {
		return ""
	}

	if !m.Config.Profiles[profile].Mattermost.HideFromEmail {
		return fmt.Sprintf("_From: **<%s> %s**_",
			userName,
			email,
		)
	}
	return fmt.Sprintf("_From: **%s**_",
		userName,
	)
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

	if m.Config.Profiles[profile].Mattermost.ConvertToMarkdown {
		var b bytes.Buffer
		err := godown.Convert(&b, strings.NewReader(body), nil)
		if err != nil {
			return err
		}
		body = b.String()
	} else if m.Config.Profiles[profile].Mattermost.StripHTML {
		body = html2text.HTML2Text(body)
		mail.Subject = html2text.HTML2Text(mail.Subject)
		mail.From[0].PersonalName = html2text.HTML2Text(mail.From[0].PersonalName)
		mail.From[0].MailboxName = html2text.HTML2Text(mail.From[0].MailboxName)
		mail.From[0].HostName = html2text.HTML2Text(mail.From[0].HostName)
	}

	if len(strings.TrimSpace(body)) < 1 {
		m.Info("dead body found", map[string]interface{}{"function": "Mail2Most.PostMattermost"})
		return nil
	}

	if m.Config.Profiles[profile].Mattermost.BodyPrefix != "" {
		body = m.Config.Profiles[profile].Mattermost.BodyPrefix + "\n" + body
	}

	if m.Config.Profiles[profile].Mattermost.BodySuffix != "" {
		body = body + "\n" + m.Config.Profiles[profile].Mattermost.BodySuffix
	}

	msg := ":email: "
	var shortmsg string

	if !m.Config.Profiles[profile].Mattermost.HideFrom {
		if len(mail.From[0].PersonalName) < 1 && len(mail.From[0].MailboxName) < 1 && len(mail.From[0].HostName) < 1 {
			// skip this message, it didn't come from anywhere
			return errors.New("Null sender, skipping message")
		}
		email := fmt.Sprintf("%s@%s", mail.From[0].MailboxName, mail.From[0].HostName)
		user, resp := c.GetUserByEmail(email, "")
		if resp.Error != nil {
			m.Debug("user not found in system", map[string]interface{}{"error": resp.Error})
			msg += m.getFromLine(profile, mail.From[0].PersonalName, email)
		} else {
			msg += m.getFromLine(profile, "@"+user.Username, email)
		}
	}

	if m.Config.Profiles[profile].Mattermost.SubjectOnly && m.Config.Profiles[profile].Mattermost.BodyOnly {
		err := fmt.Errorf("config defines SubjectOnly and BodyOnly to be true which exclude each other")
		m.Error("Configuration inconsistency found", map[string]interface{}{"Config.Profile.Mattermost.SubjectOnly": true, "Config.Profile.Mattermost.BodyOnly": true, "error": err})
		return err
	}

	if m.Config.Profiles[profile].Mattermost.SubjectOnly {
		msg += fmt.Sprintf(
			"\n>_%s_\n\n",
			mail.Subject,
		)
		shortmsg = msg
	} else {
		if m.Config.Profiles[profile].Mattermost.BodyOnly {
			mail.Subject = "\n\n\n\n\n"
		} else {
			mail.Subject = fmt.Sprintf("\n>_%s_\n\n", mail.Subject)
		}
		shortmsg = msg
		if m.Config.Profiles[profile].Mattermost.ConvertToMarkdown {
			msg += fmt.Sprintf(
				"%s\n%s\n",
				mail.Subject,
				body,
			)
		} else {
			msg += fmt.Sprintf(
				"%s```\n%s```\n",
				mail.Subject,
				body,
			)
		}
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
		":email: _%s**_\n>_%s_\n\n",
		m.getFromLine(profile, mail.From[0].PersonalName, mail.From[0].MailboxName+"@"+mail.From[0].HostName),
		mail.Subject,
	)

	if len(m.Config.Profiles[profile].Mattermost.Channels) == 0 {
		m.Debug("no channels configured to send to", nil)
	}

	for _, channel := range m.Config.Profiles[profile].Mattermost.Channels {

		channelName := strings.ReplaceAll(channel, "#", "")
		channelName = strings.ReplaceAll(channelName, "@", "")

		ch, resp := c.GetChannelByNameForTeamName(channelName, m.Config.Profiles[profile].Mattermost.Team, "")
		if resp.Error != nil {
			m.Error("Get Channel Error", map[string]interface{}{"error": resp.Error, "function": "GetCChannelByNameForTeamName"})
			return resp.Error
		}

		fileIDs := m.sendAttachments(c, ch.Id, profile, mail)

		err = m.postMsgs(c, fileIDs, ch.Id, msg, shortmsg, fallback, mail)
		if err != nil {
			return err
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

			fileIDs := m.sendAttachments(c, ch.Id, profile, mail)

			err = m.postMsgs(c, fileIDs, ch.Id, msg, shortmsg, fallback, mail)
			if err != nil {
				return err
			}
		}
	} else {
		m.Debug("no users configured to send to", nil)
	}

	return nil
}

func (m Mail2Most) sendAttachments(c *model.Client4, chID string, profile int, mail Mail) map[int][]string {

	fileIDs := make(map[int][]string)
	if m.Config.Profiles[profile].Mattermost.MailAttachments {
		i := 0
		for k, a := range mail.Attachments {
			// https://github.com/Virtomize/mail2most/issues/62
			// mattermost only allows up to 5 attachments per message
			if k != 0 && k%5 == 0 {
				i++
			}
			fileResp, resp := c.UploadFile(a.Content, chID, a.Filename)
			if resp.Error != nil {
				m.Error("Mattermost Upload File Error", map[string]interface{}{"error": resp.Error})
			} else {
				if len(fileResp.FileInfos) != 1 {
					m.Error("Mattermost Upload File Error", map[string]interface{}{"error": resp.Error, "fileinfos": fileResp.FileInfos})
				} else {
					fileIDs[i] = append(fileIDs[i], fileResp.FileInfos[0].Id)
				}
			}
		}
	}

	return fileIDs
}

func (m Mail2Most) postMsgs(c *model.Client4, fileIDs map[int][]string, chID, msg, shortmsg, fallback string, mail Mail) error {

	if len(fileIDs) > 0 {
		for k, files := range fileIDs {
			post := &model.Post{ChannelId: chID, Message: msg}
			if k > 0 {
				post.Message = shortmsg
			}
			if len(files) > 0 {
				post.FileIds = files
			}
			m.Debug("mattermost post", map[string]interface{}{"channel": chID, "subject": mail.Subject, "bytes": len(post.Message)})
			_, resp := c.CreatePost(post)
			if resp.Error != nil {
				m.Error("Mattermost Post Error", map[string]interface{}{"error": resp.Error, "status": "fallback send only subject"})
				post := &model.Post{ChannelId: chID, Message: fallback}
				_, resp = c.CreatePost(post)
				if resp.Error != nil {
					m.Error("Mattermost Post Error", map[string]interface{}{"error": resp.Error, "status": "fallback not working"})
					return resp.Error
				}
			}
		}
	} else {
		post := &model.Post{ChannelId: chID, Message: msg}
		m.Debug("mattermost post", map[string]interface{}{"channel": chID, "subject": mail.Subject, "bytes": len(post.Message)})
		_, resp := c.CreatePost(post)
		if resp.Error != nil {
			m.Error("Mattermost Post Error", map[string]interface{}{"error": resp.Error, "status": "fallback send only subject"})
			post := &model.Post{ChannelId: chID, Message: fallback}
			_, resp = c.CreatePost(post)
			if resp.Error != nil {
				m.Error("Mattermost Post Error", map[string]interface{}{"error": resp.Error, "status": "fallback not working"})
				return resp.Error
			}
		}
	}
	return nil
}
