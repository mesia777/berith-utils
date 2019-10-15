// Package utils contains internal helper functions for commands.
package utils

import (
	"github.com/urfave/cli"
	"os/user"
	"path/filepath"
)

var (
	PathFlag = cli.StringFlag{
		Name:  "path",
		Usage: "path of config file.",
	}
	NodeNameFlag = cli.StringFlag{
		Name:  "name",
		Usage: "name of a node",
	}
	HostUserFlag = cli.StringFlag{
		Name:  "host.user",
		Usage: "host username for ssh.",
	}
	HostAddressFlag = cli.StringFlag{
		Name:  "host.address",
		Usage: "node address for ssh.",
	}
	HostPortFlag = cli.IntFlag{
		Name:  "host.port",
		Usage: "node port for ssh.",
		Value: 22,
	}
	HostPasswordFlag = cli.StringFlag{
		Name:  "host.password",
		Usage: "host password for ssh.",
	}
	HostKeyPathFlag = cli.StringFlag{
		Name:  "host.keypath",
		Usage: "host key file path for ssh.",
	}
	HostDescriptionFlag = cli.StringFlag{
		Name:  "host.description",
		Usage: "description of host.",
	}
)

func NewApp() *cli.App {
	app := cli.NewApp()
	app.Name = "berithutils"
	app.Usage = "see berithutils --help"
	app.Author = "berith"
	app.Version = "0.0.1"
	return app
}

// GetDatabasePath returns a db directory
func GetDatabasePath() (string, error) {
	workspace, err := GetWorkspace()
	if err != nil {
		return "", err
	}
	return filepath.Join(workspace, "berithutilsdb"), nil
}

func GetWorkspace() (string, error) {
	cu, _ := user.Current()
	return filepath.Join(cu.HomeDir, "berithutils"), nil
}
