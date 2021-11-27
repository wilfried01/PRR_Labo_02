package main

import (
	"bufio"
	"configuration"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	file, _ := os.Open("server/configuration.json")
	defer file.Close()
	decoder := json.NewDecoder(file)
	configuration := configuration.Configuration{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Fatal(err)
		return
	}

	reader := bufio.NewReader(os.Stdin)
	var userInput string
	fmt.Println("Welcome to the room reservation system")
	fmt.Println("Please enter you name :")
	userInput, _ = reader.ReadString('\n')
	username := strings.TrimRightFunc(userInput, func(c rune) bool {
		//In windows newline is \r\n
		return c == '\r' || c == '\n'
	})
	var serverChoosen int
	fmt.Fprintf(os.Stdout, "Hello %s !\n", username)
	for {
		fmt.Fprintf(os.Stdout, "Please choose to which server you want to connect [1..%d], write 0 if you want to have a random server\n", configuration.ServerNumber)
		userInput, _ = reader.ReadString('\n')
		userInput = strings.TrimRightFunc(userInput, func(c rune) bool {
			//In windows newline is \r\n
			return c == '\r' || c == '\n'
		})
		serverChoosen, err = strconv.Atoi(userInput)
		if err == nil {
			if serverChoosen >= 0 && serverChoosen <= configuration.ServerNumber {
				break
			} else {
				fmt.Fprintf(os.Stdout, "Number must be between 0 and %d\n", configuration.ServerNumber)
			}
		} else {
			fmt.Println("Bad input")
		}
	}

	if serverChoosen == 0 {
		s1 := rand.NewSource(time.Now().UnixNano())
		r1 := rand.New(s1)
		serverChoosen = r1.Intn(4) + 1
	}
	//TODO: Handle error
	conn, err := net.Dial("tcp", configuration.Ips[serverChoosen-1])
	if err != nil {
		fmt.Println(err)
	}
	connReader := bufio.NewReader(conn)

	//Send first message from client
	fmt.Fprintf(conn, "CLIENT\n")

	//Get the welcome message
	welcomeMessage, _ := connReader.ReadString('\n')
	fmt.Print(welcomeMessage)

	displayHelp()

	for {
		userInput, _ = reader.ReadString('\n')
		userInput = strings.TrimRightFunc(userInput, func(c rune) bool {
			//In windows newline is \r\n
			return c == '\r' || c == '\n'
		})
		tokens := strings.Fields(userInput)
		if len(tokens) > 0 {
			switch tokens[0] {
			case "HELP":
				displayHelp()
			case "QUIT":
				//Assert that it is the only command
				if len(tokens) == 1 {
					fmt.Fprintf(conn, "%s\n", tokens[0])
					conn.Close()
					return
				}
			case "DISPLAY":
				if len(tokens) == 2 {
					//Verify the value
					value := checkParameter(tokens[1], 1, configuration.NumberOfDays)
					if value != -1 {
						userInput = userInput + " " + username
						fmt.Fprintf(conn, "%s\n", userInput)
						for {
							received, _ := connReader.ReadString('\n')
							if received == "END\n" {
								break
							}
							fmt.Print(received)
						}
					}
				}
				//TODO: Check number of arguments
			case "RESERVE":
				if len(tokens) == 4 {
					day := checkParameter(tokens[1], 1, configuration.NumberOfDays)
					room := checkParameter(tokens[2], 1, configuration.NumberOfRooms)
					duration := checkParameter(tokens[3], 1, configuration.NumberOfDays-day+1)
					if day != -1 && room != -1 && duration != -1 {
						userInput = userInput + " " + username
						fmt.Fprintf(conn, "%s\n", userInput)
						received, _ := connReader.ReadString('\n')
						fmt.Printf(received)
					}
				}
			case "GETFREE":
				if len(tokens) == 3 {
					day := checkParameter(tokens[1], 1, configuration.NumberOfDays)
					room := checkParameter(tokens[2], 1, configuration.NumberOfRooms)
					if day != -1 && room != -1 {
						fmt.Fprintf(conn, "%s\n", userInput)
						received, _ := connReader.ReadString('\n')
						fmt.Printf(received)
					}
				}

			default:
				fmt.Println("Unknown command, enter HELP for displaying the list of commands")
			}
		}
	}
}

func checkParameter(token string, lowerBound int, upperBound int) int {
	value, err := strconv.Atoi(token)
	//Error happened
	if err != nil {
		fmt.Printf("Invalid parameter <%s>", token)
		return -1
	}
	if value < lowerBound || value > upperBound {
		fmt.Printf("Invalid parameter <%s> must be between %d and %d", token, lowerBound, upperBound)
		return -1
	}
	return value

}

func displayHelp() {
	fmt.Println("Here are the supported commands :")
	fmt.Println("Enter QUIT to exit the application")
	fmt.Println("Enter DISPLAY <day> to display the room occupancy for the given day")
	fmt.Println("Enter RESERVE <day> <room number> <duration> to try to reserve the give room")
	fmt.Println("Enter GETFREE <day> <duration> to get a free room number")
	fmt.Println("Enter HELP to display this")
}
