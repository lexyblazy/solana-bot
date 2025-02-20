package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"solana-bot/config"
	"solana-bot/engine"
	"syscall"
)

func getConfig() *config.Config {

	var config config.Config

	file, err := os.Open("./config.json")

	if err != nil {
		log.Fatal("Failed to open config file", err)
	}

	json.NewDecoder(file).Decode(&config)

	return &config

}

func main() {

	exitChan := make(chan os.Signal, 1)
	signal.Notify(exitChan, syscall.SIGINT, syscall.SIGTERM)

	config := getConfig()

	e := engine.New(config)

	go e.Start()

	signal := <-exitChan

	e.Cleanup()

	log.Printf("Received %s signal... graceful shutdown \n", signal)

	os.Exit(0)

}
