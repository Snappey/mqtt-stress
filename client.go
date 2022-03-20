package main

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/brianvoe/gofakeit/v6"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/rs/zerolog/log"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	Id       string
	Username string
	Password string
	Url      string

	MessageDelay time.Duration
	Topics       []string
	PayloadType  int
	Payload      string
	Context      context.Context

	mqtt   mqtt.Client
	events chan ClientEvent
}

type ClientEvent struct {
	Event string
	Data  interface{}
}

func (c Client) Setup() Client {
	opts := mqtt.NewClientOptions()
	opts.AddBroker(c.Url)
	opts.SetClientID(c.Id)

	if c.Username != "" {
		opts.SetUsername(c.Username)
	}
	if c.Password != "" {
		opts.SetPassword(c.Password)
	}
	opts.SetOrderMatters(false)

	c.mqtt = mqtt.NewClient(opts)

	return c
}

func (c Client) Connect() Client {
	if token := c.mqtt.Connect(); token.Wait() && token.Error() != nil {
		log.Error().
			Err(token.Error()).
			Interface("client", c).
			Msg("error connecting to broker")
	}
	return c
}

func (c Client) Subscribe() Client {
	for _, topic := range c.Topics {
		c.mqtt.Subscribe(topic, 0, func(_ mqtt.Client, message mqtt.Message) {
			c.events <- ClientEvent{
				Event: "MessageReceived",
				Data:  1,
			}
		})
	}
	return c
}

func (c Client) Publish() Client {
	for _, topic := range c.Topics {
		go func(c *Client, topic string) {
			for {
				token := c.mqtt.Publish(topic, 0, false, c.getPayload(c.Payload, c.PayloadType))
				go c.waitTokenResponse(token)

				time.Sleep(c.MessageDelay)

				select {
				case <-c.Context.Done():
					c.mqtt.Disconnect(5)
					return
				default:
					continue
				}
			}
		}(&c, topic)
	}
	return c
}

func (c Client) waitTokenResponse(token mqtt.Token) {
	token.Wait()
	err := token.Error()
	if err != nil {
		log.Error().
			Err(err).
			Interface("client", c).
			Msg("error publishing message")
	} else {
		c.events <- ClientEvent{
			Event: "MessagePublished",
			Data:  1,
		}
	}
}

func (c Client) parsePayload(payload string) (map[string]string, error) {
	res := map[string]string{}
	fields := strings.Split(payload, ",")
	for _, field := range fields {
		data := strings.Split(field, ":")
		if len(data) != 2 {
			return nil, errors.New("failed to parse custom payload field, expected to find key and type seperated by \":\" ")
		}
		res[data[0]] = data[1]
	}
	return res, nil
}

func (c Client) generatePayload(template map[string]string) []byte {
	for key, pType := range template {
		switch pType {
		case "number":
			template[key] = strconv.Itoa(gofakeit.Number(1, 9999))
		case "string":
			template[key] = gofakeit.Username()
		case "id":
			template[key] = gofakeit.UUID()
		case "phone":
			template[key] = gofakeit.Phone()
		case "email":
			template[key] = gofakeit.Email()
		default:
			template[key] = pType
		}
	}
	resp, err := json.Marshal(template)
	if err != nil {
		log.Warn().Msg("failed to marshal generated json structure")
		return make([]byte, 0)
	}
	return resp
}

func (c *Client) getPayload(payload string, payloadType int) string {
	switch payloadType {
	case IncrementPayload:
		num, _ := strconv.Atoi(payload)
		c.Payload = strconv.Itoa(num + 1)
		return c.Payload
	case StaticPayload:
		return payload
	case GeneratedPayload:
		parsedPayload, _ := c.parsePayload(payload)
		return string(c.generatePayload(parsedPayload))
	default:
		return payload
	}
}
