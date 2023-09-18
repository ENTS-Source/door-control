package amember

import (
	"log"
	"sync"

	"github.com/ents-source/door-control/doors"
)

var syncMutex = new(sync.Mutex)

func ResyncAllUsers() {
	syncMutex.Lock()
	defer syncMutex.Unlock()

	users, err := GetAllUsers()
	if err != nil {
		log.Println("Error getting all users for resync: ", err)
		return
	}

	doors.DeleteAllUsers()
	for _, u := range users {
		updateUserFromRecord(u)
	}
}
