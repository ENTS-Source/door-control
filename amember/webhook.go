package amember

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/ents-source/door-control/doors"
)

func InstallApi(productCategoryId int) {
	http.HandleFunc("/v1/amember", func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		b, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		vals, err := url.ParseQuery(string(b))
		ev := vals.Get("am-event")
		log.Println("[aMember Pro Webhook]", ev)
		switch ev {
		case "accessAfterInsert":
			fallthrough
		case "accessAfterUpdate":
			fallthrough
		case "accessAfterDelete":
			fallthrough
		case "userAfterInsert":
			fallthrough
		case "userAfterUpdate":
			userId, err := strconv.Atoi(vals.Get("user[user_id]"))
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				return
			}
			resyncUser(userId)
			w.WriteHeader(http.StatusOK)
			return
		case "userAfterDelete":
			// TODO@@
		default:
			w.WriteHeader(http.StatusNotFound)
			return
		}
	})
}

func resyncUser(id int) {
	log.Println("Syncing access records for user ID", id)
	user, err := GetUser(id)
	if err != nil {
		log.Println("[DOOR UPDATE]", err)
		return
	}

	updateUserFromRecord(user)
}

func updateUserFromRecord(user User) {
	if user.Fob == "" || user.Fob == "N/A" {
		return // no possible change
	}

	switch user.FobAccess {
	case "enabled":
		doors.SetFobEnabled(user.Fob, user.Id, true)
	case "disabled":
		doors.SetFobEnabled(user.Fob, user.Id, false)
	case "subscription":
		fallthrough
	case "":
		// todo
	default:
		log.Println("Unknown FobAccess value", user.FobAccess)
		return
	}
}
