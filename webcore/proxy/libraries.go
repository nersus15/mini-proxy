package proxy

import (
	kafka "github.com/webcore-go/lib-kafka"
	memory "github.com/webcore-go/lib-memory"
	postgres "github.com/webcore-go/lib-postgres"
	"github.com/webcore-go/webcore/app/core"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"cache:memory":      &memory.MemoryLoader{},
	"database:postgres": &postgres.PostgresLoader{},
	"kafka:producer":    &kafka.KafkaProducerLoader{},
}
