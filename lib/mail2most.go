package mail2most

import (
	"encoding/json"
	"io"
	"os"
	"time"
)

// Run starts mail2most
func (m Mail2Most) Run() error {
	alreadySend := make([][]uint32, len(m.Config.Profiles))
	alreadySendFile := make([][]uint32, len(m.Config.Profiles))
	if _, err := os.Stat(m.Config.General.File); err == nil {
		jsonFile, err := os.Open(m.Config.General.File)
		if err != nil {
			return err
		}

		bv, err := io.ReadAll(jsonFile)
		if err != nil {
			return err
		}

		err = json.Unmarshal(bv, &alreadySendFile)
		if err != nil {
			return err
		}
	}

	// write cache to memory cache
	// this is nessasary if new profiles are added
	// and the caching file does not contain any caching
	// for this profile
	l := len(alreadySend)
	for k, v := range alreadySendFile {
		if k < l {
			alreadySend[k] = v
		}
		if k >= l {
			m.Error("data.json error", map[string]interface{}{
				"error":    "data.json contains more profile information than defined in the config",
				"cause":    "this happens if profiles are deleted from the config file and can create inconsistencies",
				"solution": "delete the data.json file",
				"note":     "by deleting the data.json file all mails are parsed and send again",
			})
		}
	}

	// set a 10 seconds sleep default if no TimeInterval is defined
	if m.Config.General.TimeInterval == 0 {
		m.Info("no check time interval set", map[string]interface{}{
			"fallback":     10,
			"unit-of-time": "second",
		})
		m.Config.General.TimeInterval = 10
	}

	for {
		for p := range m.Config.Profiles {
			mails, err := m.GetMail(p)
			if err != nil {
				m.Error("Error reaching mailserver", map[string]interface{}{
					"Error":  err,
					"Server": m.Config.Profiles[p].Mail.ImapServer,
				})
				break
			}

			for _, mail := range mails {
				send := true
				for _, id := range alreadySend[p] {
					if mail.ID == id {
						m.Debug("mail", map[string]interface{}{
							"subject":    mail.Subject,
							"status":     "already send",
							"message-id": mail.ID,
						})
						send = false
					}
				}
				if send {
					err := m.PostMattermost(p, mail)
					if err != nil {
						m.Error("Right after PostMattermost, Mattermost Error. Email not set as synced in mattermost.", map[string]interface{}{
							"Error": err,
						})
					} else {
						alreadySend[p] = append(alreadySend[p], mail.ID)
						m.Debug("In mail2most Run, Before writeToFile on " + m.Config.General.File,nil)
						err = writeToFile(alreadySend, m.Config.General.File)
											
						if err != nil {
							m.Error("File Error", map[string]interface{}{
								"Error": err,
							})
						}
					}
				}
			}
		}

		if !m.Config.General.RunAsService {
			m.Info("done", map[string]interface{}{"Config.General.RunAsService": false, "status": "configured to run only once"})
			break
		}

		//time.Sleep(time.Duration(m.Config.General.TimeInterval) * 10 * time.Second)
		m.Debug("sleeping", map[string]interface{}{
			"intervaltime": m.Config.General.TimeInterval,
			"unit-of-time": "second",
		})
		time.Sleep(time.Duration(m.Config.General.TimeInterval) * time.Second)
	}

	return nil
}
