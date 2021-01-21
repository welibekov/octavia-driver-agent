package main

import (
	"octavia-driver-agent/database"
	"octavia-driver-agent/rabbit"
	"octavia-driver-agent/logger"
	"octavia-driver-agent/server"
	"octavia-driver-agent/config"
	"flag"
	"log"
	"os"
)

var ConfigFile string

func main() {
	err, url := config.Get(ConfigFile)
	if err != nil {
		log.Fatal(err)
	}

	err, db := database.Connect(url.Database)
	if err != nil {
		logger.Debug(err)
		return
	}
	database.Database = db

	err, amqp := rabbit.Connect(url.Rabbit)
	if err != nil {
		logger.Debug(err)
		return
	}
	server.Run(amqp)
}

func init() {
	flag.StringVar(&ConfigFile,"conf", "", "path to octavia config file")
	flag.StringVar(&logger.LogFile,"log", "", "path to octavia log file")
	flag.Parse()

	if ConfigFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	} else if logger.LogFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}
}
