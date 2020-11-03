package main

import (
	"flag"
	"log"
	"os"

	"github.com/uzhinskiy/extractor/modules/config"
	"github.com/uzhinskiy/extractor/modules/router"
)

var (
	configfile string
	vBuild     string
	cnf        config.Config
	hostname   string
)

func init() {
	flag.StringVar(&configfile, "config", "main.yml", "Read configuration from this file")
	flag.StringVar(&configfile, "f", "main.yml", "Read configuration from this file")
	version := flag.Bool("V", false, "Show version")
	flag.Parse()
	if *version {
		print("Build num: ", vBuild, "\n")
		os.Exit(0)
	}

	hostname, _ = os.Hostname()
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.SetPrefix(hostname + "\t")

	log.Println("Bootstrap: build num.", vBuild)

	cnf = config.Parse(configfile)
	log.Println("Bootstrap: successful parsing config file. Items: ", cnf)
}

func main() {
	router.Run(cnf)
}
