package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
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
	fmt.Fprintln(session.textWindow, fmt.Sprintf("[%s::b]ERROR:[-::-]", colortoHexString(activeTheme.ErrorColor)), fmt.Sprint(v...))
	log.Println("[ERROR]", fmt.Sprint(v...))
}

func (session *viewerSession) logInfo(v ...interface{}) {
	fmt.Fprintln(session.textWindow, fmt.Sprintf("[%s::b]INFO:[-::-]", colortoHexString(activeTheme.InfoColor)), fmt.Sprint(v...))
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
		done := make(chan struct{})
		go func() {
			fn()
			done <- struct{}{}
		}()
		go session.blinkNode(node, done)
	}
}

func (session *viewerSession) blinkNode(node *tview.TreeNode, done chan struct{}) {
	originalText := node.GetText()
	originalColor := node.GetColor()
	node.SetText("loading...")
	for {
		select {
		case <-done:
			node.SetText(originalText)
			session.app.Draw()
			return
		default:
			node.SetColor(activeTheme.LoadingColor)
			session.app.Draw()
			time.Sleep(200 * time.Millisecond)
			node.SetColor(originalColor)
			session.app.Draw()
			time.Sleep(200 * time.Millisecond)
		}
	}

}

func hexStringToColor(hex string) tcell.Color {
	hex = strings.ReplaceAll(hex, "#", "")
	color, _ := strconv.ParseInt(hex, 16, 32)
	return tcell.NewHexColor(int32(color))
}

func colortoHexString(color tcell.Color) string {
	return fmt.Sprintf("#%06x", color.Hex())
}

func (t theme) apply() {
	if t.TerminalTextColor != "" {
		tview.Styles.PrimaryTextColor = hexStringToColor(t.TerminalTextColor)
	}

	if t.CategoryNodeColor != "" {
		activeTheme.CategoryNodeColor = hexStringToColor(t.CategoryNodeColor)
	} else {
		activeTheme.CategoryNodeColor = tcell.ColorOrange
	}
	if t.FolderNodeColor != "" {
		activeTheme.FolderNodeColor = hexStringToColor(t.FolderNodeColor)
	} else {
		activeTheme.FolderNodeColor = tcell.ColorWhite
	}
	if t.ItemNodeColor != "" {
		activeTheme.ItemNodeColor = hexStringToColor(t.ItemNodeColor)
	} else {
		activeTheme.ItemNodeColor = tcell.ColorLightGreen
	}
	if t.ActionNodeColor != "" {
		activeTheme.ActionNodeColor = hexStringToColor(t.ActionNodeColor)
	} else {
		activeTheme.ActionNodeColor = tcell.ColorDarkCyan
	}
	if t.BackgroundColor != "" {
		tview.Styles.PrimitiveBackgroundColor = hexStringToColor(t.BackgroundColor)
	}
	if t.BorderColor != "" {
		tview.Styles.BorderColor = hexStringToColor(t.BorderColor)
	}
	if t.NoContentColor != "" {
		activeTheme.NoContentColor = hexStringToColor(t.NoContentColor)
	} else {
		activeTheme.NoContentColor = tcell.ColorOrangeRed
	}
	if t.LoadingColor != "" {
		activeTheme.LoadingColor = hexStringToColor(t.LoadingColor)
	} else {
		activeTheme.LoadingColor = tcell.ColorDarkCyan
	}
	if t.LiveColor != "" {
		activeTheme.LiveColor = hexStringToColor(t.LiveColor)
	} else {
		activeTheme.LiveColor = tcell.ColorRed
	}
	if t.UpdateColor != "" {
		activeTheme.UpdateColor = hexStringToColor(t.UpdateColor)
	} else {
		activeTheme.UpdateColor = tcell.ColorDarkRed
	}
	if t.TerminalAccentColor != "" {
		activeTheme.TerminalAccentColor = hexStringToColor(t.TerminalAccentColor)
	} else {
		activeTheme.TerminalAccentColor = tcell.ColorGreen
	}
	if t.TerminalTextColor != "" {
		activeTheme.TerminalTextColor = hexStringToColor(t.TerminalTextColor)
	} else {
		activeTheme.TerminalTextColor = tview.Styles.PrimaryTextColor
	}
	if t.InfoColor != "" {
		activeTheme.InfoColor = hexStringToColor(t.InfoColor)
	} else {
		activeTheme.InfoColor = tcell.ColorGreen
	}
	if t.ErrorColor != "" {
		activeTheme.ErrorColor = hexStringToColor(t.ErrorColor)
	} else {
		activeTheme.ErrorColor = tcell.ColorRed
	}
}
