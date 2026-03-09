// PicoClaw - Ultra-lightweight personal AI agent
// Inspired by and based on nanobot: https://github.com/HKUDS/nanobot
// License: MIT
//
// Copyright (c) 2026 PicoClaw contributors

package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/strand1/fernwood/cmd/fernwood/internal"
	"github.com/strand1/fernwood/cmd/fernwood/internal/agent"
	"github.com/strand1/fernwood/cmd/fernwood/internal/auth"
	"github.com/strand1/fernwood/cmd/fernwood/internal/cron"
	"github.com/strand1/fernwood/cmd/fernwood/internal/gateway"
	"github.com/strand1/fernwood/cmd/fernwood/internal/migrate"
	"github.com/strand1/fernwood/cmd/fernwood/internal/onboard"
	"github.com/strand1/fernwood/cmd/fernwood/internal/skills"
	"github.com/strand1/fernwood/cmd/fernwood/internal/status"
	"github.com/strand1/fernwood/cmd/fernwood/internal/version"
)

func NewFernwoodCommand() *cobra.Command {
	short := fmt.Sprintf("%s fernwood - Agentic Coding Harness v%s\n\n", internal.Logo, internal.GetVersion())

	cmd := &cobra.Command{
		Use: "fernwood",
		Short:   short,
		Example: "fernwood version",
	}

	cmd.AddCommand(
		onboard.NewOnboardCommand(),
		agent.NewAgentCommand(),
		auth.NewAuthCommand(),
		gateway.NewGatewayCommand(),
		status.NewStatusCommand(),
		cron.NewCronCommand(),
		migrate.NewMigrateCommand(),
		skills.NewSkillsCommand(),
		version.NewVersionCommand(),
	)

	return cmd
}

const (
	colorGreen = "\033[1;38;2;34;197;94m"
	banner     = "\r\n" +
		colorGreen + "╔═╗┌─┐┬─┐┌┐┌┬ ┬┌─┐┌─┐┌┬┐\n" +
		colorGreen + "╠╣ ├┤ ├┬┘│││││││ ││ │ ││\n" +
		colorGreen + "╚  └─┘┴└─┘└┘└┴┘└─┘└─┘─┴┘\n" +
		"\033[0m\r\n"
)

func main() {
	fmt.Printf("%s", banner)
	cmd := NewFernwoodCommand()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
