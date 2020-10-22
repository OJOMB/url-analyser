package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/OJOMB/url-analyser/config"
	"github.com/OJOMB/url-analyser/server"
	"github.com/gorilla/mux"
)

var env = flag.String("env", "dev", "The environment in which the server is running. Options: ['dev', 'test', 'production']")

func main() {
	flag.Parse()

	fmt.Println(os.Executable())

	// get the logger
	logger := log.New(os.Stdout, "http: ", log.LstdFlags)
	logger.Printf("Server is starting...")

	// get the config
	cnfg := config.ConfigMap[*env]

	//get the router
	r := mux.NewRouter()

	// Instantiate server with shared dependencies
	s := server.New(r, logger, &cnfg)

	s.ListenAndServe()
}
