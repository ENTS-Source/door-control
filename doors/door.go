package doors

import (
	"log"
	"sync"
	"time"
)

var doors = make([]*Door, 0)
var doorMutex sync.Mutex

func All() []*Door {
	doorMutex.Lock()
	defer doorMutex.Unlock()

	newDoors := make([]*Door, 0)
	for _, d := range doors {
		newDoors = append(newDoors, d)
	}
	return newDoors
}

type Door struct {
	doorIp   string
	lastPing time.Time
}

func (d *Door) Open() error {
	return sendCommand(map[string]any{
		"cmd":    "opendoor",
		"doorip": d.doorIp,
	})
}

func (d *Door) SetFobEnabled(fob string, userName string, enabled bool) error {
	acctype := 0
	if enabled {
		acctype = 1
	}
	return sendCommand(map[string]any{
		"cmd":        "adduser",
		"doorip":     d.doorIp,
		"uid":        fob,
		"user":       userName,
		"acctype":    acctype,
		"validuntil": 2145916800, // forever (effectively)
	})
}

func (d *Door) IsOnline() bool {
	return time.Now().Sub(d.lastPing) < OfflineAfter
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
