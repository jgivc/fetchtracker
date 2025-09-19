package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jgivc/fetchtracker/internal/app"
)

func main() {
	cfgFileName := flag.String("c", "config.yml", "Path to config file")
	flag.Parse()

	app := app.New(*cfgFileName)
	go app.Start()

	c := make(chan os.Signal, 1)
	defer close(c)
	done := make(chan struct{})

	signal.Notify(c, os.Interrupt, syscall.SIGINT, syscall.SIGTERM, syscall.SIGUSR1, syscall.SIGUSR2)
	go func() {
		defer close(done)

		for sig := range c {
			switch sig {
			case syscall.SIGUSR1:
				go app.Index()
			case syscall.SIGUSR2:
				go app.Dump()
			case syscall.SIGTERM, syscall.SIGINT:
				fmt.Println("Received termination signal. Shutting down...")
				done <- struct{}{}

				return
			}
		}
	}()

	<-done
	app.Stop()
	time.Sleep(2 * time.Second)
	fmt.Println("done")
}
