package main

import (
	"fmt"
	"github.com/pnsafonov/pind/pkg"
	"github.com/pnsafonov/pind/pkg/numa"
	log "github.com/sirupsen/logrus"
	"os"
	"strings"
)

var (
	version = dev
	commit  = "none"
	date    = "unknown"
	builtBy = "manual"
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
		case "--version-only":
			{
				printVersionOnly()
			}
		case "-h", "--help":
			{
				printHelp()
			}
		case "--print-numa":
			{
				printNuma()
			}
		case "--print-numa-phys":
			{
				printNumaPhys()
			}
		case "--print-topology":
			{
				printTopology()
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
		case "--print-name":
			{
				printName()
				continue
			}
		}

	}

	version0, gitHash := GetVersion0()
	ctx := pkg.NewContext(version0, gitHash, service0)
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
	printNumaPhys()
}

func printHelp() {
	helpMsg := `Usage: pind [OPTIONS]
pin programs to CPU (affinity)

  -s, --service, --daemon    run service (daemon)
  -c, --config               config file location, default is /etc/pind/pind.conf

  -h, --help                 display this help and exit
  -v, --version              output version information and exit
      --version-only         print numeric version information and exit

      --print-procs "pattern0,pattern1"       print process filtered by ps aux | grep pattern0 | grep pattern1 
      --print-numa                            print information about numa and exit
      --print-numa-phys                       print information about numa topology and exit
      --print-topology                        print information about cpu's topology
      --print-conf                            print config file content
      --print-name                            print pind and exit`

	fmt.Println(helpMsg)
	os.Exit(0)
}

func printVersion() {
	version0, gitHash := GetVersion0()
	fmt.Printf("pind version %s %s\n", version0, gitHash)
	os.Exit(0)
}

func printVersionOnly() {
	version0 := GetVersion()
	fmt.Println(version0)
	os.Exit(0)
}

func printNuma() {
	err := numa.PrintNuma0()
	exit0(err)
}

func printName() {
	fmt.Println("pind")
	os.Exit(0)
}

func printNumaPhys() {
	err := numa.PrintNumaPhys0()
	exit0(err)
}

func printTopology() {
	err := numa.PrintTopology()
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
