package main

import (
	"context"
	"fmt"
	"github.com/rs/zerolog/log"
	"time"
)

type ClientPool struct {
	ClientConfig ClientConfig
	ClientCount  int
	Context      context.Context

	clients []Client
	events  chan ClientEvent
	stats   *ClientPoolStats
}

type ClientConfig struct {
	ClientId string
	Username string
	Password string
	Url      string

	MessageDelay time.Duration
	Topic        string
	PayloadType  int
	Payload      string
}

type ClientPoolStats struct {
	TotalMessagesSent         int
	TotalMessagesReceived     int
	MessagesReceivedPerSecond int
	MessagesSentPerSecond     int
}

func CreateClientPool(config ClientConfig, clientCount int, poolContext context.Context) ClientPool {
	pool := ClientPool{
		ClientConfig: config,
		ClientCount:  clientCount,
		Context:      poolContext,

		clients: []Client{},
		events:  make(chan ClientEvent),
		stats: &ClientPoolStats{
			TotalMessagesSent:         0,
			TotalMessagesReceived:     0,
			MessagesReceivedPerSecond: 0,
			MessagesSentPerSecond:     0,
		},
	}

	for i := 0; i < pool.ClientCount; i++ {
		clientContext := context.WithValue(poolContext, "id", pool.ClientConfig.ClientId)
		client := Client{
			Id:           fmt.Sprintf("%s-%d", pool.ClientConfig.ClientId, i),
			Username:     pool.ClientConfig.Username,
			Password:     pool.ClientConfig.Password,
			Url:          pool.ClientConfig.Url,
			MessageDelay: pool.ClientConfig.MessageDelay,
			Topics: []string{
				fmt.Sprintf("stress/%s-%d", pool.ClientConfig.ClientId, i),
			},
			PayloadType: pool.ClientConfig.PayloadType,
			Payload:     pool.ClientConfig.Payload,
			Context:     clientContext,

			events: pool.events,
		}

		pool.clients = append(pool.clients,
			client.Setup().
				Connect().
				Publish().
				Subscribe(),
		)
	}

	return pool
}

func (pool ClientPool) RunStats() {
	go pool.monitorStats()
	pool.outputStats()
}

func (pool ClientPool) monitorStats() {
	for {
		event := <-pool.events
		switch event.Event {
		case "MessagePublished":
			pool.stats.TotalMessagesSent += event.Data.(int)
		case "MessageReceived":
			pool.stats.TotalMessagesReceived += event.Data.(int)
		default:
			log.Warn().
				Str("broker_event", event.Event).
				Msg("unhandled broker event")
		}
	}
}

func (pool ClientPool) outputStats() {
	for {
		lastTotalMessagesReceived := pool.stats.TotalMessagesReceived
		lastTotalMessagesSent := pool.stats.TotalMessagesSent
		select {
		case <-pool.Context.Done():
			log.Info().
				Msg("finished..")
			return
		default:
			log.Info().
				Int("total_brokers", pool.ClientCount).
				Int("total_messages_sent", pool.stats.TotalMessagesSent).
				Int("total_messages_received", pool.stats.TotalMessagesReceived).
				Int("messages_sent_second", pool.stats.MessagesSentPerSecond).
				Int("messages_received_second", pool.stats.MessagesReceivedPerSecond).
				Msg("broker stats")
		}
		time.Sleep(time.Second)

		pool.stats.MessagesSentPerSecond = pool.stats.TotalMessagesSent - lastTotalMessagesSent
		pool.stats.MessagesReceivedPerSecond = pool.stats.TotalMessagesReceived - lastTotalMessagesReceived
	}
}
