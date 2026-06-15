package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	sqlite "github.com/nersus15/lib-sqlchiper"
	memory "github.com/webcore-go/lib-memory"
	mysql "github.com/webcore-go/lib-mysql"
	postgres "github.com/webcore-go/lib-postgres"
	sql "github.com/webcore-go/lib-sql"
	"github.com/webcore-go/webcore/app/core"
	"github.com/webcore-go/webcore/app/out"
	"github.com/webcore-go/webcore/infra/config"
)

var APP_LIBRARIES = map[string]core.LibraryLoader{
	"cache:memory":      &memory.MemoryLoader{},
	"database:postgres": &postgres.PostgresLoader{},
	"database:mysql":    &mysql.MysqlLoader{},
	"database:sqlite":   &sqlite.SqliteLoader{},
}

func main() {
	if len(os.Args) < 2 {
		log.Fatal("Usage: migrate [mysql | sqlite | postgres] [command]")
		os.Exit(1)
	}

	dialect := os.Args[1]
	validDB := map[string]bool{
		"mysql":    true,
		"sqlite":   true,
		"postgres": true,
	}

	if !validDB[dialect] {
		log.Fatal("Usage: migrate [mysql | sqlite | postgres]")
		os.Exit(1)
	}

	var flags = flag.NewFlagSet("migrate", flag.ExitOnError)
	var dir *string
	switch dialect {
	case "mysql":
		dir = flags.String("dir", "webcore/init/migrations/proxy/mysql", "direktori file migrasi")
	case "sqlite":
		dir = flags.String("dir", "webcore/init/migrations/proxy/sqlite", "direktori file migrasi")
	default:
		dir = flags.String("dir", "webcore/init/migrations/proxy/postgres", "direktori file migrasi")
	}
	command := os.Args[2]
	flags.Parse(os.Args[3:])
	args := flags.Args()

	ctx := context.Background()

	// Load configuration
	cfg := config.Config{}
	if err := config.LoadDefaultConfig(&cfg); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	out.SetEnvironment(cfg.App.Environment)

	// Reset Config
	cfg.Redis.Host = ""
	cfg.Kafka.Brokers = []string{}

	// Initialize application
	application := core.NewApp(ctx, &cfg, APP_LIBRARIES, []core.Module{})

	// Start the application
	if err := application.Context.Start(); err != nil {
		log.Fatalf("Failed to start migration Proxy [%s]: %v", dialect, err)
		os.Exit(1)
	}

	lib, ok := core.Instance().Context.GetDefaultSingletonInstance("database")
	if !ok {
		log.Fatal("Gagal memuat instance database")
		os.Exit(1)
	}

	db := lib.(*sql.SQLDatabase)
	db.StartMigration(ctx, "proxy", command, *dir, args)

	fmt.Printf("Migration %s %s run successfully\n", "proxy", command)
}
