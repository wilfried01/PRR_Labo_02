package main

import (
	"encoding/json"
	"log"
	"os"
	"server"
	"time"
)

func main() {
	file, err1 := os.Open("configuration.json")
	defer file.Close()
	if err1 != nil {
		log.Fatal(err1)
	}

	decoder := json.NewDecoder(file)
	configuration := server.Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
		return
	}
	var servers []*server.Server = make([]*server.Server, configuration.ServerNumber)
	for i := configuration.ServerNumber; i > 0; i-- {
		servers[i-1] = server.NewServer(i)
	}
	for {
		time.Sleep(time.Second * 1)
		if servers[configuration.ServerNumber-1].Available {
			break
		}
	}

	go servers[0].AskSC()
	go servers[1].AskSC()
	go servers[2].AskSC()
	go servers[3].AskSC()
	go servers[4].AskSC()

	for {
		time.Sleep(time.Second * 10)
	}
	/*
		for i:=0; i < configuration.ServerNumber; i++ {
			newServer.ConnectToOthers(configuration)
		}

	*/

}
