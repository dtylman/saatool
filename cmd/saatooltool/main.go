package main

import (
	"context"
	"fmt"

	"os"

	"github.com/dtylman/saatool/actions"
	"github.com/dtylman/saatool/config"
	"github.com/urfave/cli/v3"
)

func main() {
	cmd := &cli.Command{
		Name:     "saatool",
		Version:  config.Version,
		Usage:    "saatool - a tool for working with translation projects",
		Commands: []*cli.Command{},
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:    "api-key",
				Aliases: []string{"key"},
				Usage:   "API key for DeepSeek.ai",
			},
			&cli.StringFlag{
				Name:  "ai-vendor",
				Usage: "AI vendor to use for translation (e.g. 'deepseek', 'gemini', 'ollama')",
				Value: "deepseek",
			},
			&cli.StringFlag{
				Name:  "ai-model",
				Usage: "Model name to use. Leave empty to use the vendor's default.",
				Value: "deepseek-v4-flash",
			},
		},
		EnableShellCompletion: true,
	}
	actions.AddAction(cmd, "import", &actions.EPubImportAction{})
	actions.AddAction(cmd, "import", &actions.PDFImportAction{})
	actions.AddAction(cmd, "export", &actions.RTFExportAction{})
	err := cmd.Run(context.Background(), os.Args)
	if err != nil {
		fmt.Println("Error:", err)
		os.Exit(1)
	}
}
