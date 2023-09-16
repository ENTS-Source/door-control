package doors

import (
	"log"
	"sync"
	"time"
)

var doors = make([]*Door, 0)
var doorMutex sync.Mutex

func Count(offlineAfter time.Duration) int {
	doorMutex.Lock()
	defer doorMutex.Unlock()

	online := 0
	for _, d := range doors {
		if time.Until(d.lastPing) < offlineAfter {
			online++
		}
	}

	return online
}

type Door struct {
	doorIp   string
	lastPing time.Time
}

func (d *Door) Open() error {
	return nil
}

func discoverDoor(msg map[string]any) {
	ip, err := readMessageVal[string](msg, "ip")
	if err != nil {
		log.Println("Cannot discover door: ", err)
	}

	doorMutex.Lock()
	defer doorMutex.Unlock()

	for _, d := range doors {
		if d.doorIp == ip {
			log.Println("Upticking door ", ip)
			d.lastPing = time.Now()
			return
		}
	}

	log.Println("Adding door ", ip)
	doors = append(doors, &Door{
		doorIp:   ip,
		lastPing: time.Now(),
	})
}
