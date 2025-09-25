package main

import (
	"context"

	"os"

	"github.com/dtylman/saatool/actions"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:     "saatool",
		Version:  "0.1.0",
		Usage:    "saatool - a tool for working with translation projects",
		Commands: []*cli.Command{},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "deepseek-api-key",
				Aliases: []string{"key"},
				Usage:   "API key for DeepSeek.ai",
			},
		},
		EnableShellCompletion: true,
	}
	actions.AddAction(cmd, "import", &actions.EPubImportAction{})
	actions.AddAction(cmd, "import", &actions.PDFImportAction{})
	cmd.Run(context.Background(), os.Args)
}
