package main

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg"
	"github.com/pnsafonov/pind/pkg/numa"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	doMain(os.Args)
}

func doMain(args []string) {
	configPath := pkg.ConfPathBuiltIn
	service0 := false
	printConf0 := false

	pkg.InitConsoleLogger()

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
				patterns := ""

				i0 := i + 1
				if i0 < l0 {
					patterns = args[i0]
				}

				printProcs1(patterns)
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
		case "--print-conf":
			{
				printConf0 = true
				continue
			}
		}

	}

	ctx := pkg.NewContext()
	ctx.Service = service0
	ctx.ConfigPath = configPath
	ctx.PrintConfig = printConf0

	if printConf0 {
		printConf(ctx)
	}
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
      --print-numa                            print information about numa and exit
      --print-conf                            print config file content`

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
	var err error

	if pattern != "" {
		patterns := strings.Split(pattern, ",")
		err = pkg.PrintProcs1(patterns)
	} else {
		err = pkg.PrintProcs2()
	}

	exit1(err)
}

func exit0(err error) {
	if err == nil {
		os.Exit(0)
	}
	os.Exit(1)
}

func exitMoreArgs() {
	log.Errorf("Please specify more args.")
	os.Exit(1)
}

func exit1(err error) {
	if err == nil {
		os.Exit(0)
	}
	log.Errorf("err = %v", err)
	os.Exit(1)
}

func runService(ctx *pkg.Context) {
	err := pkg.RunService(ctx)
	exit1(err)
}

func printConf(ctx *pkg.Context) {
	err := pkg.PrintConf0(ctx)
	exit1(err)
}
