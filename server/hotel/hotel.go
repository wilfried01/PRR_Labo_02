package hotel

type Hotel struct {
	rooms [][]string
}

func (h *Hotel) newHotel(numberOfRooms int, numberOfDays int) {
	//init rooms
	h.rooms = make([][]string, numberOfRooms)
	for i := 0; i < numberOfRooms; i++ {
		h.rooms[i] = make([]string, numberOfDays)
	}
}
