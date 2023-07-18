package main

import (
	"flag"
	"os"
	"fmt"
	"os/signal"
	"syscall"
)

const (
	CONFIG_FILE = "/etc/dnstap/sensor.conf"
	VERSION = "dnstap-sensor VERSION: 20230717.1"
	KILOBYTE = 1024
	BUFFER_SIZE = 128
)

var Config = ConfigFile{}
var ConfigName string

func main() {
	config_declaration := flag.String("c",CONFIG_FILE,"Location of the configuration file")
	version := flag.Bool("v", false, "Check the version of the program")
	flag.Parse()

	if *version {
		fmt.Println(VERSION)
		os.Exit(0)
	}

	if *config_declaration == "" {
		ConfigName = CONFIG_FILE
	} else {
		ConfigName = *config_declaration
	}
	signal.Ignore(os.Signal(syscall.SIGHUP))

	createENV()

	go checkSocket()
	run()
}
