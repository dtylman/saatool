package actions

import (
	"context"

	"github.com/urfave/cli/v3"
)

// Action represents a base class for a commandline subcommand
type Action interface {
	//Name the command name
	Name() string
	//Usage the command usage
	Usage() string
	//Flags defines flags
	Flags() []cli.Flag
	//Action is the action execution
	Action(context.Context, *cli.Command) error
}

// CreateCommandFromAction creates a cli command from an action
func CreateCommandFromAction(a Action) *cli.Command {
	return &cli.Command{
		Name:   a.Name(),
		Usage:  a.Usage(),
		Action: a.Action,
		Flags:  a.Flags(),
	}
}

// GetSection retrieves a command section by name
func GetSection(app *cli.Command, section string) *cli.Command {
	for i := range app.Commands {
		if app.Commands[i].Name == section {
			return app.Commands[i]
		}
	}
	return nil
}

// AddAction adds an action to a command group, creating the group if it does not exist
func AddAction(app *cli.Command, group string, action Action) {
	section := GetSection(app, group)
	if section == nil {
		section = &cli.Command{
			Name:     group,
			Usage:    "Commands for " + group,
			Commands: make([]*cli.Command, 0),
		}
		app.Commands = append(app.Commands, section)
	}
	section.Commands = append(section.Commands, CreateCommandFromAction(action))
}
