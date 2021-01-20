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

func main() {
	configFile := flag.String("conf", "", "path to octavia config file")
	flag.Parse()

	if *configFile == "" {
		flag.PrintDefaults()
		os.Exit(1)
	}

	err, url := config.Get(*configFile)
	if err != nil {
		log.Fatal(err)
	}

	err, db := database.Connect(url.Database)
	if err != nil {
		logger.Debug(err)
		return
	}

	err, amqp := rabbit.Connect(url.Rabbit)
	if err != nil {
		logger.Debug(err)
		return
	}
	server.Run(amqp, db)
}
