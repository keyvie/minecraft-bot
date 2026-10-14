package state

import (
	"github.com/keyvie/minecraft-bot/internal/rpc"
)

type Public struct {
	RPC *rpc.Client
}

func New(rpc *rpc.Client) *Public {
	return &Public{
		RPC: rpc,
	}
}
