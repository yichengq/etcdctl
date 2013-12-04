package command

import (
	"errors"
	"fmt"
	"os"
	"os/signal"

	"github.com/codegangsta/cli"
	"github.com/coreos/go-etcd/etcd"
)

// NewWatchCommand returns the CLI command for "watch".
func NewWatchCommand() cli.Command {
	return cli.Command{
		Name:  "watch",
		Usage: "watch a key for changes",
		Flags: []cli.Flag{
			cli.BoolFlag{"forever", "forever watch a key until CTRL+C"},
			cli.IntFlag{"after-index", 0, "watch after the given index"},
			cli.BoolFlag{"recursive", "returns all values for key and child keys"},
		},
		Action: func(c *cli.Context) {
			handle(c, watchCommandFunc)
		},
	}
}

// watchCommandFunc executes the "watch" command.
func watchCommandFunc(c *cli.Context, client *etcd.Client) (*etcd.Response, error) {
	if len(c.Args()) == 0 {
		return nil, errors.New("Key required")
	}
	key := c.Args()[0]
	recursive := c.Bool("recursive")
	forever := c.Bool("forever")

	index := 0
	if c.Int("after-index") != 0 {
		index = c.Int("after-index") + 1
	}

	if forever {
		sigch := make(chan os.Signal, 1)
		signal.Notify(sigch, os.Interrupt)
		stop := make(chan bool)

		go func() {
			<-sigch
			stop <- true
			os.Exit(0)
		}()

		receiver := make(chan *etcd.Response)
		go client.Watch(key, uint64(index), recursive, receiver, stop)

		for {
			resp := <-receiver
			printResponse(resp, c.GlobalString("output"))
		}

	} else {
		var resp *etcd.Response
		var err error
		resp, err = client.Watch(key, uint64(index), recursive, nil, nil)

		if err != nil {
			fmt.Println("Error:", err)
			os.Exit(ErrorFromEtcd)
		}

		if err != nil {
			return nil, err
		}
		printResponse(resp, c.GlobalString("output"))
	}

	return nil, nil
}