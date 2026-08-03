package proxy

import (
	cron "github.com/nersus15/lib-go-cron"
	sqlite "github.com/nersus15/lib-sqlchiper"

	// kafka "github.com/webcore-go/lib-kafka"
	sentryLogger "github.com/nersus15/lib-go-sentry"
	memory "github.com/webcore-go/lib-memory"
	"github.com/webcore-go/webcore/app/core"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"cache:memory":    &memory.MemoryLoader{},
	"database:sqlite": &sqlite.SqliteLoader{},
	// "kafka:producer":  &kafka.KafkaProducerLoader{},
	"cron":      &cron.CronLoader{},
	"remotelog": &sentryLogger.SentryLoader{},
}
