package doors

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
)

type MqttOptions struct {
	Uri      string
	Username string
	Password string
	Topic    string
}

var client mqtt.Client

func waitMqtt(token mqtt.Token) error {
	finishedOk := token.WaitTimeout(10 * time.Second)
	if !finishedOk {
		return errors.New("timed out on action")
	}
	return token.Error()
}

func Connect(conn MqttOptions) error {
	opts := mqtt.NewClientOptions().
		AddBroker(conn.Uri).
		SetUsername(conn.Username).
		SetPassword(conn.Password).
		SetOrderMatters(false)
	client = mqtt.NewClient(opts)
	if err := waitMqtt(client.Connect()); err != nil {
		return errors.Join(errors.New("error during connect"), err)
	}
	if err := waitMqtt(client.Subscribe(fmt.Sprintf("%s/sync", conn.Topic), 2, onDoorSync)); err != nil {
		return errors.Join(errors.New("error during subscribe to sync"), err)
	}
	if err := waitMqtt(client.Subscribe(fmt.Sprintf("%s/send", conn.Topic), 2, onDoorSend)); err != nil {
		return errors.Join(errors.New("error during subscribe to send"), err)
	}
	if err := waitMqtt(client.Subscribe(fmt.Sprintf("%s", conn.Topic), 2, onDoorRoot)); err != nil {
		return errors.Join(errors.New("error during subscribe to root"), err)
	}
	return nil
}

func Disconnect() {
	client.Disconnect(100) // wait 100ms for work to finish
}

func onDoorSync(client mqtt.Client, message mqtt.Message) {
	log.Printf("[MQTT:Sync<<] %s %s", message.Topic(), message.Payload())

	msg := parseMessage(message.Payload())

	if t, err := readMessageVal[string](msg, "type"); err != nil {
		log.Println("[MQTT:Sync<<]", err)
		return
	} else {
		switch t {
		case "heartbeat":
			discoverDoor(msg)
			return
		}
	}
}

func onDoorSend(client mqtt.Client, message mqtt.Message) {
	log.Printf("[MQTT:Send<<] %s %s", message.Topic(), message.Payload())

}

func onDoorRoot(client mqtt.Client, message mqtt.Message) {
	log.Printf("[MQTT:Root<<] %s %s", message.Topic(), message.Payload())

}
