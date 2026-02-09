package mqtt

import (
	"log"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"greenhouse/internal/model"
)

type Client struct {
	client mqtt.Client
}

func New(broker string, clientID string, topic string, handler func(model.Measurement)) *Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(broker)
	opts.SetClientID(clientID)
	opts.SetAutoReconnect(true)
	opts.SetConnectRetry(true)

	opts.OnConnect = func(c mqtt.Client) {
		log.Println("MQTT connected")

		if token := c.Subscribe(topic, 0, func(_ mqtt.Client, msg mqtt.Message) {
			m, err := ParseMeasurement(msg.Payload())
			if err != nil {
				log.Println("invalid payload:", err)
				return
			}

		handler(m)
		}); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}
	}

	client := mqtt.NewClient(opts)

	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	return &Client{client: client}
}
