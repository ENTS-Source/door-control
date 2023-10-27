package doors

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/eclipse/paho.mqtt.golang"
	"github.com/ents-source/door-control/db"
)

type MqttOptions struct {
	Uri      string
	Username string
	Password string
	Topic    string
}

var client mqtt.Client
var sendTopic string

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
	sendTopic = conn.Topic
	return nil
}

func Disconnect() {
	client.Disconnect(100) // wait 100ms for work to finish
}

func sendCommand(cmd map[string]any) error {
	b, err := json.Marshal(cmd)
	if err != nil {
		return err
	}

	log.Println("[MQTT>>] ", string(b))
	err = waitMqtt(client.Publish(sendTopic, 1, false, string(b)))
	return err
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

	msg := parseMessage(message.Payload())

	if c, err := readMessageVal[string](msg, "cmd"); err != nil {
		log.Println("[MQTT:Send<<] [Command Parse Error]", err)
		return
	} else if c == "log" {
		if t, err := readMessageVal[string](msg, "type"); err != nil {
			log.Println("[MQTT:Send<<] [Type Parse Error]", err)
			return
		} else if t == "access" {
			ts, err := readMessageVal[float64](msg, "time")
			if err != nil {
				log.Println("[MQTT:Send<<] [TS Parse Error]", err)
				return
			}

			fob, err := readMessageVal[string](msg, "uid")
			if err != nil {
				log.Println("[MQTT:Send<<] [Fob Parse Error]", err)
				return
			}

			door, err := readMessageVal[string](msg, "door")
			if err != nil {
				log.Println("[MQTT:Send<<] [Door Parse Error]", err)
				return
			}

			accessStr, err := readMessageVal[string](msg, "access")
			if err != nil {
				log.Println("[MQTT:Send<<] [Access Parse Error]", err)
				return
			}

			access := accessStr == "Always"

			err = db.InsertAccess(door, fob, time.UnixMilli(int64(ts)*1000), access)
			if err != nil {
				log.Println("[MQTT:Send<<] [DB Insert Error]", err)
				return
			}

			if OnAccess != nil {
				OnAccess(door, fob, time.UnixMilli(int64(ts)*1000), access)
			}

			log.Println("Access record stored in database")
		}
	}
}

func onDoorRoot(client mqtt.Client, message mqtt.Message) {
	log.Printf("[MQTT:Root<<] %s %s", message.Topic(), message.Payload())

}
