package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"golang.org/x/sys/windows/svc"
)

// Service is the data required by the service to run
type Service struct {
	c      Config
	path   string
	status svc.Status
}

type nothing struct{}

// Execute implements the svc.Handler interface
// it loads the image from the path specified by the config and runs it as its child process
// if the service is stopped while the image is executing, it sends a sigkill to the image
// once the image returns, regardless of whether the service was stopped, the service itself finishes execution
// and returns the exit code of the image as a service specific exit code.
func (service Service) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (svcSpecificEC bool, exitCode uint32) {
	imagePath := service.c.Image
	if !filepath.IsAbs(imagePath) {
		imagePath = filepath.Join(service.path, imagePath)
	}
	ctx, cancel := context.WithCancel(context.Background())

	cmd := exec.CommandContext(ctx, imagePath, service.c.Options...)
	cmd.Dir = service.path

	if err := cmd.Start(); err != nil {
		cancel()
		return true, exitCantStartImage
	}

	service.status.Accepts = svc.AcceptStop
	service.status.State = svc.Running
	s <- service.status

	stop := pipeEvents(r)

	go func() {
		<-stop
		cancel()
	}()

	ec := cmd.Wait().(*exec.ExitError).ExitCode()

	service.status.State = svc.Stopped
	s <- service.status

	return true, uint32(ec)
}

func pipeEvents(req <-chan svc.ChangeRequest) <-chan nothing {
	stop := make(chan nothing)

	go func(r <-chan svc.ChangeRequest) {
		for cr := range r {
			switch cr.Cmd {
			case svc.Shutdown:
				fallthrough
			case svc.Stop:
				stop <- nothing{}
			}
		}
	}(req)

	return stop
}

// Run runs the given service
func Run(configPath string) {
	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Println("Error: Can't detect current running context")
		log.Println(err)
		os.Exit(exitCantDetectContext)
	}

	if !isService {
		log.Println("Error: Not running as a service")
		log.Println("Hint: Use -install if you want to install a service")
		os.Exit(exitNotAService)
	}

	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		log.Println("Error: Couldn't read config file")
		log.Println(err)
		os.Exit(exitCantReadConfig)
	}

	var c Config
	err = json.Unmarshal(data, &c)
	if err != nil {
		log.Println("Error: Couldn't decode config file")
		log.Println(err)
		os.Exit(exitCantReadConfig)
	}

	service := Service{
		c:    c,
		path: filepath.Dir(configPath),
		status: svc.Status{
			ProcessId: uint32(os.Getpid()),
		},
	}

	svc.Run(c.Name, service)
}
