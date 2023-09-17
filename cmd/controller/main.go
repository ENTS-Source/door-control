package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ents-source/door-control/amember"
	"github.com/ents-source/door-control/api"
	"github.com/ents-source/door-control/assets"
	"github.com/ents-source/door-control/doors"
	"github.com/kelseyhightower/envconfig"
)

type config struct {
	HttpBind string `envconfig:"http_bind" default:"0.0.0.0:8080"`

	MqttUri          string `envconfig:"mqtt_uri" default:"tcp://127.0.0.1:1883"`
	MqttUser         string `envconfig:"mqtt_username"`
	MqttPassword     string `envconfig:"mqtt_password"`
	MqttPasswordFile string `envconfig:"mqtt_password_file"`
	MqttTopic        string `envconfig:"mqtt_topic" default:"rfid"`

	EspInterval  int `envconfig:"esp_ping_seconds" default:"10"`
	EspExpectNum int `envconfig:"esp_expect_num" default:"1"`

	AmpApiKey     string `envconfig:"amp_api_key"`
	AmpApiKeyFile string `envconfig:"amp_api_key_file"`
	AmpApiUrl     string `envconfig:"amp_api_url"`
	AmpCategoryId int    `envconfig:"amp_category_id"`
}

func main() {
	var c config
	err := envconfig.Process("dc", &c)
	if err != nil {
		log.Fatal(err)
	}

	webPath := assets.SetupWeb()

	doors.OfflineAfter = time.Duration(c.EspInterval) * time.Second

	amember.ApiKey = getPassword(c.AmpApiKey, c.AmpApiKeyFile)
	amember.ApiRootUrl = c.AmpApiUrl
	amember.InstallApi(c.AmpCategoryId)

	mqttPassword := getPassword(c.MqttPassword, c.MqttPasswordFile)
	if err = doors.Connect(doors.MqttOptions{
		Uri:      c.MqttUri,
		Username: c.MqttUser,
		Password: mqttPassword,
		Topic:    c.MqttTopic,
	}); err != nil {
		log.Fatal(err)
	}

	wg := api.Start(c.HttpBind, webPath, api.HealthOptions{ExpectedDoors: c.EspExpectNum})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	go func() {
		defer close(stop)
		<-stop

		log.Println("Stopping doors...")
		doors.Disconnect()

		log.Println("Stopping api...")
		api.Stop()

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