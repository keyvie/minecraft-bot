package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"strings"
	"time"
	"os"
	"os/signal"
	"syscall"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"github.com/keyvie/minecraft-bot/internal/core"
	"github.com/keyvie/minecraft-bot/internal/modules"
	"github.com/keyvie/minecraft-bot/listeners"
	"github.com/keyvie/minecraft-bot/commands"
)

func main() {
	// Load env values from .env if it exists
	_ = godotenv.Load()

	// Setup logging
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if strings.ToLower(os.Getenv("DEBUG")) == "true" {
		slog.SetLogLoggerLevel(slog.LevelDebug)
	} else {
		slog.SetLogLoggerLevel(slog.LevelInfo)
	}

	// Listen for signals to shutdown
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Open database connection
	db, err := sql.Open("sqlite", "database.db")
	if err != nil {
		panic(err)
	}
	//defer db.Close()

	// Open the discord client
	discordClient, err := disgo.New(os.Getenv("DISCORD_TOKEN"),
		bot.WithGatewayConfigOpts(
			gateway.WithIntents(
				gateway.IntentsAll,
			),
		),
	)
	if err != nil {
		panic(err)
	}
	//defer discordClient.Close(context.Background())

	container := core.NewContainer(discordClient)
	core.AddHandlers(container)

	// Add all listeners
	modules.Listeners.Add(container, listeners.Ready)

	// Add all commands
	modules.Commands.Add(container, commands.Ping)

	// Open connection to Discord
	if err = discordClient.OpenGateway(context.Background()); err != nil {
		panic(err)
	}

	// What gets ran after shutdown is triggered
	<-ctx.Done()
	stop()

	// Make a context for the shutdown process
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	discordClient.Close(shutdownCtx)

	if err := db.Close(); err != nil {
		slog.Error("Error closing database:", err)
	}
}