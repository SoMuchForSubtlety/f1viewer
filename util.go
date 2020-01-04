package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

// takes year/race ID and returns full year and race nuber as strings
func getYearAndRace(input string) (string, string, error) {
	var fullYear string
	var raceNumber string
	if len(input) < 4 {
		return fullYear, raceNumber, errors.New("not long enough")
	}
	_, err := strconv.Atoi(input[:4])
	if err != nil {
		return fullYear, raceNumber, errors.New("not a valid RearRaceID")
	}
	// TODO fix before 2020
	if input[:4] == "2018" || input[:4] == "2019" {
		return input[:4], "0", nil
	}
	year := input[:2]
	intYear, _ := strconv.Atoi(year)
	// TODO: change before 2030
	if intYear < 30 {
		fullYear = "20" + year
	} else {
		fullYear = "19" + year
	}
	raceNumber = input[2:4]
	return fullYear, raceNumber, nil
}

func (session *viewerSession) logError(v ...interface{}) {
	fmt.Fprintln(session.textWindow, "[red::b]ERROR:[-::-]", fmt.Sprint(v...))
	log.Println("[ERROR]", fmt.Sprint(v...))
}

func (session *viewerSession) logInfo(v ...interface{}) {
	fmt.Fprintln(session.textWindow, "[green::b]INFO:[-::-]", fmt.Sprint(v...))
	log.Println("[INFO]", fmt.Sprint(v...))
}

func configureLogging(cfg config) (*os.File, error) {
	if cfg.SaveLogs {
		logPath, err := getLogPath(cfg)
		if err != nil {
			return nil, fmt.Errorf("Could not get log path: %w", err)
		}
		completePath := fmt.Sprint(logPath, time.Now().Format("2006-01-02"), ".log")
		logFile, err := os.OpenFile(completePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
			return nil, fmt.Errorf("Could not open log file: %w", err)
		}
		log.SetOutput(logFile)
		return logFile, nil
	}
	return nil, nil
}

func (session *viewerSession) withBlink(node *tview.TreeNode, fn func()) func() {
	return func() {
		done := false
		go func() {
			fn()
			done = true
		}()
		go session.blinkNode(node, &done)
	}
}

func (session *viewerSession) blinkNode(node *tview.TreeNode, done *bool) {
	originalText := node.GetText()
	originalColor := node.GetColor()
	node.SetText("loading...")
	for !*done {
		node.SetColor(tcell.ColorBlue)
		session.app.Draw()
		time.Sleep(200 * time.Millisecond)
		node.SetColor(originalColor)
		session.app.Draw()
		time.Sleep(200 * time.Millisecond)

	}
	node.SetText(originalText)
	session.app.Draw()
}
