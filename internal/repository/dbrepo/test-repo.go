package dbrepo

import (
	"errors"
	"time"

	"github.com/DmitryZzz/bookings/internal/models"
)

func (m *testDBRepo) AllUsers() bool {
	return true
}

// InsertReservation inserts a reservation into the database
func (m *testDBRepo) InsertReservation(res models.Reservation) (int, error) {
	// if the room id is 2, then fail; otherwise, pass
	if res.RoomID == 2 {
		return 0, errors.New("some error")
	}
	return 1, nil
}

// InsertRoomRestriction inserts a room restriction into the database
func (m *testDBRepo) InsertRoomRestriction(r models.RoomRestriction) error {
	if r.RoomID == 1000 {
		return errors.New("some error")	
	}
	return nil
}

// SearchAvailabilityByDates returns true if availability exists for roomID, and false if no availability exists
func (m *testDBRepo) SearchAvailabilityByDatesByRoomID(start, end time.Time, roomId int) (bool, error) {
	return false, nil
}

// SearchAvailabilityForAllRooms returns a slice of available rooms, if any, for given date range
func (m *testDBRepo) SearchAvailabilityForAllRooms(start, end time.Time) ([]models.Room, error) {
	var rooms []models.Room

	layout := "2006-01-02"
	badDate, _ := time.Parse(layout, "2050-12-31")
	noRoomsDate, _ := time.Parse(layout, "2050-12-30")

	if end == badDate {
		return rooms, errors.New("some error")
	}
	if end == noRoomsDate {
		return rooms, nil
	}
	rooms = append(rooms, models.Room{ID: 1, RoomName: "General`s Quarters"})

	return rooms, nil
}

// GetRoomByID gets a room by id
func (m *testDBRepo) GetRoomByID(id int) (models.Room, error) {
	var room models.Room
	if id > 2 {
		return room, errors.New("some error")
	}
	room = models.Room{ID: 1, RoomName: "General`s Quarters"}	
	return room, nil
}
