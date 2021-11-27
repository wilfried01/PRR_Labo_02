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
	hotel        *hotel.Hotel
	serverNumber int
	tcpListener  net.Listener
	//TODO: put this private, used right now for debugging purposes
	OutConnections       []net.Conn
	InConnections        []net.Conn
	Available            bool
	stamp                int
	internalChanIn       chan string
	internalChanOut      chan string
	internalHotelChanIn  chan string
	internalHotelChanOut chan string

	scChan       chan bool
	inSc         bool
	debugMode    bool
	lamportArray []LamportState
	Config       configuration.Configuration
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
	server.internalHotelChanIn = make(chan string)
	server.internalHotelChanOut = make(chan string)
	server.scChan = make(chan bool)
	server.lamportArray = make([]LamportState, configuration.ServerNumber)
	server.inSc = false
	server.hotel = hotel.NewHotel(configuration.NumberOfRooms, configuration.NumberOfDays, server.debugMode)
	go server.hotel.HandleInternalMessages(server.internalHotelChanIn, server.internalHotelChanOut)
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
	//Pass the configuration numbers to the client

	fmt.Fprintf(conn, "Welcome to the reservation server number %d\n", s.serverNumber)

	for {
		userInput, _ := bufio.NewReader(conn).ReadString('\n')
		tokens := strings.Fields(userInput)
		fmt.Printf("RECEIVED %s", userInput)
		if tokens[0] == "QUIT" {
			conn.Close()
			return
		} else {
			s.AskSC()

			needSync := false
			if tokens[0] != "DISPLAY" {
				needSync = true
			}

			s.internalHotelChanIn <- userInput
			output := <-s.internalHotelChanOut
			fmt.Fprintf(conn, "%s", output)
			if needSync {
				for i := 0; i < s.Config.ServerNumber; i++ {
					if i != s.serverNumber {
						s.internalHotelChanIn <- "GETDATA"
						output = <-s.internalHotelChanOut
						fmt.Fprintf(s.OutConnections[i], "SYNC %s\n", output)
					}
				}
			}

			s.releaseSC()

		}
	}
}

func (s *Server) StartListening() {
	for {
		//TODO: Handle errors
		//TODO: Add defer
		conn, _ := s.tcpListener.Accept()
		if !s.GetAvailable() {
			input, _ := bufio.NewReader(conn).ReadString('\n')
			input = strings.TrimSuffix(input, "\n")
			tokens := strings.Fields(input)
			size := len(tokens)
			if size >= 2 && tokens[0] == "HELLO" {
				inNumber, _ := strconv.Atoi(tokens[1])
				s.InConnections[inNumber] = conn
				go s.handleLamport(conn)
				correct := true
				for i := 0; i < s.Config.ServerNumber; i++ {
					if i != s.serverNumber {
						if s.InConnections[i] == nil {
							correct = false
							break
						}
					}
				}
				if correct {
					s.internalChanIn <- "AVAILABLE"
					_ = <-s.internalChanOut
				}
			} else {
				//In the case the servers aren't ready we tell the client the system is unavailable
				fmt.Fprintf(conn, "System is not availabe now, please retry later\n")
				conn.Close()
				//TODO: Gracefully quit the client
			}
		} else {
			//Read first line sent by client to init
			_, _ = bufio.NewReader(conn).ReadString('\n')
			go s.handleClient(conn)
		}
	}
}

func (s *Server) GetAvailable() bool {
	s.internalChanIn <- "GETAVAILABLE"
	output := <-s.internalChanOut
	return output == "TRUE"
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
		tokens := strings.Fields(input)
		if tokens[0] == "SYNC" {
			s.internalHotelChanIn <- fmt.Sprintf("SETDATA %s", tokens[1])
			_ = <-s.internalHotelChanOut
			continue
		}
		//Format pour recevoir : ACK <estampille> <num server expediteur>
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
	_ = <-s.scChan
	output := fmt.Sprintf("Server %d entering SC", s.serverNumber)
	fmt.Println(output)

	//Sleeping to simulate SC treatement
	if s.debugMode {
		time.Sleep(time.Second * 5)
	}

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
