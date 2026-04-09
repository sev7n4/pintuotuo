package services

import (
	"context"
	"log"
	"time"

	"github.com/lib/pq"
)

const platformSettingsNotifyChannel = "platform_settings_changed"

// StartPlatformSettingsListener 订阅 PostgreSQL NOTIFY，热更新健康调度等 platform_settings。
func StartPlatformSettingsListener(dsn string) {
	if dsn == "" {
		return
	}
	reportProblem := func(ev pq.ListenerEventType, err error) {
		if err != nil {
			log.Printf("platform_settings listener [%v]: %v", ev, err)
		}
	}
	const minReconnect = 10 * time.Second
	const maxReconnect = time.Minute
	listener := pq.NewListener(dsn, minReconnect, maxReconnect, reportProblem)
	if err := listener.Listen(platformSettingsNotifyChannel); err != nil {
		log.Printf("platform_settings LISTEN: %v", err)
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
					if err := ReloadPlatformSettingsCache(context.Background()); err != nil {
						log.Printf("platform_settings reload cache: %v", err)
					}
					GetHealthScheduler().SignalReload()
				}
			case <-ticker.C:
				if err := listener.Ping(); err != nil {
					log.Printf("platform_settings listener ping: %v", err)
				}
			}
		}
	}()
}
