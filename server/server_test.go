package server

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"testing"
	"time"
)

func connection(fonctionality string, ip string) string {
	flag.Parse()
	//buffer for HELP
	//buffer for responses
	buf3 := make([]byte, 1024)
	// connection to localhost 333
	conn, err := net.Dial("tcp", ip)
	if err == nil {
		connReader := bufio.NewReader(conn)

		fmt.Fprintf(conn, "CLIENT\n")

		//Get the welcome message
		welcomeMessage, _ := connReader.ReadString('\n')
		fmt.Print(welcomeMessage)
		// Write fonctionality

		fmt.Fprintf(conn, "%s\n", fonctionality)
		if err != nil {

		} else {
			n, err := conn.Read(buf3)
			if err != nil {
				fmt.Println("Error reading: 1", err.Error())

			}

			//return of fonctionality cast of buffer into string
			reservationString := string(buf3[:n])
			fmt.Print(reservationString)
			time.Sleep(time.Second*10)
			fmt.Fprintf(conn, "%s\n", "QUIT")
			return  reservationString
		}



	}

	return ""
}



func TestReserve(t *testing.T) {
	reservationString := connection("RESERVE 1 3 2 MAXIME","localhost:8000")
	b := "Reservation complete ! The room 3 is reserved to you at day 1 for 2 nights\n"
	if reservationString != b {
		t.Errorf("Reservation should be possible!")
	}

}

func TestReservationNotPossible(t *testing.T) {

	reservationString := connection("RESERVE 1 3 2 MAXINE","localhost:8000")
	b := "Reservation is impossible, room already occupied\n"
	if reservationString != b {
		t.Errorf("Reservation should not be possible!")
	}

}

func TestDisplay(t *testing.T) {
	reservationString := connection("DISPLAY 1 MAXIME","localhost:8000")
	b := ""
	for i := 0; i < 10; i++ {
		if i != 2 {
			b += fmt.Sprint("ROOM ", i+1, ": EMPTY\n")
		} else {
			b += fmt.Sprint("ROOM ", i+1, ": ALREADY OCCUPIED BY YOU\n")
		}

	}
	b+= fmt.Sprint("END\n")
	fmt.Println(b)
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}
}


func TestGetFree(t *testing.T) {
	reservationString := connection("GETFREE 3 3","localhost:8000")
	b := "Room 1 is free at day 3 for 3 nights\n"
	if reservationString != b {
		t.Errorf("Room should be Available")
	}
}
