package main

import (
	"flag"
	"fmt"
	"log"

	m2m "github.com/virtomize/mail2most/lib"
)

var Version string

func main() {
	confFile := flag.String("c", "conf/mail2most.conf", "path to config file")
	version := flag.Bool("version", false, "display mail2most version")
	flag.Parse()

	m, err := m2m.New(*confFile)
	if err != nil {
		log.Fatal(err)
	}

	if *version {
		if Version == "" {
			fmt.Print("unknown")
		}
		fmt.Print(Version)
		return
	}

	err = m.Run()
	if err != nil {
		log.Fatal(err)
	}
}
