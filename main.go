package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path"
	"syscall"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/config"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/ui"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
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
		fmt.Printf("Could not open config: %v\n", err)
		os.Exit(1)
	}
	if openConfig {
		cfgPath, err := config.GetConfigPath()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = util.Open(path.Join(cfgPath, "config.toml"))
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}
	if openLogs {
		logPath, err := config.GetLogPath()
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		err = util.Open(logPath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		return
	}

	ui := ui.NewUI(cfg, version)
	go func() {
		if err := ui.Run(); err != nil {
			fmt.Println(err)
			os.Exit(1)
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
