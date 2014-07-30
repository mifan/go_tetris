package main

import "flag"

var (
	confPath = flag.String("conf", "./default.conf", "config file path")
)

func initFlags() {
	flag.Parse()
}
