package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"github.com/brianvoe/gofakeit"
	"github.com/rs/zerolog/log"
	"time"
)

type AppConfig struct {
	Help         bool
	Username     string
	Password     string
	Url          string
	ClientCount  int
	MessageDelay time.Duration
	RunFor       time.Duration

	IncrementPayload    bool
	CustomPayload       string
	CustomPayloadFields string
}

const (
	StaticPayload    = 1
	GeneratedPayload = 2
	IncrementPayload = 3
)

func setupConfig() (AppConfig, error) {
	config := AppConfig{}
	flag.BoolVar(&config.Help, "help", false, "print application options")
	flag.StringVar(&config.Username, "username", "", "username to authenticate with mqtt broker")
	flag.StringVar(&config.Password, "password", "", "password to authenticate with mqtt broker")
	flag.StringVar(&config.Url, "url", "", "scheme://host:port Where \"scheme\" is one of \"tcp\", \"ssl\", or \"ws\", \"host\" is the ip-address (or hostname) and \"port\" is the port on which the broker is accepting connections.")
	flag.IntVar(&config.ClientCount, "workers", 10, "amount of clients to connect to mqtt broker with")
	delayString := flag.String("delay", "500ms", "delay per message per worker(s), e.g. 500ms or 1s or 1m")
	runString := flag.String("run", "15s", "length of time to run application before exiting")

	flag.BoolVar(&config.IncrementPayload, "increment", false, "payload is incremented for each message sent, each worker has an independent count (overrides payload flag)")
	flag.StringVar(&config.CustomPayload, "payload", "", "custom payload sent for each message, if left random payloads will be generated for each message")
	flag.StringVar(&config.CustomPayloadFields, "fields", "", "<name>:<type> custom fields to generate structs from seperated by a \",\" where type is number, string, id, phone or email. e.g. customer:id,customer_email:email (overrides payload flag)")

	flag.Parse()

	runFor, runErr := time.ParseDuration(*runString)
	if runErr != nil {
		flag.Usage()
		log.Error().Err(runErr).Msg("invalid run duration")
	}
	config.RunFor = runFor

	messageDelay, delayErr := time.ParseDuration(*delayString)
	if delayErr != nil {
		flag.Usage()
		log.Error().Err(delayErr).Msg("invalid message delay duration")
		return config, fmt.Errorf("invalid message delay duration")
	}
	config.MessageDelay = messageDelay

	if config.Url == "" {
		flag.Usage()
		return config, errors.New("missing url")
	}

	return config, nil
}

func main() {
	config, err := setupConfig()
	if err != nil {
		log.Error().Err(err).Msg("error setting up config")
		return
	}

	timeoutContext, cancelTimeout := context.WithTimeout(context.Background(), config.RunFor)
	defer cancelTimeout()

	log.Info().Msgf("running for %s", config.RunFor)

	gofakeit.Seed(time.Now().UnixMicro())

	payload, payloadType := configPayload(config)
	pool := CreateClientPool(ClientConfig{
		ClientId: "mqtt-stress-worker",
		Username: config.Username,
		Password: config.Password,
		Url:      config.Url,

		MessageDelay: config.MessageDelay,

		PayloadType: payloadType,
		Payload:     payload,
	}, config.ClientCount, timeoutContext)

	pool.RunStats()
}

func configPayload(config AppConfig) (string, int) {
	if config.IncrementPayload {
		return "0", IncrementPayload
	}
	if config.CustomPayloadFields != "" {
		return config.CustomPayloadFields, GeneratedPayload
	}
	if config.CustomPayload != "" {
		return config.CustomPayload, StaticPayload
	}
	return "mqtt-stress", StaticPayload
}
