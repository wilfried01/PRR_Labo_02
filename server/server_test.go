package main

import (
	"flag"
	"fmt"
	"net"
	"testing"
)

func connection(fonctionality string) string {
	flag.Parse()
	//buffer for reading greetings
	buf := make([]byte, 1024)
	//buffer for HELP
	buf2 := make([]byte, 1024)
	//buffer for responses
	buf3 := make([]byte, 1024)
	// connection to localhost 3333
	conn, err := net.Dial("tcp", "localhost:3333")
	if err != nil {
		if _, t := err.(*net.OpError); t {
			fmt.Println("Some problem connecting.")
		} else {
			fmt.Println("Unknown error: " + err.Error())
		}
	} else {
		conn.Read(buf)
		//login
		_, err := conn.Write([]byte("Maxime Lagrave"))
		if err != nil {
			return ""
		} else {
			conn.Read(buf2)
			// Write fonctionality
			_, err := conn.Write([]byte(fonctionality))
			if err != nil {
				return ""
			} else {
				n, err := conn.Read(buf3)
				if err != nil {
					fmt.Println("Error reading: 1", err.Error())
					return ""
				}
				conn.Close()
				//return of fonctionality cast of buffer into string
				reservationString := string(buf3[:n])
				return reservationString

			}
		}
	}
	//forcing connection to close
	conn.Close()
	return ""
}

func TestReserve(t *testing.T) {
	reservationString := connection("RESERVE 1 3 2")
	b := "Reservation complete !\r\nThe room 3 is reserved to you at day 1 for 2 nights\r\n"
	if reservationString != b {
		t.Errorf("Reservation should be possible!")
	}

}

func TestReservationNotPossible(t *testing.T) {

	reservationString := connection("RESERVE 1 3 2")
	b := "Reservation is impossible, room already occupied\r\n"
	if reservationString != b {
		t.Errorf("Reservation should not be possible!")
	}

}
func TestReservationInvalidArguments(t *testing.T) {
	reservationString := connection("RESERVE 1 3 2 4")
	b := "Invalid parameters, enter HELP to get the list of supported commands\r\n"
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}

}
func TestReservationInvalidArgumentsIsNotANumber(t *testing.T) {
	reservationString := connection("RESERVE maxime 2 4")
	b := "Invalid parameter : <maxime> is not a number\r\n"
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}

}

func TestDisplay(t *testing.T) {
	reservationString := connection("DISPLAY 1")
	b := ""
	for i := 0; i < 10; i++ {
		if i != 2 {
			b += fmt.Sprint("ROOM ", i+1, ": EMPTY\r\n")
		} else {
			b += fmt.Sprint("ROOM ", i+1, ": ALREADY OCCUPIED BY YOU\r\n")
		}
	}
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}
}

func TestDisplayInvalidArguments(t *testing.T) {
	reservationString := connection("DISPLAY 3 3")
	b := "Invalid parameters, enter HELP to get the list of supported commands\r\n"
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}
}
func TestDisplayInvalidArgumentsIsNotANumber(t *testing.T) {
	reservationString := connection("DISPLAY maxime")
	b := "Invalid parameter : <maxime> is not a number\r\n"
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}

}
func TestGetFree(t *testing.T) {
	reservationString := connection("GETFREE 3 3")
	b := "Room 1 is free at day 3 for 3 nights\r\n"
	if reservationString != b {
		t.Errorf("Room should be Available")
	}
}

func TestGetFreeInvalidArgument(t *testing.T) {
	reservationString := connection("GETFREE 1 3 3")
	b := "Invalid parameters, enter HELP to get the list of supported commands\r\n"
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}
}
func TestGetFreeInvalidArgumentsIsNotANumber(t *testing.T) {
	reservationString := connection("GETFREE maxime 3")
	b := "Invalid parameter : <maxime> is not a number\r\n"
	if reservationString != b {
		t.Errorf("Parameters should be invalid!")
	}

}
