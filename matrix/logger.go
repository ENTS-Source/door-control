package matrix

import (
	"fmt"
	"time"

	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func LogAccess(door string, fob string, time time.Time, granted bool) error {
	if client != nil {
		_, err := client.SendMessageEvent(id.RoomID(LogRoomId), event.EventMessage, map[string]interface{}{
			"msgtype": "m.notice",
			"body":    fmt.Sprintf("%s accessed %s at %s. Granted access? %t", fob, door, time.Format("Jan 2, 2006 3:04:05 PM MST"), granted),
			"ca.ents.door": map[string]interface{}{
				"door":    door,
				"fob":     fob,
				"time":    time.UnixMilli(),
				"granted": granted,
			},
		})
		return err
	}

	return nil
}

func AnnounceAccess(door string, displayName string) error {
	if client != nil {
		_, err := client.SendMessageEvent(id.RoomID(AnnounceRoomId), event.EventMessage, map[string]interface{}{
			"msgtype": "m.notice",
			"body":    fmt.Sprintf("%s entered the space", displayName),
			"ca.ents.door": map[string]interface{}{
				"door":        door,
				"displayName": displayName,
			},
		})
		return err
	}

	return nil
}
