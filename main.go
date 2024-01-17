package main

import "pind/pkg"

func main() {
	doMain()
}

func doMain() {
	//_ = pkg.GetProcs0()
	//pkg.DoTicker()
	doService()
}

func doService() {
	ctx := pkg.NewContext()
	pkg.RunService(ctx)
}
