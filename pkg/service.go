package pkg

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"pind/pkg/config"
	"syscall"
	"time"
)

type Context struct {
	ConfigPath string
	Service    bool

	Config *config.Config

	Done chan int
}

func NewContext() *Context {
	ctx := &Context{}
	ctx.Done = make(chan int, 1)
	ctx.ConfigPath = DefConfPath
	ctx.Service = false
	ctx.Config = config.NewDefaultConfig()
	return ctx
}

func RunService(ctx *Context) error {
	err := loadConfigAndInit(ctx)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "RunService, loadConfigAndInit err = %v", err)
		return err
	}

	doSignals(ctx)
	doLoop(ctx)

	return nil
}

func doSignals(ctx *Context) {
	done := ctx.Done

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGHUP, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	go signalsLoop(sigChan, done)
}

func signalsLoop(sigChan chan os.Signal, done chan int) {
	for {
		sig := <-sigChan
		log.Infof("received signal = %d", sig)
		switch sig {
		case syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT:
			done <- 1
			return
		case syscall.SIGHUP:
		default:
			return
		}
	}
}

func doLoop(ctx *Context) {
	interval := time.Duration(ctx.Config.Service.Interval)
	done := ctx.Done

	ticker := time.NewTicker(interval * time.Millisecond)
	for {
		select {
		case <-done:
			log.Infof("doLoop received done")
			return
		case t := <-ticker.C:
			log.Infof("Before sleep %v", t)
			time.Sleep(3 * time.Second)
			log.Infof("After Sleep  %v", t)
		}
	}
}

func handleTimer(ctx *Context, time0 time.Time) {

}
