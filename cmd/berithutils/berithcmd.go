package main

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/mesia777/berith-utils/node"
	"github.com/mesia777/berith-utils/types"
	"github.com/mesia777/berith-utils/utils"
	"github.com/pkg/sftp"
	"github.com/urfave/cli"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"os"
	"path/filepath"
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
			{
				Name:      "upload",
				Usage:     "upload files in workspace/berith dir to node",
				Action:    uploadFiles,
				ArgsUsage: "[node name or empty if all]",
			},
			{
				Name:      "command",
				Usage:     "execute a command",
				Action:    executeCommand,
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
		return errors.New("empty nodes to build")
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
		return errors.New("empty nodes to start")
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
		return errors.New("empty nodes to stop")
	}

	executesCommand(nodes, func(n *types.Node) string {
		return STOP + " " + n.Name
	})
	return nil
}

// uploadConfigs upload files from {workspace}/berith to "~/berith-test/"
func uploadFiles(ctx *cli.Context) error {
	nodes, err := extractNodes(ctx)
	if err != nil {
		return err
	}

	if len(nodes) == 0 {
		return errors.New("empty nodes to upload files")
	}

	var success, fail []string
	upload := func(n *types.Node) (bool, string) {
		// setup sftp
		var out bytes.Buffer
		out.WriteString("------------------------------------------------\n")
		out.WriteString(fmt.Sprintf("try to upload files. node : %s\n", n.Name))

		c, err := createSSHClient(n)
		if err != nil {
			out.WriteString(fmt.Sprintf("failed to create a ssh client. node %s", n.Name))
			return false, out.String()
		}
		defer c.Close()

		client, err := sftp.NewClient(c)
		if err != nil {
			out.WriteString(fmt.Sprintf("failed to create a sftp client. node %s, %v", n.Name, err))
		}
		defer client.Close()

		workspace, err := utils.GetWorkspace()
		berithDir := filepath.Join(workspace, "berith")
		dir, err := ioutil.ReadDir(berithDir)
		if err != nil {
			out.WriteString(fmt.Sprintf("failed to read dir. node %s, %v", n.Name, err))
		}
		out.WriteString(fmt.Sprintf("Upload files(#%d) in %s\n", len(dir), berithDir))

		var uploadWait sync.WaitGroup
		uploadWait.Add(len(dir))

		// upload files
		for _, info := range dir {
			file, err := os.Open(filepath.Join(berithDir, info.Name()))
			if err != nil {
				out.WriteString(fmt.Sprintf("failed to open a file: %s", file.Name()))
				uploadWait.Done()
				continue
			}

			go func(file *os.File, info os.FileInfo) {
				defer uploadWait.Done()
				f, err := client.Create("berith-test/" + info.Name())
				if err != nil {
					fmt.Println("Failed to create a file:", info.Name())
					out.WriteString(fmt.Sprintf("failed to upload a file: %s,%v", file.Name(), err))
					return
				}
				defer f.Close()
				read, err := ioutil.ReadAll(file)
				if err != nil {
					fmt.Println("Failed to create a file:", info.Name())
					out.WriteString(fmt.Sprintf("failed to read a file: %s after create,%v", file.Name(), err))
					return
				}

				if _, err := f.Write(read); err != nil {
					fmt.Println("Failed to create a file:", info.Name())
					out.WriteString(fmt.Sprintf("failed to write content a file: %s,%v", file.Name(), err))
					return
				}
				err = client.Chmod(f.Name(), info.Mode())
				if err != nil {
					out.WriteString(fmt.Sprintf("failed to change permission%v", err))
				}
			}(file, info)
		}
		uploadWait.Wait()
		return true, out.String()
	}

	var waitGroup sync.WaitGroup
	waitGroup.Add(len(nodes))
	for _, n := range nodes {
		go func(n *types.Node) {
			defer waitGroup.Done()
			result, out := upload(n)
			if result {
				success = append(success, n.Name)
			} else {
				fail = append(fail, n.Name)
			}
			fmt.Println(out)
		}(n)
	}
	waitGroup.Wait()
	fmt.Printf("## Complete to upload. success nodes : %v / failures : %v\n", success, fail)
	return nil
}

func executeCommand(ctx *cli.Context) error {
	var nodeName, command string
	switch ctx.NArg() {
	case 1:
		command = ctx.Args()[0]
	case 2:
		nodeName = ctx.Args()[0]
		command = ctx.Args()[1]
	default:
		return errors.New("invalid args")
	}

	var nodes []*types.Node
	var err error

	if nodeName == "" {
		nodes, err = node.GetNodes(app.db)
		if err != nil {
			return err
		}
	} else {
		n, err := node.GetNode(app.db, ctx.Args()[0])
		if err != nil {
			return err
		}
		nodes = append(nodes, n)
	}
	if len(nodes) == 0 {
		return errors.New("empty nodes to execute command")
	}

	executesCommand(nodes, func(n *types.Node) string {
		return command
	})
	return nil
}

func executesCommand(nodes []*types.Node, cmdGen commandGenerator) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(nodes))

	var success []string
	var fail []string

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
				fail = append(fail, n.Name)
				return
			}
			defer conn.Close()

			session, err := conn.NewSession()
			if err != nil {
				b.WriteString("failed to execute. reason : cannot create a session\n")
				fail = append(fail, n.Name)
				return
			}
			defer session.Close()

			var stdOut bytes.Buffer
			var stdErr bytes.Buffer
			session.Stdout = &stdOut
			session.Stderr = &stdErr
			err = session.Run(cmd)
			if err != nil {
				b.WriteString(fmt.Sprintf("failed to execute a node %s. reason: %v\n", n.Name, err))
				b.Write(stdErr.Bytes())
				fail = append(fail, n.Name)
				return
			}
			b.WriteString("success to execute. node: " + n.Name + "\n")
			b.Write(stdOut.Bytes())
			b.WriteByte('\n')
			success = append(success, n.Name)
		}(n)
	}
	waitGroup.Wait()
	fmt.Printf("## Complete to execute nodes. success %v, fail : %v\n", success, fail)
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
