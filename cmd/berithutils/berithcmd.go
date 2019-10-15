package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mesia777/berith-utils/node"
	"github.com/mesia777/berith-utils/types"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"strconv"
	"sync"
)

const (
	WORKSPACE = "~/berith-test"
	INIT      = WORKSPACE + "/init.sh"
	BUILD     = WORKSPACE + "/build.sh"
	START     = WORKSPACE + "/start.sh"
	STOP      = WORKSPACE + "/stop.sh"
)

var (
	berithCommand = cli.Command{
		Action:   ShowSubCommand,
		Name:     "berith",
		Usage:    "[subcommands]",
		Category: "BERITH COMMANDS",
		Subcommands: []cli.Command{
			{
				Name:      "init",
				Usage:     "init nodes",
				Action:    initNodes,
				ArgsUsage: "[node name or empty if all]",
			},
			{
				Name:      "build",
				Usage:     "build nodes",
				Action:    buildNodes,
				ArgsUsage: "[node name or empty if all]",
			},
			{
				Name:      "start",
				Usage:     "start nodes",
				Action:    startNodes,
				ArgsUsage: "[node name or empty if all]",
			},
			{
				Name:      "stop",
				Usage:     "stop nodes",
				Action:    stopNodes,
				ArgsUsage: "[node name or empty if all]",
			},
		},
	}
)

type commandGenerator func(n *types.Node) string

// initNodes initialize berith node given cli context
func initNodes(ctx *cli.Context) error {
	nodes, err := extractNodes(ctx)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("empty nodes to init")
	}

	executesCommand(nodes, func(n *types.Node) string {
		return INIT + " " + n.Name
	})
	return nil
}

// buildNodes build berith nodes given cli context
func buildNodes(ctx *cli.Context) error {
	nodes, err := extractNodes(ctx)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("empty nodes to init")
	}

	executesCommand(nodes, func(n *types.Node) string {
		return BUILD + " " + n.Name
	})
	return nil
}

// startNodes start berith nodes given cli context
func startNodes(ctx *cli.Context) error {
	nodes, err := extractNodes(ctx)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("empty nodes to init")
	}

	executesCommand(nodes, func(n *types.Node) string {
		return START + " " + n.Name
	})
	return nil
}

// stopNodes stop berith nodes given cli context
func stopNodes(ctx *cli.Context) error {
	nodes, err := extractNodes(ctx)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("empty nodes to init")
	}

	executesCommand(nodes, func(n *types.Node) string {
		return STOP + " " + n.Name
	})
	return nil
}

func executesCommand(nodes []*types.Node, cmdGen commandGenerator) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(nodes))

	for _, n := range nodes {
		go func(n *types.Node) {
			defer waitGroup.Done()
			var b bytes.Buffer
			defer func() {
				fmt.Println(b.String())
			}()
			cmd := cmdGen(n)
			b.WriteString("------------------------------------------------\n")
			b.WriteString(fmt.Sprintf("try to execute a command. node : %s, command : %s\n", n.Name, cmd))

			conn, err := createSSHClient(n)
			if err != nil {
				b.WriteString("failed to execute. reason : cannot create a ssh client\n")
				return
			}
			defer conn.Close()

			session, err := conn.NewSession()
			if err != nil {
				b.WriteString("failed to execute. reason : cannot create a session\n")
				return
			}
			defer session.Close()

			var stdOut bytes.Buffer
			session.Stdout = &stdOut
			err = session.Run(cmd)
			if err != nil {
				b.WriteString(fmt.Sprintf("failed to execute a node %s. reason: %v\n", n.Name, err))
				return
			}
			b.WriteString("success to execute. node: " + n.Name + "\n")
			b.Write(stdOut.Bytes())
			b.WriteByte('\n')
		}(n)
	}

	waitGroup.Wait()
}

// extractNodes extract nodes given cli context
func extractNodes(ctx *cli.Context) ([]*types.Node, error) {
	if ctx.NArg() < 1 {
		return node.GetNodes(app.db)
	}

	var nodes []*types.Node
	n, err := node.GetNode(app.db, ctx.Args()[0])
	if err != nil {
		return nil, err
	}
	nodes = append(nodes, n)
	return nodes, nil
}

// createSSHClient create a ssh client given node
func createSSHClient(n *types.Node) (*ssh.Client, error) {
	h := n.Host
	if h == nil {
		return nil, errors.New("cannot create a ssh client. host is nil")
	}

	var auth ssh.AuthMethod
	if h.Password != "" {
		auth = ssh.Password(h.Password)
	} else {
		pemBytes, err := ioutil.ReadFile(h.KeyPath)
		if err != nil {
			return nil, err
		}
		key, err := ssh.ParsePrivateKey(pemBytes)
		auth = ssh.PublicKeys(key)
	}

	config := &ssh.ClientConfig{
		User: h.User,
		Auth: []ssh.AuthMethod{
			auth,
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
	}

	addr := h.Address + ":" + strconv.Itoa(h.Port)
	return ssh.Dial("tcp", addr, config)
}
