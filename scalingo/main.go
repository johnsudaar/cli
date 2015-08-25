package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/Scalingo/cli/Godeps/_workspace/src/github.com/Scalingo/codegangsta-cli"
	"github.com/Scalingo/cli/Godeps/_workspace/src/github.com/stvp/rollbar"
	"github.com/Scalingo/cli/cmd"
	"github.com/Scalingo/cli/cmd/autocomplete"
	"github.com/Scalingo/cli/config"
	"github.com/Scalingo/cli/signals"
	"github.com/Scalingo/cli/update"
)

func DefaultAction(c *cli.Context) {
	completeMode := false

	for i := range os.Args {
		if strings.Contains(os.Args[i], "generate-bash-completion") {
			completeMode = true
			break
		}
	}

	if !completeMode {
		cmd.HelpCommand.Action(c)
	} else {
		i := len(os.Args) - 2
		if i > 0 {
			autocomplete.FlagsAutoComplete(c, os.Args[i])
		}
	}
}

func ScalingoAppComplete(c *cli.Context) {

	for _, flag := range c.App.Flags {
		names := strings.Split(cli.GetFlagName(flag), ",")
		for i := range names {
			if i == 0 {
				fmt.Fprintln(c.App.Writer, "--"+names[i])
			} else {
				fmt.Fprintln(c.App.Writer, "-"+strings.TrimSpace(names[i]))
			}
		}
	}

	for _, command := range c.App.Commands {
		for _, name := range command.Names() {
			fmt.Fprintln(c.App.Writer, name)
		}
	}
}

func main() {
	app := cli.NewApp()
	app.Name = "Scalingo Client"
	app.Author = "Scalingo Team"
	app.Email = "hello@scalingo.com"
	app.Usage = "Manage your apps and containers"
	app.Version = config.Version
	app.CategorizedHelp = true
	app.Flags = []cli.Flag{
		cli.StringFlag{Name: "app, a", Value: "<name>", Usage: "Name of the app", EnvVar: "SCALINGO_APP"},
		cli.StringFlag{Name: "remote, r", Value: "scalingo", Usage: "Name of the remote", EnvVar: ""},
	}
	app.EnableBashCompletion = true
	app.BashComplete = func(c *cli.Context) {
		ScalingoAppComplete(c)
	}
	app.Action = DefaultAction

	// Commands
	for _, command := range cmd.Commands {
		oldFunc := command.BashComplete
		command.BashComplete = func(c *cli.Context) {
			n := len(os.Args) - 2
			if n > 0 && !autocomplete.FlagsAutoComplete(c, os.Args[n]) && oldFunc != nil {
				oldFunc(c)
			}
		}
		app.Commands = append(app.Commands, command)
	}

	go signals.Handle()

	bashComplete := false
	for i := range os.Args {
		if strings.Contains(os.Args[i], "generate-bash-completion") {
			bashComplete = true
		}
	}

	if !bashComplete {
		if len(os.Args) >= 2 && os.Args[1] == cmd.UpdateCommand.Name {
			err := update.Check()
			if err != nil {
				rollbar.Error(rollbar.ERR, err)
			}
			return
		} else {
			defer update.Check()
		}
	}

	if err := app.Run(os.Args); err != nil {
		fmt.Println("Fail to run scalingo", err)
	}
}
