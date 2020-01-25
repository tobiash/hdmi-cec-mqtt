package main

import (
	"bytes"
	"fmt"
	"github.com/chbmuc/cec"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"log"
	"net/url"
	"os"
)

func main()  {
	uri, err := url.Parse(os.Getenv("MQTT_URL"))
	if err != nil {
		log.Fatal(err)
	}

	options := mqtt.NewClientOptions()
	options.AddBroker(uri.String())

	client := mqtt.NewClient(options)

	c, err := cec.Open("", "cec.go")
	if err != nil {
		fmt.Println(err)
	}

	client.Subscribe("cec", 0, func(client mqtt.Client, message mqtt.Message) {
		c.Transmit(string(message.Payload()))
	})
	client.Subscribe("cec/power", 0, func(client mqtt.Client, message mqtt.Message) {
		if bytes.Equal(message.Payload(), []byte("on")) {
			c.Po
		}
	})

}