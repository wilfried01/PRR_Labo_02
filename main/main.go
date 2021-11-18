package main

import (
	"encoding/json"
	"fmt"
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
	var newServer *server.Server
	for i := configuration.ServerNumber; i > 0; i-- {
		newServer = server.NewServer(i)
	}
	for {
		time.Sleep(time.Second * 1)
		if newServer.Available {
			break
		}
	}
	fmt.Fprintf(newServer.OutConnections[1], "REQ 10 3\n")

	for {
		time.Sleep(time.Second * 10)
	}
	/*
		for i:=0; i < configuration.ServerNumber; i++ {
			newServer.ConnectToOthers(configuration)
		}

	*/

}
