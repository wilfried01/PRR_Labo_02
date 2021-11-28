package hotel

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Hotel represents a hotel with some rooms for a certain amount of time
type Hotel struct {
	rooms       [][]string
	numberRooms int
	numberDays  int
	debugMode   bool
}

// NewHotel return a struct Hotel initialized with the parameters
func NewHotel(numberOfRooms int, numberOfDays int, debugMode bool) *Hotel {
	//init rooms
	h := Hotel{
		numberRooms: numberOfRooms,
		numberDays:  numberOfDays,
		debugMode:   debugMode,
	}
	h.rooms = make([][]string, numberOfRooms)
	for i := 0; i < numberOfRooms; i++ {
		h.rooms[i] = make([]string, numberOfDays)
	}
	return &h
}

// HandleInternalMessages is responsible for accessing rooms data.
// This takes 2 channels as parameters, one for receiving, one for sending
func (h *Hotel) HandleInternalMessages(in <-chan string, out chan<- string) {
	stop := false
	for stop == false {
		message := <-in
		params := strings.Fields(message)
		command := params[0]
		outMessage := ""
		if h.debugMode {
			time.Sleep(5 * time.Second)
		}
		switch command {
		case "DISPLAY":
			day, _ := strconv.Atoi(params[1])
			username := params[2]
			for i := 0; i < h.numberRooms; i++ {
				roomOccupancy := h.rooms[i][day-1]
				if roomOccupancy == "" {
					roomOccupancy = "EMPTY"
				} else if roomOccupancy == username {
					roomOccupancy = "ALREADY OCCUPIED BY YOU"
				} else {
					roomOccupancy = "OCCUPIED"
				}
				outMessage += fmt.Sprint("ROOM ", i+1, ": ", roomOccupancy, "\n")
			}
			outMessage += "END\n"
			out <- outMessage

		case "RESERVE":
			day, _ := strconv.Atoi(params[1])
			roomNumber, _ := strconv.Atoi(params[2])
			duration, _ := strconv.Atoi(params[3])
			username := params[4]
			free := true
			for i := 0; i < duration; i++ {
				if h.rooms[roomNumber-1][day-1+i] != "" {
					free = false
					out <- "Reservation is impossible, room already occupied\n"
					break
				}
			}
			if free {
				for i := 0; i < duration; i++ {
					h.rooms[roomNumber-1][day-1+i] = username
				}
				output := fmt.Sprint("Reservation complete ! ",
					"The room ", roomNumber, " is reserved to you at day ", day, " for ", duration, " nights\n")
				out <- output
			}

		case "GETFREE":
			day, _ := strconv.Atoi(params[1])
			duration, _ := strconv.Atoi(params[2])
			free := -1
			for i := 0; i < h.numberRooms; i++ {
				if h.rooms[i][day-1] == "" {
					free = i
					for j := 1; j <= duration; j++ {
						if h.rooms[i][day-1+j] != "" {
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
				output = "Free room not found\n"
			} else {
				output = fmt.Sprint("Room ", free+1, " is free at day ", day, " for ", duration, " nights\n")
			}
			out <- output

		case "GETDATA":
			bytes, err := json.Marshal(h.rooms)
			if err != nil {
				fmt.Println("Something went bad")
			}
			out <- string(bytes)
		case "SETDATA":
			data := params[1]
			var updatedRooms [][]string
			err := json.Unmarshal([]byte(data), &updatedRooms)
			if err != nil {
				fmt.Println("Something went wrong ...")
			}
			h.rooms = updatedRooms
			out <- "OK"
		case "STOP":
			return
		}

	}
}
