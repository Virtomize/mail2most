package main

import (
	"flag"
	"log"

	m2m "github.com/virtomize/mail2most/lib"
)

func main() {
	confFile := flag.String("c", "conf/mail2most.conf", "path to config file")
	flag.Parse()

	m, err := m2m.New(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	err = m.Run()
	if err != nil {
		log.Fatal(err)
	}
}
