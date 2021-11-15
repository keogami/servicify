package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/sys/windows/svc/mgr"
)

// Install registers a service using the given config filename
func Install(conf string) {
	exe, err := os.Executable()
	if err != nil {
		log.Println("Error: Couldn't find servicify's path")
		log.Println(err)
		os.Exit(exitPathNotFound)
	}

	confPath, err := filepath.Abs(conf)
	if err != nil {
		log.Println("Error: Couldn't find the absolute path to config file")
		log.Println(err)
		os.Exit(exitCantFindConfigAbsPath)
	}

	file, err := os.Open(confPath)
	if err != nil {
		log.Println("Error: Couldn't open config file")
		log.Println(err)
		os.Exit(exitCantReadConfig)
	}

	var c Config
	err = json.NewDecoder(file).Decode(&c)
	if err != nil {
		log.Println("Error: Couldn't decode config file")
		log.Println(err)
		os.Exit(exitCantReadConfig)
	}

	mc, err := c.Mold()
	if err != nil {
		log.Println("Error: Invalid config file")
		log.Println(err)
		os.Exit(exitInvalidConfigFile)
	}

	m, err := mgr.Connect()
	if err != nil {
		log.Println("Error: Couldn't connect to Service Control Manage")
		log.Println(err)
		os.Exit(exitCantConnectSCM)
	}
	defer m.Disconnect()

	s, err := m.CreateService(c.Name, exe, mc, "-run", confPath)
	if err != nil {
		log.Println("Error: Couldn't create service")
		log.Println(err)
		os.Exit(exitCantCreateService)
	}

	log.Printf("Success: Service{ %s } created\n", s.Name)
}
