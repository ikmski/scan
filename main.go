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

type portStatus int

const (
	closed portStatus = iota
	open
)

type protocol int

const (
	tcp protocol = iota
	udp
)

type scanRequest struct {
	host  string
	port  int
	proto protocol
}

type scanResult struct {
	host   string
	port   int
	proto  protocol
	status portStatus
}

func (p protocol) toString() string {
	if p == tcp {
		return "tcp"
	}
	return "udp"
}

func scanTCP(host string, port int) scanResult {

	result := scanResult{
		host:   host,
		port:   port,
		proto:  tcp,
		status: closed,
	}
	address := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", address, time.Duration(1)*time.Second)
	if err != nil {
		return result
	}
	defer conn.Close()

	result.status = open
	return result
}

func scanUDP(host string, port int) scanResult {

	result := scanResult{
		host:   host,
		port:   port,
		proto:  udp,
		status: closed,
	}
	address := &net.UDPAddr{
		IP:   net.ParseIP(host),
		Port: port,
	}

	conn, err := net.DialUDP("udp", nil, address)
	if err != nil {
		return result
	}
	defer conn.Close()

	count := 0
	for i := 0; i < 10; i++ {
		_, err = conn.Write([]byte{0xFF})
		if err != nil {
			count++
		}
	}
	if count > 0 {
		return result
	}

	result.status = open
	return result
}

func scan(req <-chan scanRequest, res chan<- scanResult, wg *sync.WaitGroup) {

	for r := range req {

		if r.proto == tcp {
			res <- scanTCP(r.host, r.port)
		} else {
			res <- scanUDP(r.host, r.port)
		}
		wg.Done()
	}
}

func listen(res <-chan scanResult) {

	for r := range res {
		if r.status == open {
			fmt.Printf("opening %d/%s port.\n", r.port, r.proto.toString())
		}
	}
}

func scanPorts(host string, startPort int, endPort int, isUDP bool) {

	req := make(chan scanRequest)
	res := make(chan scanResult)

	var wg sync.WaitGroup

	go listen(res)

	// worker
	worker := 100
	if endPort-startPort+1 < worker {
		worker = endPort - startPort + 1
	}

	for i := 0; i < worker; i++ {
		go scan(req, res, &wg)
	}

	for port := startPort; port <= endPort; port++ {

		request := scanRequest{
			host:  host,
			port:  port,
			proto: tcp,
		}
		if isUDP {
			request.proto = udp
		}

		wg.Add(1)
		go func(r scanRequest) {
			req <- r
		}(request)
	}

	wg.Wait()

	close(req)
	time.Sleep(1 * time.Second)
	close(res)
	time.Sleep(1 * time.Second)
}

func mainAction(c *cli.Context) error {

	args := c.Args()
	if args.Len() == 0 {
		return fmt.Errorf("Error: %s", "host must be specified.")
	}

	host := args.First()
	isUDP := c.Bool("udp")

	var startPort int
	var endPort int
	if c.IsSet("port") {
		startPort, endPort, err := parsePort(c.String("port"))
		if err != nil {
			return err
		}
		scanPorts(host, startPort, endPort, isUDP)
		return nil
	}

	startPort = 1
	endPort = 1023
	scanPorts(host, startPort, endPort, isUDP)
	return nil
}

func parsePort(str string) (start int, end int, err error) {

	regex := regexp.MustCompile(`(\d*)\-(\d*)`)
	match := regex.FindStringSubmatch(str)
	if match != nil && len(match) == 3 {
		if len(match[1]) > 0 {
			start, _ = strconv.Atoi(match[1])
		}
		if len(match[2]) > 0 {
			end, _ = strconv.Atoi(match[2])
		}

		if start > 0 && end >= start {
			return
		}
	}

	regex = regexp.MustCompile(`(\d*)`)
	match = regex.FindStringSubmatch(str)
	if match != nil && len(match) == 2 {
		if len(match[1]) > 0 {
			start, _ = strconv.Atoi(match[1])
			end = start
		}

		if start > 0 {
			return
		}
	}

	return 0, 0, fmt.Errorf("Error: %s", "invalid format for port tange.")
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
		&cli.StringFlag{
			Name:    "port",
			Aliases: []string{"p"},
			Usage:   "port number. ex) 22, 1-1023",
		},
		&cli.BoolFlag{
			Name:  "udp",
			Usage: "scan udp ports",
		},
	}

	app.Action = mainAction

	app.RunAndExitOnError()
}
