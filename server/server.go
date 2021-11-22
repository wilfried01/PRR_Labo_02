package server

import (
	"bufio"
	"configuration"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"server/hotel"
	"strconv"
	"strings"
	"time"
)

type Server struct {
	hotel        hotel.Hotel
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
	inSc            bool
	debugMode       bool
	lamportArray    []LamportState
	Config          configuration.Configuration
}

//NewServer handles creating a new server with correct parameters,
//server number starts at 1
func NewServer(serverNumber int) *Server {
	number := serverNumber - 1
	server := Server{serverNumber: number, Available: false}
	server.stamp = 0
	//Read config
	file, _ := os.Open("server/configuration.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := configuration.Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
		return nil
	}
	server.Config = configuration

	//Create internal variables
	server.debugMode = false
	server.OutConnections = make([]net.Conn, configuration.ServerNumber)
	server.InConnections = make([]net.Conn, configuration.ServerNumber)
	server.internalChanIn = make(chan string)
	server.internalChanOut = make(chan string)
	server.scChan = make(chan bool)
	server.lamportArray = make([]LamportState, configuration.ServerNumber)
	server.inSc = false
	for i := 0; i < configuration.ServerNumber; i++ {
		server.lamportArray[i] = LamportState{State: REL, Stamp: 0}
	}

	server.tcpListener, _ = net.Listen("tcp", configuration.Ips[number])
	go server.handleInternalMessages()
	go server.StartListening()
	go server.ConnectToOthers()

	return &server
}

func (s *Server) handleClient(conn net.Conn) {
	fmt.Println("ASDASD")
	//fmt.Println(bufio.NewReader(conn).ReadString('\n'))
	fmt.Println("ASDASDASD")
	//fmt.Println(bufio.NewReader(conn).ReadString('\n'))
	//fmt.Fprintf(conn, "Hello client\n")
	for {
		userInput, _ := bufio.NewReader(conn).ReadString('\n')
		fmt.Printf("RECEIVED %s", userInput)
		//fmt.Fprintf(conn, userInput)
		//s.AskSC()
		//time.Sleep(time.Second * 3)
		//fmt.Fprintf(conn, "WELCOME \n")
		fmt.Fprintf(conn, userInput)
		//s.releaseSC()
	}
}

func (s *Server) StartListening() {
	for {
		//TODO: Handle errors
		//TODO: Add defer
		conn, _ := s.tcpListener.Accept()
		input, _ := bufio.NewReader(conn).ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		tokens := strings.Fields(input)
		size := len(tokens)
		if size >= 2 && tokens[0] == "HELLO" {
			inNumber, _ := strconv.Atoi(tokens[1])
			s.InConnections[inNumber] = conn
			go s.handleLamport(conn)
		} else {
			fmt.Println("RECEIVED CLIENT")
			s.internalChanIn <- "GETAVAILABLE"
			response := <-s.internalChanOut
			fmt.Println("RECEIVED CLIENT 2")
			if response != "TRUE" {
				fmt.Fprintf(conn, "System is not availabe now, please retry later\n")
			} else {
				go s.handleClient(conn)
			}
		}
		//fmt.Println(input)
	}
}

func (s *Server) GetAvailable() bool {
	s.internalChanIn <- "GETAVAILABLE"
	output := <-s.internalChanOut
	if output == "TRUE" {
		return true
	}
	return false
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
		case "LOCALREL":
			s.stamp += 1
			s.lamportArray[s.serverNumber] = LamportState{
				State: REL,
				Stamp: s.stamp,
			}
			s.inSc = false
			s.internalChanOut <- strconv.Itoa(s.stamp)
		//Write external data
		case "SYNCDATA":

		//Return local data
		case "GETDATA":

		case "AVAILABLE":
			s.Available = true
			s.internalChanOut <- "OK"
		case "GETAVAILABLE":
			if s.Available {
				s.internalChanOut <- "TRUE"
			} else {
				s.internalChanOut <- "FALSE"
			}
		}

		//Check if sc is asked and do not check if already in sc
		if s.lamportArray[s.serverNumber].State == REQ && !s.inSc {
			correct := true
			for i := 0; i < s.Config.ServerNumber; i++ {
				//Check only other servers state
				if i != s.serverNumber {
					//Check if stamp is bigger than the REQ
					if s.lamportArray[i].Stamp < s.lamportArray[s.serverNumber].Stamp {
						correct = false
						break
						//Check if other is also demanding the sc and if it's server number is lower than this one
					} else if s.lamportArray[i].State == REQ && s.lamportArray[i].Stamp == s.lamportArray[s.serverNumber].Stamp {
						//Server doesn't have priority
						if i <= s.serverNumber {
							correct = false
							break
							//Server has priority
						} else {
							break
						}
					}
				}
			}
			//If sc is available send something to release waiting goroutine
			if correct {
				s.inSc = true
				s.scChan <- true
			}
		}
		if s.debugMode {
			time.Sleep(time.Second)
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
			//Make handshake
			s.OutConnections[i] = conn
			fmt.Fprintf(conn, "HELLO "+strconv.Itoa(s.serverNumber)+"\n")
		}
	}
	fmt.Println("CONNECTING SUCCESSFUL ON SERVER " + strconv.Itoa(s.serverNumber+1))
	s.internalChanIn <- "AVAILABLE"
	_ = <-s.internalChanOut
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

	//TODO: Do something while you have SC

	//TODO: Send modified content to other servers, could send userinput if rooms are modified
	_ = <-s.scChan
	output := fmt.Sprintf("Server %d entering SC", s.serverNumber)
	fmt.Println(output)

	//Sleeping to simulate SC treatement
	if s.debugMode {
		time.Sleep(time.Second * 5)
	}

	//TODO: Remove this, used for tests RN
	//s.releaseSC()

}

func (s *Server) releaseSC() {
	fmt.Println("Sending LOCALREL")
	s.internalChanIn <- "LOCALREL"
	if s.debugMode {
		time.Sleep(time.Second * 3)
	}
	actualStamp := <-s.internalChanOut
	output := fmt.Sprintf("Server %d leaving SC", s.serverNumber)
	fmt.Println(output)
	for i := 0; i < s.Config.ServerNumber; i++ {
		if i != s.serverNumber {
			fmt.Fprintf(s.OutConnections[i], "REL %s %d\n", actualStamp, s.serverNumber)
		}
	}
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
