package main

import (
	"fmt"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"path"
	"strings"
)

type PrefixRouter struct {
	client mqtt.Client
	prefix string
}

func (r *PrefixRouter) AddRoute(pattern string, handler func (client mqtt.Client, message *ParsedMessage))  {
	parse := parseWildcards(pattern)
	r.client.AddRoute(path.Join(r.prefix, pattern), func(client mqtt.Client, message mqtt.Message) {
		match, plus, rest := parse(strings.TrimPrefix(message.Topic(), r.prefix + "/"))
		if !match {
			panic(fmt.Sprintf("pattern didnt match: %q", message.Topic()))
		}
		handler(client, &ParsedMessage{message, plus, rest})
	})
}

func (r *PrefixRouter) Subscribe(qos byte) {
	r.client.Subscribe(path.Join(r.prefix, "#"), qos, nil)
}

type ParsedMessage struct {
	mqtt.Message
	plus []string
	rest []string
}

func (p *ParsedMessage) Plus(idx int) string {
	return p.plus[idx]
}

func (p *ParsedMessage) Rest(idx int) string {
	return p.rest[idx]
}

func parseWildcards(pattern string) func (topic string) (bool, []string, []string) {
	parts := strings.Split(pattern, "/")
	return func(topic string) (bool, []string, []string) {
		topicParts := strings.Split(topic, "/")
		var plus []string
		if len(topicParts) != len(parts) && !(parts[len(parts) - 1] != "#") {
			return false, nil, nil
		}
		for idx := range topicParts {
			switch parts[idx] {
			case "+":
				plus = append(plus, topicParts[idx])
			case "#":
				return true, plus, topicParts[idx:]
			default:
				if topicParts[idx] != parts[idx] {
					return false, nil, nil
				}
			}
		}
		return true, plus, nil
	}
}