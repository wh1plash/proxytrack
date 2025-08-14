package main

import (
	"log"
	"os"
	"os/signal"
	"proxytrack/server"
	"syscall"

	"github.com/joho/godotenv"
)

const listenAddr = "0.0.0.0:8084"

func init() {
	mustLoadEnvVariables()
}

func main() {
	s := server.NewServer(listenAddr)

	go s.Run()

	sigch := make(chan os.Signal, 1)
	signal.Notify(sigch, os.Interrupt, syscall.SIGTERM)
	<-sigch
	log.Println("Received shutdown signal, shutting down server...")
	s.Stop()
}

func mustLoadEnvVariables() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

}
