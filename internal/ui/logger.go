package ui

import (
	"fmt"
	"log"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
	"github.com/rivo/tview"
)

type tviewLogger struct {
	*tview.TextView
}

func (s *UIState) Logger() *tviewLogger {
	return &tviewLogger{s.textWindow}
}

func (l *tviewLogger) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
}

func (l *tviewLogger) Error(v ...interface{}) {
	fmt.Fprintln(l.TextView, fmt.Sprintf("[%s::b]ERROR:[-::-]", util.ColortoHexString(activeTheme.ErrorColor)), fmt.Sprint(v...))
	log.Println("[ERROR]", fmt.Sprint(v...))
	l.ScrollToEnd()
}

func (l *tviewLogger) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

func (l *tviewLogger) Info(v ...interface{}) {
	fmt.Fprintln(l.TextView, fmt.Sprintf("[%s::b]INFO:[-::-]", util.ColortoHexString(activeTheme.InfoColor)), fmt.Sprint(v...))
	log.Println("[INFO]", fmt.Sprint(v...))
	l.ScrollToEnd()
}
