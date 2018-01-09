package main

import (
	"net"
	"os"
	"os/signal"

	"github.com/ghetzel/cli"
	"github.com/ghetzel/go-stockutil/stringutil"
	"github.com/ghetzel/mdns"
	"github.com/op/go-logging"
)

var log = logging.MustGetLogger(`main`)

func main() {
	app := cli.NewApp()
	app.Name = `mdns`
	app.Usage = `Simple mDNS/DNS-SD client/server`
	app.Version = `0.0.1`

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:   `log-level, L`,
			Usage:  `Level of log output verbosity`,
			Value:  `debug`,
			EnvVar: `LOGLEVEL`,
		},
	}

	app.Before = func(c *cli.Context) error {
		logging.SetFormatter(logging.MustStringFormatter(`%{color}%{level:.4s}%{color:reset}[%{id:04d}] %{message}`))

		if level, err := logging.LogLevel(c.String(`log-level`)); err == nil {
			logging.SetLevel(level, ``)
		} else {
			return err
		}

		log.Infof("Starting %s %s", c.App.Name, c.App.Version)
		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:      `publish`,
			Usage:     `Publish a mDNS service.`,
			ArgsUsage: `NAME SERVICE_TYPE PORT [TXT ..]`,
			Flags:     []cli.Flag{},
			Action: func(c *cli.Context) {
				instanceName := c.Args().Get(0)
				serviceType := c.Args().Get(1)
				portS := c.Args().Get(2)
				port := 0

				if instanceName == `` {
					log.Fatalf("must specify NAME")
				}

				if serviceType == `` {
					log.Fatalf("must specify SERVICE_TYPE")
				}

				if portS == `` {
					log.Fatalf("must specify PORT")
				} else if v, err := stringutil.ConvertToInteger(portS); err == nil {
					if v > 0 && v < 65536 {
						port = int(v)
					} else {
						log.Fatalf("port not in range [0, 65535]")
					}
				} else {
					log.Fatalf("invalid port: %v", err)
				}

				txt := make([]string, 0)

				if len(c.Args()) > 3 {
					txt = c.Args()[3:]
				}

				if service, err := mdns.NewMDNSService(
					instanceName,
					serviceType,
					``,
					``,
					port,
					[]net.IP{
						net.ParseIP(`10.200.55.215`),
					},
					txt,
				); err == nil {
					service := &mdns.DNSSDService{
						MDNSService: service,
					}

					// Create the mDNS server, defer shutdown
					if server, err := mdns.NewServer(&mdns.Config{
						Zone: service,
					}); err == nil {
						defer server.Shutdown()

						ch := make(chan os.Signal)
						signal.Notify(ch, os.Interrupt, os.Kill)
						<-ch
					} else {
						log.Fatalf("publish failed: %v", err)
					}
				} else {
					log.Fatalf("invalid service: %v", err)
				}
			},
		},
	}

	app.Run(os.Args)
}
