package main

import (
	"bufio"
	"fmt"
	rand2 "math/rand"
	"net"
	"os"
	"strconv"
	"strings"

)
//N Number of servers
const N = 3
// CONN_HOST Host de la connexion
const CONN_HOST = "localhost"
//servers existing server ports
var servers[3]string = [3]string{"3330", "3333", "3339}

func main() {
	arguments := os.Args
	CONNECT := ""
	if len(arguments) == 2 {
		value, _ := strconv.Atoi(arguments[1])
		CONNECT = CONN_HOST+":"+servers[value-1]
	}
	if len(arguments) == 1 {
		CONNECT = CONN_HOST+":"+ servers[rand2.Intn(N)]

	}


	c, err := net.Dial("tcp", CONNECT)
	if err != nil {
		fmt.Println(err)
		return
	}

	for {
		reader := make([]byte, 3000)
		writer := bufio.NewReader(os.Stdin)
		n, _ := c.Read(reader)
		userString := string(reader[:n])
		fmt.Println(userString)
		text, _ := writer.ReadString('\n')
		c.Write([]byte(text));
		if strings.TrimSpace(string(text)) == "QUIT" {
			fmt.Println("TCP client exiting...")
			return
		}
	}
}
