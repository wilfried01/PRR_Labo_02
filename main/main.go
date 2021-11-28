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

	//Modify the filepath to adapt to your environment
	file, err1 := os.Open("main/configuration.json")
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
		debug := false

	args := os.Args[1:]
	if len(args) == 1 {
		arg := args[0]

		if arg == "DEBUG" {
			fmt.Println("Servers are running on debug mode!")
			debug = true
		}
	}

	var servers = make([]*server.Server, configFile.ServerNumber)
	for i := configFile.ServerNumber; i > 0; i-- {
		servers[i-1] = server.NewServer(i,debug)
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
}
