package main

import (
	"flag"
	"fmt"
	"github.com/fogcreek/mini"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

var (
	router *mux.Router         = mux.NewRouter()
	routes map[string]*handler = map[string]*handler{
		"/blocklists": BlocklistsHandler,
        "/blocks": BlocksHandler,
	}

	// Config vars
	blocklistBuckets []string
	blockBuckets     []string
	aws_key          string
	aws_sec          string
	pool_size        int = 1
	host             string
	port             int
	num_expeditions  int = 1
	regions []string

	pool Pool
)

func init() {

	for r, h := range routes {
		router.HandleFunc(r, h.Function).Methods(h.Methods...)
	}
}

func handle_config(cfg *string) {
	config, err := mini.LoadConfiguration(*cfg)
	if err != nil {
		log.Panicf("Failed to load config file %s %s", *config, err.Error())
	}

	blocklistBuckets = config.StringsFromSection("blocklist", "buckets")
	blockBuckets = config.StringsFromSection("block", "buckets")

	aws_key = config.StringFromSection("certs", "key", "")
	aws_sec = config.StringFromSection("certs", "secret", "")

	pool_size = int(config.IntegerFromSection("general", "pool-size", 2))

	host = config.StringFromSection("general", "host", "")
	port = int(config.IntegerFromSection("general", "port", 8598))

	num_expeditions = int(config.IntegerFromSection("general", "expedition-parties", 2))

	regions = config.StringsFromSection("region-mappings", "regions")

	if len(blocklistBuckets) == 0 {
		log.Panicf("No buckets found config file for blocklists, must have at least one\n")
	}

	if len(blockBuckets) == 0 {
		log.Panicf("No buckets found config file for blocks, must have at least one\n")
	}

	if aws_key == "" || aws_sec == "" {
		log.Panicf("Invalid certs, one or more of your keys was empty\n")
	}
}

func main() {

	_config := flag.String("c", "", "path to config file to use")
	flag.Parse()
	handle_config(_config)

	FillPool(pool_size)

	http.ListenAndServe(fmt.Sprintf("%s:%d", host, port), router)
}
