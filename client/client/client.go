package client

import (
	"bufio"
	"fmt"
	"os"
)

func main() {
	reader := bufio.NewReader(os.Stdin)
	var userInput string
	fmt.Println("Welcome to the room reservation system")
	fmt.Println("Please enter you name :")
	userInput, _ = reader.ReadString('\n')
	username := userInput
	fmt.Fprintf(os.Stdout, "Hello %s !\n", username)

}
