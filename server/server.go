// Package server
// Ce package s'occupe de la gestion du serveur de réservation des chambres
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

type Server struct {
	Address	string 	`json:"address"`
	Port	int 	`json:"port"`
	Number	int    	`json:"number"`
}

type AppConfig struct {
	Debug 		bool `json:"debug"`
	NumberRooms int `json:"rooms"`
	NumberDays	int `json:"days"`
	ConnType 	string `json:"connexion"`
}

type Config struct {
	Servers 	[]Server `json:"servers"`
	AppConfig 	AppConfig `json:"appConfig"`
}

func loadConfiguration(filename string) (Config, error) {
	jsonFile, err := os.Open("config.json")

	config := Config{}
	if err != nil {
		fmt.Println(err)
		return config, err
	}

	jsonParser := json.NewDecoder(jsonFile)
	err = jsonParser.Decode(&config)
	return config, err
}


// HandleInternalMessage s'occupe de de l'accès aux données des chambres
func HandleInternalMessage(in <-chan string, out chan<- string, appConfig AppConfig) {
	rooms := make([][]string, appConfig.NumberRooms)
	for i := 0; i < appConfig.NumberRooms; i++ {
		rooms[i] = make([]string, appConfig.NumberDays)
	}
	for {
		message := <-in
		params := strings.Fields(message)
		command := params[0]
		outMessage := ""
		if appConfig.Debug {
			time.Sleep(10 * time.Second)
		}
		switch command {
		case "DISPLAY":
			day, _ := strconv.Atoi(params[1])
			username := params[2]
			for i := 0; i < appConfig.NumberRooms; i++ {
				roomOccupancy := rooms[i][day-1]
				if roomOccupancy == "" {
					roomOccupancy = "EMPTY"
				} else if roomOccupancy == username {
					roomOccupancy = "ALREADY OCCUPIED BY YOU"
				} else {
					roomOccupancy = "OCCUPIED"
				}
				outMessage += fmt.Sprint("ROOM ", i+1, ": ", roomOccupancy, "\r\n")
			}
			out <- outMessage

		case "RESERVE":
			day, _ := strconv.Atoi(params[1])
			roomNumber, _ := strconv.Atoi(params[2])
			duration, _ := strconv.Atoi(params[3])
			username := params[4]
			free := true
			for i := 0; i < duration; i++ {
				if rooms[roomNumber-1][day-1+i] != "" {
					free = false
					out <- "Reservation is impossible, room already occupied\r\n"
					break
				}
			}
			if free {
				for i := 0; i < duration; i++ {
					rooms[roomNumber-1][day-1+i] = username
				}
				output := fmt.Sprint("Reservation complete !\r\n",
					"The room ", roomNumber, " is reserved to you at day ", day, " for ", duration, " nights\r\n")
				out <- output
			}

		case "GETFREE":
			day, _ := strconv.Atoi(params[1])
			duration, _ := strconv.Atoi(params[2])
			free := -1
			for i := 0; i < appConfig.NumberRooms; i++ {
				if rooms[i][day-1] == "" {
					free = i
					for j := 1; j <= duration; j++ {
						if rooms[i][day-1+j] != "" {
							free = -1
						} else {
							break
						}
					}
				}
				if free != -1 {
					break
				}
			}
			output := ""
			if free == -1 {
				output = "Free room not found\r\n"
			} else {
				output = fmt.Sprint("Room ", free+1, " is free at day ", day, " for ", duration, " nights\r\n")
			}
			out <- output
		}

	}
}

//array with other messages stored
var messages[3]Message

type Message struct{
	typ string
	clock uint64
	server int

}
//handle received messages from other servers will be running all the time
func HandleMessageLamport(message Message, clock uint64){
	for {
		if message.typ == "ACK" {
			//compare local clock with received clock
			Compare(clock, message)


			//updates messages with the message received

		}
		if message.typ == "REL" {
			//compare local clock with received clock
			Compare(clock, message)

			//updates messages with the message received
			messages[message.server-1]=message

		}
		if message.typ == "REQ" {
			//compare local clock with received clock
			Compare(clock, message)

			//updates messages with the message received

			// sends ACK to message provider

		}
	}
}

// Increment is used to increment and return the value of the lamport clock
func  Increment(clock uint64)uint64 {
	return clock+1
}

// Compare is called to update our local clock if necessary after
// witnessing a clock value received from another process
func  Compare(clock uint64, otherMessage Message) {
WITNESS:
	// If the other value is old, we do not need to do anything
	if otherMessage.clock < clock {
		Increment(clock)
		return
	}

	// Ensure that our local clock is at least one ahead.
	if !atomic.CompareAndSwapUint64(&clock, clock, otherMessage.clock+1) {
		// The CAS failed, so we just retry. Eventually our CAS should
		// succeed or a future witness will pass us by and our witness
		// will end.
		goto WITNESS
	}
}

func HandleConnexion(server Server, appConfig AppConfig, wg *sync.WaitGroup) {
	defer wg.Done()
	//Create communication channels
	in := make(chan string)
	out := make(chan string)
	//Start reservation goroutine
	go HandleInternalMessage(in, out, appConfig)
	// Listen for incoming connections.
	l, err := net.Listen(appConfig.ConnType, server.Address+":"+strconv.Itoa(server.Port))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	fmt.Println("Listening on " + server.Address + ":" + strconv.Itoa(server.Port))
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go HandleRequest(conn, in, out, appConfig)
	}
}

//TODO On server start create a new clock at 0 when accessing RESERVE SECTION Send message REQ to all other servers and begin lamport algorithm
//main Fonction de base
func main() {
	var wg sync.WaitGroup

	config, errLoadingConfig := loadConfiguration("config.json")

	for i := 0; i < len(config.Servers); i++ {
		wg.Add(1)
		go HandleConnexion(config.Servers[i], config.AppConfig, &wg)
	}
	if errLoadingConfig != nil {
		fmt.Println("An error occur. Please check your config file")
		return
	}

	wg.Wait()
}

// CheckParamBounds is used to check if a string can be converted to int and if it is in the required bounds (included)
// It will return two bool : the value is ok and the connection is ok
func CheckParamBounds(conn net.Conn, param string, lowerBound int, upperBound int) (bool, bool) {
	value, err := strconv.Atoi(param)
	output := ""
	if err != nil {
		output = "Invalid parameter : <" + param + "> is not a number\r\n"
		//If an error happened return the appropriated values
		_, err = conn.Write([]byte(output))
		if err != nil {
			return false, false
		}
		return false, true
	}
	if value < lowerBound || value > upperBound {
		output = fmt.Sprint("Parameter out of bounds : <", param, "> must be between ", lowerBound, " and ", upperBound, "\r\n")
		_, err = conn.Write([]byte(output))
		//If an error happened return the appropriated values
		if err != nil {
			return false, false
		}
		return false, true
	}
	return true, true

}

// HandleRequest is responsible for communicating with the user through the conn variable
// it also uses 2 chanels : in and out, which are used to communicate with the backend goroutine
// In case of sudden disconnect with the user, it will simply stop itself
func HandleRequest(conn net.Conn, in chan<- string, out <-chan string, appConfig AppConfig) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Send the welcome message to the user
	_, err := conn.Write([]byte("Welcome to the hotel reservation service\r\n" +
		"Please enter your name to continue :\r\n"))
	if err != nil {
		return
	}
	// Read the incoming connection into the buffer.
	n, err := conn.Read(buf)
	// If read failed, stop the goroutine
	if err != nil {
		//fmt.Println("Error reading:", err.Error())
		return
	}
	// Send a response back to person contacting us.
	userString := string(buf[:n])
	userString = strings.TrimSuffix(userString, "\n")

	_, err = conn.Write([]byte("Welcome " +
		userString + " !\r\n"))
	if err != nil {
		return
	}
	username := userString

	// Handle user commands
	for {
		n, err := conn.Read(buf)
		//If error in communication stop
		if err != nil {
			//fmt.Println("Error reading:", err.Error())
			return
		}
		//Trim user input into a good string
		userString := string(buf[:n])
		//userString = strings.TrimSuffix(userString, "\r")
		userString = strings.TrimSuffix(userString, "\n")
		//Handle the quit command
		if userString == "QUIT" {
			err = conn.Close()
			if err != nil {
				return
			}
			return
		} else if userString == "HELP" {
			_, err := conn.Write([]byte("Here are the supported commands :\r\n" +
				"Enter QUIT to exit the application\r\n" +
				"Enter DISPLAY <day> to display the room occupancy for the given day\r\n" +
				"Enter RESERVE <day> <room number> <duration> to try to reserve the give room\r\n" +
				"Enter GETFREE <day> <duration> to get a free room number\r\n"))
			if err != nil {
				return
			}
		} else {
			//Parse the user input
			params := strings.Fields(userString)
			command := ""
			//Check there's actually a command
			if len(params) > 0 {
				command = params[0]
				switch command {
				case "DISPLAY":
					//Check the parameters
					if len(params) != 2 {
						_, err := conn.Write([]byte("Invalid parameters, enter HELP to get the list of supported commands\r\n"))
						if err != nil {
							return
						}
					} else {
						valueOk, connOk := CheckParamBounds(conn, params[1], 1, appConfig.NumberDays)
						if !connOk {
							return
						}
						if valueOk {
							//Communicate with the reservation goroutine
							userString += " " + username
							in <- userString
							_, err := conn.Write([]byte(<-out))
							if err != nil {
								return
							}
						}
					}

				case "RESERVE":
					//Check params
					if len(params) != 4 {
						_, err := conn.Write([]byte("Invalid parameters, enter HELP to get the list of supported commands\r\n"))
						if err != nil {
							return
						}
					} else {
						ok1, connOk1 := CheckParamBounds(conn, params[1], 1, appConfig.NumberDays)
						if !connOk1 {
							return
						}
						ok2, connOk2 := CheckParamBounds(conn, params[2], 1, appConfig.NumberRooms)
						if !connOk2 {
							return
						}
						maxDuration := math.MaxInt32
						if ok1 {
							value, _ := strconv.Atoi(params[1])
							maxDuration = appConfig.NumberDays - value + 1
						}
						ok3, connOk3 := CheckParamBounds(conn, params[3], 1, maxDuration)
						if !connOk3 {
							return
						}
						if ok1 && ok2 && ok3 {
							//Communicate with reservation goroutine
							userString += " " + username
							in <- userString
							_, err := conn.Write([]byte(<-out))
							if err != nil {
								return
							}
						}
					}

				case "GETFREE":
					//Check params
					if len(params) != 3 {
						_, err := conn.Write([]byte("Invalid parameters, enter HELP to get the list of supported commands\r\n"))
						if err != nil {
							return
						}
					} else {
						ok1, connOk1 := CheckParamBounds(conn, params[1], 1, appConfig.NumberDays)
						if !connOk1 {
							return
						}
						maxDuration := math.MaxInt32
						if ok1 {
							value, _ := strconv.Atoi(params[1])
							maxDuration = appConfig.NumberDays - value + 1
						}
						ok2, connOk2 := CheckParamBounds(conn, params[2], 1, maxDuration)
						if !connOk2 {
							return
						}
						if ok1 && ok2 {
							//Communicate with reservation goroutine
							in <- userString
							_, err := conn.Write([]byte(<-out))
							if err != nil {
								return
							}
						}
					}

				default:
					_, err := conn.Write([]byte("Unknown command, enter HELP to get the list of supported commands\r\n"))
					if err != nil {
						return
					}
				}
			}
		}
	}
}
