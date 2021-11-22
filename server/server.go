package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	serverNumber int
	tcpListener  net.Listener
	//TODO: put this private, used right now for debugging purposes
	OutConnections  []net.Conn
	InConnections   []net.Conn
	Available       bool
	stamp           int
	internalChanIn  chan string
	internalChanOut chan string
	scChan          chan bool
	lamportArray    []LamportState
	Config          Configuration
}

//Server number starts at 1
func NewServer(serverNumber int) *Server {
	number := serverNumber - 1
	server := Server{serverNumber: number, Available: false}
	server.stamp = 0
	//Read config
	file, _ := os.Open("configuration.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	server.Config = configuration

	//Create internal variables
	server.OutConnections = make([]net.Conn, configuration.ServerNumber)
	server.InConnections = make([]net.Conn, configuration.ServerNumber)
	server.internalChanIn = make(chan string)
	server.internalChanOut = make(chan string)
	server.scChan = make(chan bool)
	server.lamportArray = make([]LamportState, configuration.ServerNumber)
	for i := 0; i < configuration.ServerNumber; i++ {
		server.lamportArray[i] = LamportState{State: REL, Stamp: 0}
	}

	server.tcpListener, _ = net.Listen("tcp", configuration.Ips[number])
	go server.handleInternalMessages()
	go server.StartListening()
	go server.ConnectToOthers()

	return &server
}

func (s *Server) StartListening() {
	for {
		//TODO: Handle errors
		//TODO: Add defer
		conn, _ := s.tcpListener.Accept()
		input, _ := bufio.NewReader(conn).ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		tokens := strings.Fields(input)
		if tokens[0] == "HELLO" {
			inNumber, _ := strconv.Atoi(tokens[1])
			s.InConnections[inNumber] = conn
			go s.handleLamport(conn)
		} else {

		}
		//fmt.Println(input)
	}
}

func (s *Server) handleInternalMessages() {
	for {
		//Format CMD Params
		input := <-s.internalChanIn
		tokens := strings.Fields(input)
		switch tokens[0] {
		case "ACK":
			externalStamp, _ := strconv.Atoi(tokens[1])
			externalServerNumber, _ := strconv.Atoi(tokens[2])
			if s.lamportArray[externalServerNumber].State != REQ {
				s.lamportArray[externalServerNumber] = LamportState{
					State: ACK,
					Stamp: externalStamp,
				}
			}
			s.stamp = maxOf(s.stamp, externalStamp)
			s.stamp += 1
			s.internalChanOut <- strconv.Itoa(s.stamp)
		case "REQ":
			externalStamp, _ := strconv.Atoi(tokens[1])
			externalServerNumber, _ := strconv.Atoi(tokens[2])
			s.lamportArray[externalServerNumber] = LamportState{
				State: REQ,
				Stamp: externalStamp,
			}
			s.stamp = maxOf(s.stamp, externalStamp)
			s.stamp += 1
			s.internalChanOut <- strconv.Itoa(s.stamp)
		case "REL":
			externalStamp, _ := strconv.Atoi(tokens[1])
			externalServerNumber, _ := strconv.Atoi(tokens[2])
			s.lamportArray[externalServerNumber] = LamportState{
				State: REL,
				Stamp: externalStamp,
			}
			s.stamp = maxOf(s.stamp, externalStamp)
			s.stamp += 1
			s.internalChanOut <- strconv.Itoa(s.stamp)
		case "LOCALREQ":
			s.stamp += 1
			s.lamportArray[s.serverNumber] = LamportState{
				State: REQ,
				Stamp: s.stamp,
			}
			s.internalChanOut <- strconv.Itoa(s.stamp)
		}

		//Grant SC access if needed
		if s.lamportArray[s.serverNumber].State == REQ {
			correct := true
			for i := 0; i < s.Config.ServerNumber; i++ {
				if i != s.serverNumber {
					if s.lamportArray[i].Stamp <= s.lamportArray[s.serverNumber].Stamp {
						correct = false
						break
					} else if s.lamportArray[i].State == REQ {
						correct = false
						break
					}
				}
			}
		}

	}

}

func maxOf(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func (s *Server) handleLamport(conn net.Conn) {
	for {
		//TODO: Handle error
		input, _ := bufio.NewReader(conn).ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		//Format pour recevoir : ACK <estampille> <num server expediteur>
		tokens := strings.Fields(input)
		//externalStamp, _ := strconv.Atoi(tokens[1])
		externalServerNumber, _ := strconv.Atoi(tokens[2])
		switch tokens[0] {
		case "ACK":
			s.internalChanIn <- input
			_ = <-s.internalChanOut
			fmt.Println(strconv.Itoa(s.serverNumber) + ": ACK " + tokens[1] + " RECEIVED FROM " + tokens[2])

		case "REQ":
			//TODO: REFACTOR THIS SHIT
			s.internalChanIn <- input
			stamp := <-s.internalChanOut
			//LAMPORT NOT OPTIMIZED YET
			fmt.Fprintf(s.OutConnections[externalServerNumber], "ACK %s %d\n", stamp, s.serverNumber)
			fmt.Println(strconv.Itoa(s.serverNumber) + ": REQ " + tokens[1] + " RECEIVED FROM " + tokens[2])
			fmt.Println(strconv.Itoa(s.serverNumber) + ": ACK " + stamp + " SENT TO " + tokens[2])
		case "REL":
			s.internalChanIn <- input
			_ = <-s.internalChanOut
			fmt.Println(strconv.Itoa(s.serverNumber) + ": REL " + tokens[1] + " RECEIVED FROM " + tokens[2])
		}

	}

}

func (s *Server) ConnectToOthers() {
	conf := s.Config
	for i := 0; i < conf.ServerNumber; i++ {
		if i != s.serverNumber {
			//TODO: Handle error
			var conn net.Conn
			var err2 error
			for {
				conn, err2 = net.Dial("tcp", conf.Ips[i])
				if err2 == nil {
					break
				}
				time.Sleep(time.Second)
			}
			s.OutConnections[i] = conn
			fmt.Fprintf(conn, "HELLO "+strconv.Itoa(s.serverNumber)+"\n")
		}
	}
	fmt.Println("CONNECTING SUCCESSFUL ON SERVER " + strconv.Itoa(s.serverNumber+1))
	s.Available = true
	return
}

func (s *Server) AskSC() {
	fmt.Println("SENDING REQUESTS FROM SERVER : " + strconv.Itoa(s.serverNumber))
	numberOfServers := s.Config.ServerNumber
	//Pas besoin de fournir de stamp si on édite notre tableau local
	s.internalChanIn <- "LOCALREQ"
	actualStamp := <-s.internalChanOut

	for i := 0; i < numberOfServers; i++ {
		if i != s.serverNumber {
			fmt.Fprintf(s.OutConnections[i], "REQ %s %d\n", actualStamp, s.serverNumber)
		}
	}

	for {

	}

	//TODO: start waiting for SC
}

type Configuration struct {
	ServerNumber int
	Ips          []string
}

type LamportState struct {
	State LamportType
	Stamp int
}

type LamportType int

const (
	ACK LamportType = iota
	REQ
	REL
)
