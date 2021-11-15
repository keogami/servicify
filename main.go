package main

import (
	"flag"
	"os"
)

const (
	exitPathNotFound = iota + 1
	exitCantReadConfig
	exitInvalidConfigFile
	exitCantConnectSCM
	exitCantCreateService
	exitCantFindConfigAbsPath
	exitNotAService
	exitCantDetectContext

	exitCantFindImage
	exitCantStartImage
)

var (
	install = flag.String("install", "", "Install a service using given `config` file")
	remove  = flag.String("remove", "", "Remove the service by the given `name`")
	run     = flag.String("run", "", "Used internally for running the service")
)

func main() {
	flag.Parse()

	switch {
	case *install != "":
		Install(*install)
	case *remove != "":
		Remove(*remove)
	case *run != "":
		Run(*run)
	default:
		flag.PrintDefaults()
	}

	os.Exit(0)
}
