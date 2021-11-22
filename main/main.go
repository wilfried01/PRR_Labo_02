package main

import (
	"configuration"
	"encoding/json"
	"log"
	"os"
	"server"
	"time"
)

func main() {
	file, err1 := os.Open("server/configuration.json")
	defer file.Close()
	if err1 != nil {
		log.Fatal(err1)
	}

	decoder := json.NewDecoder(file)
	configFile := configuration.Configuration{}
	err := decoder.Decode(&configFile)
	if err != nil {
		log.Fatal(err)
		return
	}
	var servers = make([]*server.Server, configFile.ServerNumber)
	for i := configFile.ServerNumber; i > 0; i-- {
		servers[i-1] = server.NewServer(i)
	}
	for {
		time.Sleep(time.Second * 1)
		if servers[configFile.ServerNumber-1].GetAvailable() {
			break
		}
	}

	for {
		time.Sleep(time.Second * 10)
	}
	/*
		for i:=0; i < configuration.ServerNumber; i++ {
			newServer.ConnectToOthers(configuration)
		}

	*/

}
