package main

import (
	"fmt"
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
		}
	}

	printNuma()
}

func printHelp() {
	helpMsg := `Usage: pind [OPTIONS]
pin programs to CPU (affinity)

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

func doService() {
	ctx := pkg.NewContext()
	pkg.RunService(ctx)
}
