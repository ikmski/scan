package main

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"sync"
	"time"

	"github.com/urfave/cli/v2"
)

var (
	version  string
	revision string
)

func scan(host string, port int, network string) bool {

	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout(network, address, time.Duration(1)*time.Second)
	if err != nil {
		return false
	}
	conn.Close()

	return true
}

func scanPorts(host string, startPort int, endPort int, udp bool) {

	network := "tcp"
	if udp {
		network = "udp"
	}

	var wg sync.WaitGroup
	for port := startPort; port <= endPort; port++ {
		wg.Add(1)
		go func(h string, p int, n string) {
			ok := scan(h, p, n)
			if ok {
				fmt.Printf("opening %d/%s port.\n", p, n)
			}
			wg.Done()
		}(host, port, network)
	}
	wg.Wait()
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
		return nil
	}

	var startPort int
	var endPort int
	if c.IsSet("port-range") {
		regex := regexp.MustCompile(`(\d*)\-(\d*)`)
		match := regex.FindStringSubmatch(c.String("port-range"))
		if match != nil && len(match) == 3 {
			if len(match[1]) > 0 {
				startPort, _ = strconv.Atoi(match[1])
			}
			if len(match[2]) > 0 {
				endPort, _ = strconv.Atoi(match[2])
			}
		}

		if startPort > 0 && endPort >= startPort {
			scanPorts(host, startPort, endPort, udp)
			return nil
		} else {
			return fmt.Errorf("Error: %s", "invalid format for port tange.")
		}
	}

	startPort = 1
	endPort = 1023
	scanPorts(host, startPort, endPort, udp)
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
