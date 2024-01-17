package main

import (
	"fmt"
	"os"
	"pind/pkg"
	"pind/pkg/numa"
)

func main() {
	//_ = pkg.GetProcs0()
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
		}
	}

	printNuma()
}

func printHelp() {
	helpMsg := `Usage: pind [OPTIONS]
pin programs to CPU (affinity)

  -h, --help                 display this help and exit
  -v, --version              output version information and exit

      --print-numa          print information about numa and exit`

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

func exit0(err error) {
	if err == nil {
		os.Exit(0)
	}
	os.Exit(1)
}

func doService() {
	ctx := pkg.NewContext()
	pkg.RunService(ctx)
}
