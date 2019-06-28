package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"os"
	"time"

	m2m "github.com/cseeger-epages/mail2most/library"
)

const datafile = "data.json"

func main() {
	confFile := flag.String("c", "conf/mail2most.conf", "path to config file")
	flag.Parse()

	m, err := m2m.New(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	alreadySend := make([][]uint32, len(m.Config.Profiles))
	if _, err := os.Stat(datafile); err == nil {
		jsonFile, err := os.Open(datafile)
		if err != nil {
			log.Fatal(err)
		}

		bv, err := ioutil.ReadAll(jsonFile)
		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(bv, &alreadySend)
		if err != nil {
			log.Fatal(err)
		}
	}

	for {
		for p := range m.Config.Profiles {
			mails, err := m.GetMail(p)
			if err != nil {
				log.Fatal(err)
			}

			for _, mail := range mails {
				send := true
				for _, id := range alreadySend[p] {
					if mail.ID == id {
						send = false
					}
				}
				if send {
					err := m.PostMattermost(p, mail)
					if err != nil {
						log.Fatal(err)
					}
					alreadySend[p] = append(alreadySend[p], mail.ID)
					err = writeToFile(alreadySend, datafile)
					if err != nil {
						log.Fatal(err)
					}
				}
			}
		}
		time.Sleep(10 * time.Second)
	}
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
