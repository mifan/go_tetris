package main

import "flag"

var (
	confPath = flag.String("conf", "./default.conf", "path to configuration file")
)

func initFlags() { flag.Parse() }
