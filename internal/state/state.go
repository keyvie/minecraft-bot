package state

import (
	"github.com/keyvie/minecraft-bot/internal/database"
	"github.com/keyvie/minecraft-bot/internal/database/models"
	"github.com/keyvie/minecraft-bot/internal/mojang"
	"github.com/keyvie/minecraft-bot/internal/rpc"
)

type Repos struct {
	Accounts *database.Repository[models.Account]
}

type Public struct {
	RPC    *rpc.Client
	Mojang *mojang.Client
	Repos  Repos
}

func New(rpc *rpc.Client, m *mojang.Client, repos Repos) *Public {
	return &Public{
		RPC:    rpc,
		Repos:  repos,
		Mojang: m,
	}
}
