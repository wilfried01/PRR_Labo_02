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
	fmt.Println(serverChoosen)
	//TODO: Handle error
	conn, err := net.Dial("tcp", configuration.Ips[serverChoosen-1])
	if err != nil {
		fmt.Println(err)
	}
	fmt.Fprintf(conn, "CLIENT\n")
	connReader := bufio.NewReader(conn)
	//fmt.Println(connReader.ReadString('\n'))
	fmt.Fprintf(os.Stdout, "%d\n", serverChoosen)
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
				//TODO: Tell the server to close the connection
				return
			case "DISPLAY":
				if len(tokens) == 2 {
					fmt.Fprintf(conn, "DISPLAY\n")
					fmt.Printf(connReader.ReadString('\n'))
				}
				break
			case "RESERVE":
			case "GETFREE":
			}
		}
	}
}

func checkParameter() {

}

func displayHelp() {
	fmt.Println("Here are the supported commands :")
	fmt.Println("Enter QUIT to exit the application")
	fmt.Println("Enter DISPLAY <day> to display the room occupancy for the given day")
	fmt.Println("Enter RESERVE <day> <room number> <duration> to try to reserve the give room")
	fmt.Println("Enter GETFREE <day> <duration> to get a free room number")
	fmt.Println("Enter HELP to display this")
}
