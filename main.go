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
	flag.IntVar(&config.KeyLimit, "key-limit", 100, "key limit for set command")
	flag.StringVar(&config.AOFFILE, "aof", "./wumi-afile.aof", "name of aof file")

	flag.StringVar(&config.EvictionStrategy, "eviction", "allkeys-random", "explictly set key eviction strategies")
	flag.Float64Var(&config.EvictionRatio, "eviction-ratio", .4, "key eviction ratio")
	flag.Parse()
}

func main() {
	setupFlags()
	log.Println("server started")

	server.RunASyncTCPServer()
}
