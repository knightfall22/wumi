package main

import (
	"flag"
	"log"

	"github.com/knightfall22/wumi/config"
	"github.com/knightfall22/wumi/server"
)

func setupFlags() {
	flag.StringVar(&config.Host, "host", "0.0.0.0", "host for server")
	flag.IntVar(&config.Port, "port", 7379, "port for server")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("server started")

	server.RunASyncTCPServer()
}
