package models

import "time"

type Account struct {
	UUID      string    `db:"uuid" json:"uuid"`
	DiscordID string    `db:"discord_id" json:"discord_id"`
	CreatedAt time.Time `db:"created_at" json:"created_at"`
}
