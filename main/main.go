package main

import (
	"configuration"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"server"
	"time"
)

func main() {
	file, err1 := os.Open("configuration.json")
	debug := false
	defer file.Close()
	if err1 != nil {
		log.Fatal(err1)
	}
	args := os.Args[1:]
	if len(args) == 1 {
		arg := args[0]

		if arg == "DEBUG" {
			fmt.Println("Servers are running on debug mode!")
			debug = true
		}
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
		servers[i-1] = server.NewServer(i, debug, configFile)
	}
	for {
		//time.Sleep(time.Second * 1)
		if servers[configFile.ServerNumber-1].GetAvailable() {
			break
		}
	}

	for {
		time.Sleep(time.Second * 10)
	}
}
