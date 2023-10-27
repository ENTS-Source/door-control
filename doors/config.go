package doors

import "time"

var OfflineAfter time.Duration
var OnAccess func(door string, fob string, time time.Time, granted bool)
