package main

import (
	"fmt"
	"github.com/mesia777/berith-utils/db"
	"github.com/mesia777/berith-utils/utils"
	"github.com/urfave/cli"
	"log"
	"os"
)

type App struct {
	cliApp *cli.App
	db     *db.Database
}

var (
	app = &App{
		cliApp: utils.NewApp(),
		db:     createDatabase(),
	}
)

func init() {
	app.cliApp.Action = func(ctx *cli.Context) error {
		return cli.ShowAppHelp(ctx)
	}

	app.cliApp.Commands = []cli.Command{
		nodeCommand,
		berithCommand,
	}
}

func main() {
	err := app.cliApp.Run(os.Args)
	defer app.db.Close()
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

// ShowSubCommand display sub commands help
func ShowSubCommand(ctx *cli.Context) error {
	return cli.ShowSubcommandHelp(ctx)
}

// createDataStore create a database
func createDatabase() *db.Database {
	path, err := utils.GetDatabasePath()
	if err != nil {
		log.Fatal("failed to create data store.", err)
		os.Exit(1)
	}
	database, err := db.NewDatabase(path, nil)
	if err != nil {
		log.Fatal("failed to create data store.", err)
		os.Exit(1)
	}
	return database
}
