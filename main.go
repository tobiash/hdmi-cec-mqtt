package main

import (
	"bytes"
	"encoding/json"
	"github.com/caarlos0/env/v6"
	"github.com/chbmuc/cec"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"net/url"
	"path"
	"strconv"
	"time"
)

type config struct {
	MqttUsername string  `env:"MQTT_USERNAME"`
	MqttPassword string  `env:"MQTT_PASSWORD"`
	MqttUrl      url.URL `env:"MQTT_URL"`
	MqttTopic    string  `env:"MQTT_TOPIC" envDefault:"cec"`
	CecDevice    string  `env:"CEC_DEVICE" envDefault:"/dev/ttyACM0"`
}

func main() {
	//stdlog.SetOutput(ioutil.Discard)
	var cfg config
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	if err := env.Parse(&cfg); err != nil {
		log.Fatal().Err(err).Msg("error reading configuration")
	}

	options := mqtt.NewClientOptions()
	options.AddBroker(cfg.MqttUrl.String())
	options.SetUsername(cfg.MqttUsername)
	options.SetPassword(cfg.MqttPassword)
	options.SetAutoReconnect(true)
	client := mqtt.NewClient(options)

	if t := client.Connect(); t.Wait() && t.Error() != nil {
		log.Fatal().Err(t.Error()).Msg("error connecting mqtt server")
	}
	log.Info().Str("MQTT_BROKER", cfg.MqttUrl.String()).Msg("connected to mqtt broker")

	c, err := cec.Open("", cfg.CecDevice)
	if err != nil {
		log.Fatal().Str("CEC_DEVICE", cfg.CecDevice).Err(err).Msg("error opening cec device")
	}
	log.Info().Str("CEC_DEVICE", cfg.CecDevice).Msg("connected to cec device")

	ticker := time.NewTicker(10 * time.Second)

	router := &PrefixRouter{client, cfg.MqttTopic }

	router.AddRoute("transmit", func(client mqtt.Client, message *ParsedMessage) {
		log.Info().Str("op", "transmit").Bytes("payload", message.Payload()).Msg("transmitting cec command")
		c.Transmit(string(message.Payload()))
	})

	router.AddRoute("mute", func(client mqtt.Client, message *ParsedMessage) {
		l := log.With().Str("op", "mute").Logger()
		l.Info().Msg("muting")
		if err := c.Mute(); err != nil {
			l.Err(err).Msg("error muting")
		}
	})

	router.AddRoute("key/+", func(client mqtt.Client, message *ParsedMessage) {
		addr, err := strconv.ParseInt(message.Plus(0), 10, 64)
		l := log.With().Str("op", "key").
			Str("address", message.Plus(0)).
			Bytes("payload", message.Payload()).
			Logger()
		if err != nil {
			l.Err(err).Msg("invalid device address")
			return
		}
		l.Info().Msg("keypress")
		c.Key(int(addr), string(message.Payload()))
	})

	router.AddRoute("power/+", func(client mqtt.Client, message *ParsedMessage) {
		addr, err := strconv.ParseInt(message.Plus(0), 10, 64)
		var op func(int) error
		var opName string
		if bytes.Equal(message.Payload(), []byte("on")) {
			op = c.PowerOn
			opName = "powerOn"
		} else {
			op = c.Standby
			opName = "standby"
		}

		l := log.With().Str("addr", message.Plus(0)).
			Str("op", opName).
			Logger()
		l.Info().Msg("switching power status")
		if err != nil {
			l.Err(err).Msg("error parsing device address")
			return
		}

		if err := op(int(addr)); err != nil {
			l.Err(err).Msg("error setting power status")
		}
	})

	router.AddRoute("volume", func(client mqtt.Client, message *ParsedMessage) {
		var err error
		l := log.With().Str("op", "volume").
			Bytes("payload", message.Payload()).
			Logger()
		l.Info().Msg("changing volume")
		if bytes.Equal(message.Payload(), []byte("up")) {
			err = c.VolumeUp()
		} else {
			err = c.VolumeDown()
		}
		if err != nil {
			l.Err(err).Msg("error changing volume")
		}
	})

	router.Subscribe(0)

	for range ticker.C {
		l := log.With().Str("op", "ticker").Logger()
		if !client.IsConnected() {
			if t := client.Connect(); t.Error() != nil {
				l.Fatal().Err(t.Error()).Msg("connection error")
			}
		}
		d, err := json.Marshal(c.List())
		if err != nil {
			l.Err(err).Msg("error marshaling device list")
			continue
		}
		t := client.Publish(path.Join(cfg.MqttTopic, "list"), 0, true, d)
		if t.Wait() && t.Error() != nil {
			l.Err(t.Error()).Msg("error publishing list")
		}
	}
}

