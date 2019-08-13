package mail2most

import (
	"encoding/json"
	"io/ioutil"
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

		bv, err := ioutil.ReadAll(jsonFile)
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
	for k, v := range alreadySendFile {
		alreadySend[k] = v
	}

	// set a 10 seconds sleep default if no TimeInterval is defined
	if m.Config.General.TimeInterval == 0 {
		m.Debug("no check time interval set", map[string]interface{}{
			"fallback":     10,
			"unit-of-time": "second",
		})
		m.Config.General.TimeInterval = 10
	}

	for {
		for p := range m.Config.Profiles {
			mails, err := m.GetMail(p)
			if err != nil {
				return err
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
						m.Error("Mattermost Error", map[string]interface{}{
							"Error": err,
						})
					}
					alreadySend[p] = append(alreadySend[p], mail.ID)
					err = writeToFile(alreadySend, m.Config.General.File)

					if err != nil {
						return err
					}
				}
			}
		}
		//time.Sleep(time.Duration(m.Config.General.TimeInterval) * 10 * time.Second)
		m.Debug("sleeping", map[string]interface{}{
			"intervaltime": m.Config.General.TimeInterval,
			"unit-of-time": "second",
		})
		time.Sleep(time.Duration(m.Config.General.TimeInterval) * time.Second)
	}

}
