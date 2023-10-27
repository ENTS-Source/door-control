package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ents-source/door-control/amember"
	"github.com/ents-source/door-control/api"
	"github.com/ents-source/door-control/api/auth"
	"github.com/ents-source/door-control/assets"
	"github.com/ents-source/door-control/db"
	"github.com/ents-source/door-control/doors"
	"github.com/ents-source/door-control/matrix"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	HttpBind string `envconfig:"http_bind" default:"0.0.0.0:8080"`

	ApiSharedKey     string `envconfig:"api_shared_key"`
	ApiSharedKeyFile string `envconfig:"api_shared_key_file"`

	MqttUri          string `envconfig:"mqtt_uri" default:"tcp://127.0.0.1:1883"`
	MqttUser         string `envconfig:"mqtt_username"`
	MqttPassword     string `envconfig:"mqtt_password"`
	MqttPasswordFile string `envconfig:"mqtt_password_file"`
	MqttTopic        string `envconfig:"mqtt_topic" default:"rfid"`

	EspInterval  int `envconfig:"esp_ping_seconds" default:"10"`
	EspExpectNum int `envconfig:"esp_expect_num" default:"1"`

	AmpApiKey      string `envconfig:"amp_api_key"`
	AmpApiKeyFile  string `envconfig:"amp_api_key_file"`
	AmpApiUrl      string `envconfig:"amp_api_url"`
	AmpCategoryId  int    `envconfig:"amp_category_id"`
	AmpBufferDays  int    `envconfig:"amp_buffer_days" default:"3"`
	AmpResyncHours int    `envconfig:"amp_resync_hours" default:"4"`

	DbPath string `envconfig:"db_path" default:"./controller.db"`

	MxHomeserverUrl   string `envconfig:"matrix_hs_url"`
	MxUserId          string `envconfig:"matrix_user_id"`
	MxAccessToken     string `envconfig:"matrix_access_token"`
	MxAccessTokenFile string `envconfig:"matrix_access_token_file"`
	MxAnnounceRoomId  string `envconfig:"matrix_announce_room_id"`
	MxLogRoomId       string `envconfig:"matrix_log_room_id"`
}

func main() {
	var c config
	err := envconfig.Process("dc", &c)
	if err != nil {
		log.Fatal(err)
	}

	webPath := assets.SetupWeb()

	db.ConnectionString = c.DbPath

	auth.ApiAuthKey = getPassword(c.ApiSharedKey, c.ApiSharedKeyFile)

	doors.OfflineAfter = time.Duration(c.EspInterval) * time.Second
	doors.InstallApi()

	amember.ApiKey = getPassword(c.AmpApiKey, c.AmpApiKeyFile)
	amember.ApiRootUrl = c.AmpApiUrl
	amember.AccessBufferDays = c.AmpBufferDays
	amember.ProductCategoryId = c.AmpCategoryId
	amember.InstallApi()

	accessToken := getPassword(c.MxAccessToken, c.MxAccessTokenFile)
	matrix.LogRoomId = c.MxLogRoomId
	matrix.AnnounceRoomId = c.MxAnnounceRoomId
	if err = matrix.Connect(c.MxHomeserverUrl, c.MxUserId, accessToken); err != nil {
		log.Fatal(err)
	}

	mqttPassword := getPassword(c.MqttPassword, c.MqttPasswordFile)
	if err = doors.Connect(doors.MqttOptions{
		Uri:      c.MqttUri,
		Username: c.MqttUser,
		Password: mqttPassword,
		Topic:    c.MqttTopic,
	}); err != nil {
		log.Fatal(err)
	}

	doors.OnAccess = func(door string, fob string, time time.Time, granted bool) {
		var err2 error
		if err2 = matrix.LogAccess(door, fob, time, granted); err2 != nil {
			log.Println("Error logging door access record: ", err2)
		}

		if granted {
			if announce, nickname, err2 := db.IsAnnounceEnabled(fob); err2 != nil {
				log.Println("Error checking announcement status: ", err2)
			} else if announce && nickname != "" {
				if err2 = matrix.AnnounceAccess(door, nickname); err2 != nil {
					log.Println("Error announcing door access record: ", err2)
				}
			}
		}
	}

	wg := api.Start(c.HttpBind, webPath, api.HealthOptions{ExpectedDoors: c.EspExpectNum})

	timer := time.NewTicker(time.Duration(c.AmpResyncHours) * time.Hour)
	go func() {
		for {
			select {
			case <-timer.C:
				log.Println("Resyncing all users on timer")
				amember.ResyncAllUsers()
			}
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer close(stop)
		<-stop

		log.Println("Stopping timer...")
		timer.Stop()

		log.Println("Stopping doors...")
		doors.Disconnect()

		log.Println("Stopping api...")
		api.Stop()

		log.Println("Stopping matrix...")
		matrix.Stop()

		log.Println("Stopping database...")
		db.Stop()

		log.Println("Cleaning up...")
		_ = os.RemoveAll(webPath)

		log.Println("Done stopping")
	}()

	wg.Add(1)
	wg.Wait()

	log.Println("Goodbye!")
}

func getPassword(in string, f string) string {
	passwd := in
	if f != "" {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		passwd = string(b)
	}
	return passwd
}
