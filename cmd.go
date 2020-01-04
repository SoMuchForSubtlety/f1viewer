package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/user"
	"regexp"
	"strings"
)

type command struct {
	Title   string `json:"title"`
	Command string `json:"command"`
}

type commandContext struct {
	EpID          string
	CustomOptions command
	Title         string
}

func (session *viewerSession) runCustomCommand(cc commandContext) error {
	// custom command
	com := cc.CustomOptions
	url, err := getPlayableURL(cc.EpID)
	if err != nil {
		return err
	}
	tmpCommand := com.Command
	// replace $url, $file and $cookie
	var filepath string
	if strings.Contains(tmpCommand, "$file") && filepath == "" {
		filepath, _, err = session.con.downloadAsset(url, cc.Title)
		if err != nil {
			return err
		}
	}
	tmpCommand = strings.Replace(tmpCommand, "$file", filepath, -1)
	tmpCommand = strings.Replace(tmpCommand, "$url", url, -1)
	splitCommand := strings.Split(tmpCommand, " ")
	return session.runCmd(exec.Command(splitCommand[0], splitCommand[1:]...))
}

func (session *viewerSession) runCmd(cmd *exec.Cmd) error {
	wdir, _ := os.Getwd()
	user, _ := user.Current()
	hostname, _ := os.Hostname()
	if wdir == user.HomeDir {
		wdir = "~"
	} else {
		re := regexp.MustCompile("[^\\/]+$")
		wdir = re.FindString(wdir)
	}
	fmt.Fprintln(session.textWindow, fmt.Sprintf("\n[green::b][%s@%s [white]%s[green]]$[-::-] %s", user.Username, hostname, wdir, strings.Join(cmd.Args, " ")))

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}
	go io.Copy(session.textWindow, stdout)
	go io.Copy(session.textWindow, stderr)
	err = cmd.Start()
	if err != nil {
		return err
	}
	return cmd.Process.Release()
}
