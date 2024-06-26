package cmd

import (
	"github.com/urfave/cli/v2"

	"github.com/Scalingo/cli/apps"
	"github.com/Scalingo/cli/cmd/autocomplete"
	"github.com/Scalingo/cli/detect"
)

var (
	psCommand = cli.Command{
		Name:     "ps",
		Category: "App Management",
		Usage:    "Display your application containers",
		Flags:    []cli.Flag{&appFlag},
		Description: `Display your application containers
	Example
	  'scalingo --app my-app ps'`,
		Action: func(c *cli.Context) error {
			currentApp := detect.CurrentApp(c)
			if c.Args().Len() != 0 {
				cli.ShowCommandHelp(c, "ps")
				return nil
			}

			err := apps.Ps(c.Context, currentApp)
			if err != nil {
				errorQuit(err)
			}
			return nil
		},
		BashComplete: func(c *cli.Context) {
			autocomplete.CmdFlagsAutoComplete(c, "ps")
		},
	}
)
