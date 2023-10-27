package amember

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/ents-source/door-control/api/auth"
	"github.com/ents-source/door-control/db"
	"github.com/ents-source/door-control/doors"
)

var ProductCategoryId int
var AccessBufferDays int

func InstallApi() {
	http.HandleFunc("/v1/amember", doWebhook)
	http.HandleFunc("/v1/amember/resync", auth.Require(doResync))
}

func doResync(w http.ResponseWriter, r *http.Request) {
	ResyncAllUsers()
	w.WriteHeader(http.StatusOK)
}

func doWebhook(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("User %s was deleted - they will be caught in the next global sync", vals.Get("user[user_id]"))
	default:
		w.WriteHeader(http.StatusNotFound)
		return
	}
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
	productIds, err := GetProductIdsForCategory(ProductCategoryId)
	if err != nil {
		log.Println("Error asking for category information", err)
		return
	}

	switch user.FobAccess {
	case "enabled":
		doors.SetFobEnabled(user.Fob, user.Id, true)
	case "disabled":
		doors.SetFobEnabled(user.Fob, user.Id, false)
	case "subscription":
		fallthrough
	case "":
		// First we need to find an access record for the product category we should be
		// looking for.
		didUpdate := false
		for _, row := range user.Nested.Access {
			usefulRow := false
			for _, pid := range productIds {
				if row.ProductId == strconv.Itoa(pid) {
					usefulRow = true
					break
				}
			}
			if !usefulRow {
				continue
			}

			startDate, err := time.Parse("2006-01-02", row.BeginDate)
			if err != nil {
				log.Println("Error parsing start date", row.BeginDate, err)
			}
			endDate, err := time.Parse("2006-01-02", row.EndDate)
			if err != nil {
				log.Println("Error parsing end date", row.EndDate, err)
			}
			endDate = endDate.Add(time.Duration(AccessBufferDays) * 24 * time.Hour)

			if startDate.Before(time.Now()) && endDate.After(time.Now()) {
				doors.SetFobEnabled(user.Fob, user.Id, true)
				didUpdate = true
				break
			}
		}
		if !didUpdate {
			doors.SetFobEnabled(user.Fob, user.Id, false) // no useful access records
		}
	default:
		log.Println("Unknown FobAccess value", user.FobAccess)
	}

	announceEnabled := false
	for _, v := range user.Announce {
		if v == "announce" {
			announceEnabled = true
			if err = db.UpsertAnnounce(user.Fob, true, user.Nickname); err != nil {
				log.Println("Error upserting announce status=true", err)
			}
			break
		}
	}
	if !announceEnabled {
		if err = db.UpsertAnnounce(user.Fob, false, user.Nickname); err != nil {
			log.Println("Error upserting announce status=false", err)
		}
	}
}
