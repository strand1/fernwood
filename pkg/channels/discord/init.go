package discord

import (
	"github.com/strand1/fernwood/pkg/bus"
	"github.com/strand1/fernwood/pkg/channels"
	"github.com/strand1/fernwood/pkg/config"
)

func init() {
	channels.RegisterFactory("discord", func(cfg *config.Config, b *bus.MessageBus) (channels.Channel, error) {
		return NewDiscordChannel(cfg.Channels.Discord, b)
	})
}
