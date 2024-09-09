package main

import (
	"os"
	"os/signal"
	"context"
	"syscall"
	"time"
	"log"
)

func main() {
	fileName := "watcher.log"
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Panic(err)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	ctx, stop := signal.NotifyContext(context.Background(),
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
    cfg, err := LoadConfig(ctx, "./conf.yaml")
	if err != nil {
		panic(err)
	}
	newStatusChan := make(chan string)
	err = launchPoller(ctx, cfg.AccountID, cfg.VKToken, cfg.PollingInterval, newStatusChan)
	if err != nil {
		stop()
		time.Sleep(2 * time.Second)
		return
	}
	err = launchNotifier(ctx, cfg.ChatID, cfg.TGToken, newStatusChan)
	if err != nil {
		stop()
		time.Sleep(2 * time.Second)
		return
	}
	<-ctx.Done()
	log.Print("received exit code, gracefully shutting down...")
	time.Sleep(2 * time.Second)
	return
}