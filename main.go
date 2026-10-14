package main

import (
	"context"
	"database/sql"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/disgoorg/disgo"
	"github.com/disgoorg/disgo/bot"
	"github.com/disgoorg/disgo/gateway"

	"github.com/joho/godotenv"
	_ "modernc.org/sqlite"

	"github.com/keyvie/minecraft-bot/commands"
	"github.com/keyvie/minecraft-bot/internal/core"
	"github.com/keyvie/minecraft-bot/internal/modules"
	"github.com/keyvie/minecraft-bot/internal/rpc"
	"github.com/keyvie/minecraft-bot/internal/state"
	"github.com/keyvie/minecraft-bot/listeners"
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

	rpcClient := rpc.NewClient(rpc.WithReconnect(true), rpc.WithReconnectTimeout(30))
	var secret string
	if s, ok := os.LookupEnv("RPC_SECRET"); ok && s != "" {
		secret = s
	}
	if err = rpcClient.Connect(ctx, os.Getenv("RPC_URL"), rpc.WithSecret(secret)); err != nil {
		panic(err)
	}

	container := core.NewContainer(discordClient, core.WithState(*state.New(rpcClient)))
	core.AddHandlers(container)

	// Add all listeners
	if err = modules.AddListeners(container, listeners.Ready); err != nil {
		panic(err)
	}

	// Add all commands
	if err = modules.AddCommands(container,
		commands.Ping,
		commands.RPC,
		commands.Whitelist,
	); err != nil {
		panic(err)
	}

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

	if err := rpcClient.Close(shutdownCtx); err != nil {
		slog.Error("RPC", "Error closing", err)
	}

	if err := db.Close(); err != nil {
		slog.Error("Database", "Error closing", err)
	}
}
