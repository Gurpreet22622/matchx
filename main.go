package main

import (
	"log"
	"matchx/config"
	"matchx/dbServer"
	"matchx/server"

	_ "github.com/lib/pq"
)

func main() {
	log.Println("Starting Runners App")
	log.Println("Initializing configuration")
	config := config.InitConfig("runner")
	log.Println("Initializing Database")
	dbHandler := dbServer.InitDatabase(config)
	log.Println("Initializing HTTP server")
	httpServer := server.InitHttpServer(config, dbHandler)
	httpServer.Start()
}
