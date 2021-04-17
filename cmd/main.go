package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/SoMuchForSubtlety/f1viewer/internal/config"
	"github.com/SoMuchForSubtlety/f1viewer/internal/ui"
	"github.com/skratchdot/open-golang/open"
)

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	var showVersion bool
	var openConfig bool
	var openLogs bool
	flag.BoolVar(&showVersion, "v", showVersion, "show version information")
	flag.BoolVar(&showVersion, "version", showVersion, "show version information")
	flag.BoolVar(&openConfig, "config", openConfig, "open config file")
	flag.BoolVar(&openLogs, "logs", openLogs, "open logs directory")
	flag.Parse()
	if showVersion {
		fmt.Println(buildVersion())
		return
	}
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Could not open config: %v", err)
	}
	if openConfig {
		cfgPath, err := config.GetConfigPath()
		if err != nil {
			log.Fatal(err)
		}
		err = open.Start(path.Join(cfgPath, "config.json"))
		if err != nil {
			log.Fatal(err)
		}
		return
	}
	if openLogs {
		logPath, err := config.GetLogPath(cfg)
		if err != nil {
			log.Fatal(err)
		}
		err = open.Start(logPath)
		if err != nil {
			log.Fatal(err)
		}
		return
	}

	ui := ui.NewUI(cfg, version)
	go func() {
		if err := ui.Run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c

	ui.Stop()
}

func buildVersion() string {
	result := fmt.Sprintf("Version:     %s", version)
	if commit != "" {
		result += fmt.Sprintf("\nGit commit:  %s", commit)
	}
	if date != "" {
		result += fmt.Sprintf("\nBuilt:       %s", date)
	}
	return result
}
