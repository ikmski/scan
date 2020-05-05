package main

import (
	"fmt"
	"net"

	"github.com/urfave/cli/v2"
)

var (
	version  string
	revision string
)

func scan(host string, port int, network string) bool {

	address := fmt.Sprintf("%s:%d", host, port)
	_, err := net.Dial(network, address)
	if err != nil {
		return false
	}

	return true
}

func scanPorts() {

}

func scanSpecificPort(host string, port int, udp bool) {

	network := "tcp"
	if udp {
		network = "udp"
	}

	ok := scan(host, port, network)
	if ok {
		fmt.Printf("opening %d/%s port.\n", port, network)
	} else {
		fmt.Printf("%d/%s port is closed.\n", port, network)
	}
}

func mainAction(c *cli.Context) error {

	args := c.Args()
	if args.Len() == 0 {
		return fmt.Errorf("Error: %s", "host must be specified.")
	}

	host := args.First()
	udp := c.Bool("udp")

	if c.IsSet("port") {
		port := c.Int("port")
		scanSpecificPort(host, port, udp)
	}

	return nil
}

func main() {

	app := cli.NewApp()
	app.Name = "scan"
	app.Usage = "scan port"
	app.ArgsUsage = "host"
	app.Description = "command-line port scan tool"
	app.Version = version
	app.HideHelpCommand = true

	app.Flags = []cli.Flag{
		&cli.IntFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "port number",
		},
		&cli.StringFlag{
			Name:    "port-range",
			Aliases: []string{"r"},
			Usage:   "port range. ex) 1-1023",
		},
		&cli.BoolFlag{
			Name:    "udp",
			Aliases: []string{"u"},
			Usage:   "scan udp ports",
		},
	}

	app.Action = mainAction

	app.RunAndExitOnError()
}
