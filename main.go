package main

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"pind/pkg"
	"pind/pkg/numa"
	"strings"
)

func main() {
	//_ = pkg.GetProcs0()
	//_ = pkg.PrintProcs0()
	//pkg.DoTicker()
	//doService()
	//numa.PrintNuma0()
	doMain(os.Args)
}

func doMain(args []string) {
	configPath := pkg.DefConfPath
	service0 := false

	l0 := len(args)
	for i := 1; i < l0; i++ {
		arg := args[i]
		switch arg {
		case "-v", "--version":
			{
				printVersion()
			}
		case "-h", "--help":
			{
				printHelp()
			}
		case "--print-numa":
			{
				printNuma()
			}
		case "--print-procs":
			{
				i0 := i + 1
				if i0 < l0 {
					patterns := args[i0]
					printProcs1(patterns)
				} else {
					exitMoreArgs()
				}
				i++
				continue
			}
		case "-c", "--config":
			{
				i0 := i + 1
				if i0 < l0 {
					configPath = args[i0]
				}
				i++
				continue
			}
		case "-s", "--service", "--daemon":
			{
				service0 = true
				continue
			}
		}

	}

	ctx := pkg.NewContext()
	ctx.Service = service0
	ctx.ConfigPath = configPath

	if service0 {
		runService(ctx)
	}

	// default action
	printNuma()
}

func printHelp() {
	helpMsg := `Usage: pind [OPTIONS]
pin programs to CPU (affinity)

  -s, --service, --daemon    run service (daemon)
  -c, --config               config file location, default is /etc/pind/pind.conf

  -h, --help                 display this help and exit
  -v, --version              output version information and exit

      --print-procs "pattern0,pattern1"       print process filtered by ps aux | grep pattern0 | grep pattern1 
      --print-numa                            print information about numa and exit`

	fmt.Println(helpMsg)
	os.Exit(0)
}

func printVersion() {
	versionMsg := `pind 1.0.0`
	fmt.Println(versionMsg)
	os.Exit(0)
}

func printNuma() {
	err := numa.PrintNuma0()
	exit0(err)
}

func printProcs1(pattern string) {
	patterns := strings.Split(pattern, ",")
	err := pkg.PrintProcs1(patterns)
	exit1(err)
}

func exit0(err error) {
	if err == nil {
		os.Exit(0)
	}
	os.Exit(1)
}

func exitMoreArgs() {
	_, _ = fmt.Fprintf(os.Stderr, "Please specify more args.")
	os.Exit(1)
}

func exit1(err error) {
	if err == nil {
		os.Exit(0)
	}
	_, _ = fmt.Fprintf(os.Stderr, "err = %v\n", err)
	os.Exit(1)
}

func exit2(err error) {
	if err == nil {
		os.Exit(0)
	}
	log.Errorf("exit with err = %v\n", err)
	os.Exit(1)
}

func runService(ctx *pkg.Context) {
	err := pkg.RunService(ctx)
	exit2(err)
}
