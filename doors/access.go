package doors

import (
	"fmt"
	"log"
	"time"
)

func SetFobEnabled(fob string, amemberUserId int, enabled bool) {
	for _, d := range All() {
		if err := d.SetFobEnabled(fob, fmt.Sprintf("i:%d", amemberUserId), enabled); err != nil {
			log.Printf("[Fob:%s aMemberId:%d door:%s] Enable(%t) error: %s", fob, amemberUserId, d.doorIp, enabled, err.Error())
		}
	}
	time.Sleep(250 * time.Millisecond) // give the device some time to run the command
}

func DeleteAllUsers() {
	for _, d := range All() {
		if err := d.DeleteUserRecords(); err != nil {
			log.Printf("Delete user records on door %s error: %s", d.doorIp, err.Error())
		}
	}
	time.Sleep(15 * time.Second) // give the device some time to run the command
}
