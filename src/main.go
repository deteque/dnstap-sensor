package main

import (
	"flag"
	"os"
	"io"
	"log"
	"fmt"
	"os/signal"
	"syscall"
)

const (
	CONFIG_FILE = "/etc/dnstap/sensor.conf"
	VERSION = "dnstap-sensor VERSION: 20250724.0"
	KILOBYTE = 1024
	BUFFER_SIZE = 128
)

var Config = ConfigFile{}
var ConfigName string

func main() {
	config_declaration := flag.String("c",CONFIG_FILE,"Location of the configuration file")
	version := flag.Bool("v", false, "Check the version of the program")
	logDeclaration := flag.String("log", "", "Enable logging. Specify a file path or use 'stdout' to log to standard output")
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

	switch *logDeclaration {
	case "": 
		log.SetOutput(io.Discard)
	case "stdout": 
		log.SetOutput(os.Stdout)
	default:
		logfile, err := os.OpenFile(*logDeclaration, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			fmt.Println("Error creating log file: ", *logDeclaration, err)
			os.Exit(1)
		}
		log.SetOutput(logfile)
	}

	signal.Ignore(os.Signal(syscall.SIGHUP))

	createENV()

	if Config.ListenerType == "socket" {
		go checkSocket()
	}
	run()
}
