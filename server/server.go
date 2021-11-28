package server

import (
	"bufio"
	"configuration"
	"fmt"
	"net"
	"os"
	"os/signal"
	"server/hotel"
	"strconv"
	"strings"
	"syscall"
	"time"
)

type Server struct {
	hotel                *hotel.Hotel
	serverNumber         int
	tcpListener          net.Listener
	OutConnections       []net.Conn
	InConnections        []net.Conn
	Available            bool
	stamp                int
	internalChanIn       chan string
	internalChanOut      chan string
	internalHotelChanIn  chan string
	internalHotelChanOut chan string
	scChan               chan bool
	inSc                 bool
	debugMode            bool
	lamportArray         []LamportState
	Config               configuration.Configuration
}

//NewServer handles creating a new server with correct parameters,
//server number starts at 1
func NewServer(serverNumber int, debug bool, inputConfig configuration.Configuration) *Server {
	number := serverNumber - 1
	server := Server{serverNumber: number, Available: false}
	server.stamp = 0

	server.Config = inputConfig

	//Create internal variables
	server.debugMode = debug
	server.OutConnections = make([]net.Conn, server.Config.ServerNumber)
	server.InConnections = make([]net.Conn, server.Config.ServerNumber)
	server.internalChanIn = make(chan string)
	server.internalChanOut = make(chan string)
	server.internalHotelChanIn = make(chan string)
	server.internalHotelChanOut = make(chan string)
	server.scChan = make(chan bool)
	server.lamportArray = make([]LamportState, server.Config.ServerNumber)
	server.inSc = false
	server.hotel = hotel.NewHotel(server.Config.NumberOfRooms, server.Config.NumberOfDays, server.debugMode)
	go server.hotel.HandleInternalMessages(server.internalHotelChanIn, server.internalHotelChanOut)
	for i := 0; i < server.Config.ServerNumber; i++ {
		server.lamportArray[i] = LamportState{State: REL, Stamp: 0}
	}

	server.tcpListener, _ = net.Listen("tcp", server.Config.Ips[number])
	go server.HandleInternalMessages()
	go server.StartListening()
	go server.ConnectToOthers()

	return &server
}

// HandleClient is responsible to handle a client via the connection passed in parameter.
// This method will read what the client sent, ask the other servers the SC, pass the command to the goroutine responsible for
// handling the rooms, sync the data with the other servers and then release th SC
func (s *Server) HandleClient(conn net.Conn) {

	// Write hello message
	fmt.Fprintf(conn, "Welcome to the reservation server number %d\n", s.serverNumber)

	for {
		//Read user input
		userInput, _ := bufio.NewReader(conn).ReadString('\n')
		tokens := strings.Fields(userInput)
		fmt.Printf("RECEIVED %s", userInput)
		if tokens[0] == "QUIT" {
			conn.Close()
			return
		} else {
			//Ask SC
			s.AskSC()

			//Check if sync will be needed
			needSync := false
			if tokens[0] != "DISPLAY" {
				needSync = true
			}

			//Execute user command
			s.internalHotelChanIn <- userInput
			output := <-s.internalHotelChanOut

			//Send answer to the user
			fmt.Fprintf(conn, "%s", output)

			//Sync the data with the other servers
			if needSync {
				for i := 0; i < s.Config.ServerNumber; i++ {
					if i != s.serverNumber {
						s.internalHotelChanIn <- "GETDATA"
						output = <-s.internalHotelChanOut
						fmt.Fprintf(s.OutConnections[i], "SYNC %s\n", output)
					}
				}
			}

			//Release th SC
			s.ReleaseSC()

		}
	}
}

// StartListening starts listening for incoming connections. It checks whether it's a client or another server, handles
// server handshakes or pass the connection to the goroutine handling the clients
func (s *Server) StartListening() {
	for {
		conn, _ := s.tcpListener.Accept()
		c := make(chan os.Signal)
		// handle panic quits
		signal.Notify(c, os.Interrupt, syscall.SIGINT)
		go func() {
			<-c
			conn.Close()
			os.Exit(1)
		}()
		//If all the servers aren't here
		if !s.GetAvailable() {
			input, _ := bufio.NewReader(conn).ReadString('\n')
			input = strings.TrimSuffix(input, "\n")
			tokens := strings.Fields(input)
			size := len(tokens)

			//Check if it's a server
			if size >= 2 && tokens[0] == "HELLO" {
				inNumber, _ := strconv.Atoi(tokens[1])
				s.InConnections[inNumber] = conn

				//Start the goroutine responsible for handling the lamport clock
				go s.HandleLamport(conn)
				correct := true

				//Check if all the servers are here
				for i := 0; i < s.Config.ServerNumber; i++ {
					if i != s.serverNumber {
						if s.InConnections[i] == nil {
							correct = false
							break
						}
					}
				}

				//Change the state
				if correct {
					s.internalChanIn <- "AVAILABLE"
					_ = <-s.internalChanOut
				}
			} else {
				//In the case the servers aren't ready we tell the client the system is unavailable
				fmt.Fprintf(conn, "System is not availabe now, please retry later\n")
				conn.Close()
			}
		} else {
			//Read first line sent by client to init
			_, _ = bufio.NewReader(conn).ReadString('\n')

			//Pass the connection to the goroutine responsible for handling the client
			go s.HandleClient(conn)
		}
	}
}

// GetAvailable checks if the server is ready to handle clients
func (s *Server) GetAvailable() bool {
	s.internalChanIn <- "GETAVAILABLE"
	output := <-s.internalChanOut
	return output == "TRUE"
}

// HandleInternalMessages is responsible for managing all the data which needs to be accessed from multiple goroutines.
// It handles the lamport array and also checks if the server is authorized to go into SC
func (s *Server) HandleInternalMessages() {
	for {
		//Format CMD Params
		input := <-s.internalChanIn
		tokens := strings.Fields(input)
		switch tokens[0] {
		//Case acknowledge
		case "ACK":
			externalStamp, _ := strconv.Atoi(tokens[1])
			externalServerNumber, _ := strconv.Atoi(tokens[2])
			if s.lamportArray[externalServerNumber].State != REQ {
				s.lamportArray[externalServerNumber] = LamportState{
					State: ACK,
					Stamp: externalStamp,
				}
			}
			s.stamp = MaxOf(s.stamp, externalStamp)
			s.stamp += 1
			s.internalChanOut <- strconv.Itoa(s.stamp)

		//Case require
		case "REQ":
			externalStamp, _ := strconv.Atoi(tokens[1])
			externalServerNumber, _ := strconv.Atoi(tokens[2])
			s.lamportArray[externalServerNumber] = LamportState{
				State: REQ,
				Stamp: externalStamp,
			}
			s.stamp = MaxOf(s.stamp, externalStamp)
			s.stamp += 1
			s.internalChanOut <- strconv.Itoa(s.stamp)
		//Case release
		case "REL":
			externalStamp, _ := strconv.Atoi(tokens[1])
			externalServerNumber, _ := strconv.Atoi(tokens[2])
			s.lamportArray[externalServerNumber] = LamportState{
				State: REL,
				Stamp: externalStamp,
			}
			s.stamp = MaxOf(s.stamp, externalStamp)
			s.stamp += 1
			s.internalChanOut <- strconv.Itoa(s.stamp)
		//Case local require is to only modify local entry
		case "LOCALREQ":
			s.stamp += 1
			s.lamportArray[s.serverNumber] = LamportState{
				State: REQ,
				Stamp: s.stamp,
			}
			s.internalChanOut <- strconv.Itoa(s.stamp)
		//Case local release is to only modify local entry
		case "LOCALREL":
			s.stamp += 1
			s.lamportArray[s.serverNumber] = LamportState{
				State: REL,
				Stamp: s.stamp,
			}
			s.inSc = false
			s.internalChanOut <- strconv.Itoa(s.stamp)

		//Set the server to ready for handling clients
		case "AVAILABLE":
			s.Available = true
			s.internalChanOut <- "OK"
		//Check server state
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
			time.Sleep(time.Second * 10)
		}
	}

}

// MaxOf returns the maximum between two integers
func MaxOf(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

// HandleLamport handles receiving and sending lamport message to another server
func (s *Server) HandleLamport(conn net.Conn) {
	for {
		//Read server input
		input, _ := bufio.NewReader(conn).ReadString('\n')
		input = strings.TrimSuffix(input, "\n")
		tokens := strings.Fields(input)

		//If we have to sync data
		if tokens[0] == "SYNC" {
			s.internalHotelChanIn <- fmt.Sprintf("SETDATA %s", tokens[1])
			_ = <-s.internalHotelChanOut
			continue
		}

		externalServerNumber, _ := strconv.Atoi(tokens[2])
		switch tokens[0] {
		//Case acknowledge
		case "ACK":
			s.internalChanIn <- input
			_ = <-s.internalChanOut
			fmt.Println(strconv.Itoa(s.serverNumber) + ": ACK " + tokens[1] + " RECEIVED FROM " + tokens[2])
		//Case require
		case "REQ":
			s.internalChanIn <- input
			stamp := <-s.internalChanOut
			fmt.Fprintf(s.OutConnections[externalServerNumber], "ACK %s %d\n", stamp, s.serverNumber)
			fmt.Println(strconv.Itoa(s.serverNumber) + ": REQ " + tokens[1] + " RECEIVED FROM " + tokens[2])
			fmt.Println(strconv.Itoa(s.serverNumber) + ": ACK " + stamp + " SENT TO " + tokens[2])
		//Case Release
		case "REL":
			s.internalChanIn <- input
			_ = <-s.internalChanOut
			fmt.Println(strconv.Itoa(s.serverNumber) + ": REL " + tokens[1] + " RECEIVED FROM " + tokens[2])
		}

	}

}

// ConnectToOthers establishes connection to all the other servers specified in the configuration
func (s *Server) ConnectToOthers() {
	conf := s.Config

	// Try to connect to every server
	for i := 0; i < conf.ServerNumber; i++ {
		if i != s.serverNumber {

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

// AskSC asks the SC to the other servers
func (s *Server) AskSC() {

	fmt.Println("SENDING REQUESTS FROM SERVER : " + strconv.Itoa(s.serverNumber))
	numberOfServers := s.Config.ServerNumber

	s.internalChanIn <- "LOCALREQ"
	actualStamp := <-s.internalChanOut

	//Send the require to the other servers
	for i := 0; i < numberOfServers; i++ {
		if i != s.serverNumber {
			fmt.Fprintf(s.OutConnections[i], "REQ %s %d\n", actualStamp, s.serverNumber)
		}
	}

	//Wait till we have the sc
	_ = <-s.scChan

	output := fmt.Sprintf("Server %d entering SC", s.serverNumber)
	fmt.Println(output)

	//Sleeping to simulate SC treatement
	if s.debugMode {
		time.Sleep(time.Second * 10)
	}

}

// ReleaseSC releases the SC
func (s *Server) ReleaseSC() {
	fmt.Println("Sending LOCALREL")
	s.internalChanIn <- "LOCALREL"
	if s.debugMode {
		time.Sleep(time.Second * 10)
	}
	actualStamp := <-s.internalChanOut
	output := fmt.Sprintf("Server %d leaving SC", s.serverNumber)
	fmt.Println(output)
	// Send release messages to the other servers
	for i := 0; i < s.Config.ServerNumber; i++ {
		if i != s.serverNumber {
			fmt.Fprintf(s.OutConnections[i], "REL %s %d\n", actualStamp, s.serverNumber)
		}
	}
}

// The LamportState represents the different messages exchanged by the servers.
// It contains the type of message (LamportType) and the stamp
type LamportState struct {
	State LamportType // Type
	Stamp int         // Stamp
}

// The LamportType represents the three different states of a Lamport clock :
//  - Acknowledge
//  - Require
//  - Release
type LamportType int

const (
	ACK LamportType = iota
	REQ
	REL
)
