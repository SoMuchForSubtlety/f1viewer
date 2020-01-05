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

type theme struct {
	PrimitiveBackgroundColor    string `json:"primitive_background_color"`     // Main background color for primitives.
	ContrastBackgroundColor     string `json:"contrast_background_color"`      // Background color for contrasting elements.
	MoreContrastBackgroundColor string `json:"more_contrast_background_color"` // Background color for even more contrasting elements.
	BorderColor                 string `json:"border_color"`                   // Box borders.
	TitleColor                  string `json:"title_color"`                    // Box titles.
	GraphicsColor               string `json:"graphics_color"`                 // Graphics.
	PrimaryTextColor            string `json:"primary_text_color"`             // Primary text.
	SecondaryTextColor          string `json:"secondary_text_color"`           // Secondary text (e.g. labels).
	TertiaryTextColor           string `json:"tertiary_text_color"`            // Tertiary text (e.g. subtitles, notes).
	InverseTextColor            string `json:"inverse_text_color"`             // Text on primary-colored backgrounds.
	ContrastSecondaryTextColor  string `json:"contrast_secondary_text_color"`  // Secondary text on ContrastBackgroundColor-colored backgrounds.
}

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

func hexStringToColor(hex string) tcell.Color {
	hex = strings.ReplaceAll(hex, "#", "")
	color, _ := strconv.ParseInt(hex, 16, 32)
	return tcell.NewHexColor(int32(color))
}

func (t theme) apply() {
	if t.PrimitiveBackgroundColor != "" {
		tview.Styles.PrimitiveBackgroundColor = hexStringToColor(t.PrimitiveBackgroundColor)
	}
	if t.ContrastBackgroundColor != "" {
		tview.Styles.ContrastBackgroundColor = hexStringToColor(t.ContrastBackgroundColor)
	}
	if t.MoreContrastBackgroundColor != "" {
		tview.Styles.MoreContrastBackgroundColor = hexStringToColor(t.MoreContrastBackgroundColor)
	}
	if t.BorderColor != "" {
		tview.Styles.BorderColor = hexStringToColor(t.BorderColor)
	}
	if t.TitleColor != "" {
		tview.Styles.TitleColor = hexStringToColor(t.TitleColor)
	}
	if t.GraphicsColor != "" {
		tview.Styles.GraphicsColor = hexStringToColor(t.GraphicsColor)
	}
	if t.PrimaryTextColor != "" {
		tview.Styles.PrimaryTextColor = hexStringToColor(t.PrimaryTextColor)
	}
	if t.SecondaryTextColor != "" {
		tview.Styles.SecondaryTextColor = hexStringToColor(t.SecondaryTextColor)
	}
	if t.TertiaryTextColor != "" {
		tview.Styles.TertiaryTextColor = hexStringToColor(t.TertiaryTextColor)
	}
	if t.InverseTextColor != "" {
		tview.Styles.InverseTextColor = hexStringToColor(t.InverseTextColor)
	}
	if t.ContrastSecondaryTextColor != "" {
		tview.Styles.ContrastSecondaryTextColor = hexStringToColor(t.ContrastSecondaryTextColor)
	}
}
