package doors

import (
	"fmt"
	"log"
)

func SetFobEnabled(fob string, amemberUserId int, enabled bool) {
	for _, d := range All() {
		if err := d.SetFobEnabled(fob, fmt.Sprintf("i:%d", amemberUserId), enabled); err != nil {
			log.Printf("[Fob:%s aMemberId:%d] Enable(%t) error: %s", fob, amemberUserId, enabled, err.Error())
		}
	}
}
