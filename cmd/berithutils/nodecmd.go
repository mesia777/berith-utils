package main

import (
	"encoding/json"
	"errors"
	"github.com/mesia777/berith-utils/node"
	"github.com/mesia777/berith-utils/types"
	"github.com/mesia777/berith-utils/utils"
	"github.com/urfave/cli"
	"io/ioutil"
	"log"
	"os"
)

var (
	nodeFlags = []cli.Flag{
		utils.NodeNameFlag,
		utils.HostUserFlag,
		utils.HostAddressFlag,
		utils.HostPortFlag,
		utils.HostPasswordFlag,
		utils.HostKeyPathFlag,
		utils.HostDescriptionFlag,
	}

	nodeCommand = cli.Command{
		Action:   ShowSubCommand,
		Name:     "node",
		Usage:    "manage nodes",
		Category: "NODE COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:   "import",
				Usage:  "Import nodes from given json file into local store",
				Action: importNodes,
				Flags: []cli.Flag{
					utils.PathFlag,
				},
			},
			{
				Name:   "add",
				Usage:  "Adds a node",
				Action: addNode,
				Flags:  nodeFlags,
			},
			{
				Name:   "get",
				Usage:  "Get a node",
				Action: displayNode,
				Flags:  nodeFlags,
			},
			{
				Name:   "gets",
				Usage:  "Get nodes",
				Action: displayNodes,
				Flags:  nodeFlags,
			},
			{
				Name:   "update",
				Usage:  "Update a node",
				Action: updateNode,
				Flags:  nodeFlags,
			},
			{
				Name:   "delete",
				Usage:  "Delete a node",
				Action: deleteNode,
				Flags:  nodeFlags,
			},
		},
	}
)

// importNodes import nodes given json path
func importNodes(ctx *cli.Context) error {
	path := ctx.String(utils.PathFlag.Name)
	if path == "" {
		return errors.New(`path must not not empty`)
	}

	jsonFile, err := os.Open(path)
	if err != nil {
		return err
	}
	defer jsonFile.Close()

	readBytes, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}

	var nodes []*types.Node
	err = json.Unmarshal(readBytes, &nodes)
	if err != nil {
		return err
	}

	var failures []string
	for _, n := range nodes {
		err = node.AddNode(app.db, n)

		if err != nil {
			failures = append(failures, n.Name)
			continue
		}
	}

	log.Printf("import hosts result >> try : %d / failures : %d. >>>> %v\n", len(nodes), len(failures), failures)
	return nil
}

// addNode save a node given cli context
func addNode(ctx *cli.Context) error {
	n, err := parseNode(ctx)
	if err != nil {
		return err
	}
	return node.AddNode(app.db, n)
}

// displayNode display a node given node name in cli context
func displayNode(ctx *cli.Context) error {
	n, err := parseNode(ctx)
	if err != nil {
		return err
	}
	if n.Name == "" {
		return errors.New(`can't find a node given node name ""`)
	}

	n, err = node.GetNode(app.db, n.Name)
	if err != nil {
		return err
	}
	displayNode0(n)
	return nil
}

// displayNodes display all nodes in local store
func displayNodes(ctx *cli.Context) error {
	nodes, err := node.GetNodes(app.db)
	if err != nil {
		return nil
	}
	displayNode0(nodes...)
	return nil
}

// displayNode0 show all hosts to console.
func displayNode0(nodes ...*types.Node) {
	if nodes == nil || len(nodes) == 0 {
		log.Printf("> empty nodes in local store")
		return
	}

	for i, n := range nodes {
		s, err := json.Marshal(n)
		if err != nil {
			log.Printf("%v -> %s\n", i+1, n.Name)
		} else {
			log.Printf("%v -> %s\n", i+1, string(s))
		}
	}
}

// updateNode update a node
func updateNode(ctx *cli.Context) error {
	n, err := parseNode(ctx)
	if err != nil {
		return err
	}
	return node.UpdateNode(app.db, n)
}

// deleteNode delete a node
func deleteNode(ctx *cli.Context) error {
	n, err := parseNode(ctx)
	if err != nil {
		return err
	}
	return node.DeleteHost(app.db, n.Name)
}

// parseNode extract node from cli context
func parseNode(ctx *cli.Context) (*types.Node, error) {
	host := &types.Host{
		User:        ctx.String(utils.HostUserFlag.Name),
		Address:     ctx.String(utils.HostAddressFlag.Name),
		Port:        ctx.Int(utils.HostPortFlag.Name),
		Password:    ctx.String(utils.HostPasswordFlag.Name),
		KeyPath:     ctx.String(utils.HostKeyPathFlag.Name),
		Description: ctx.String(utils.HostDescriptionFlag.Name),
	}
	return &types.Node{
		Name: ctx.String(utils.NodeNameFlag.Name),
		Host: host,
	}, nil
}
