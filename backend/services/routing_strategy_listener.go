package services

import (
	"log"
	"time"

	"github.com/lib/pq"
)

const routingStrategiesNotifyChannel = "routing_strategies_changed"

// StartRoutingStrategiesListener subscribes to PostgreSQL NOTIFY so all pods reload strategy cache
// when routing_strategies changes (including direct SQL). The writing process also calls
// ReloadRoutingStrategies() for immediate consistency.
func StartRoutingStrategiesListener(dsn string) {
	if dsn == "" {
		return
	}
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("routing_strategies listener [%v]: %v", ev, err)
		}
	}
	const minReconnect = 10 * time.Second
	const maxReconnect = time.Minute
	listener := pq.NewListener(dsn, minReconnect, maxReconnect, reportProblem)
	if err := listener.Listen(routingStrategiesNotifyChannel); err != nil {
		log.Printf("routing_strategies LISTEN: %v", err)
		return
	}
	go func() {
		defer func() { _ = listener.Close() }()
		ticker := time.NewTicker(90 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case n := <-listener.Notify:
				if n != nil {
					GetSmartRouter().ReloadRoutingStrategies()
				}
			case <-ticker.C:
				if err := listener.Ping(); err != nil {
					log.Printf("routing_strategies listener ping: %v", err)
				}
			}
		}
	}()
}
