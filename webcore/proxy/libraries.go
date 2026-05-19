package proxy

import (
	cron "github.com/nersus15/lib-go-cron"
	kafka "github.com/webcore-go/lib-kafka"
	memory "github.com/webcore-go/lib-memory"
	mysql "github.com/webcore-go/lib-mysql"
	postgres "github.com/webcore-go/lib-postgres"
	"github.com/webcore-go/webcore/app/core"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"cache:memory":      &memory.MemoryLoader{},
	"database:postgres": &postgres.PostgresLoader{},
	"database:mysql":    &mysql.MysqlLoader{},
	"kafka:producer":    &kafka.KafkaProducerLoader{},
	"cron":              &cron.CronLoader{},
}
