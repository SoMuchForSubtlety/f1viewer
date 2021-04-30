package ui

import (
	"fmt"
	"io"
	"log"

	"github.com/SoMuchForSubtlety/f1viewer/internal/util"
)

type tviewLogger struct {
	io.Writer
}

func (s *UIState) Logger() *tviewLogger {
	return &tviewLogger{s.textWindow}
}

func (l *tviewLogger) Errorf(format string, v ...interface{}) {
	l.Error(fmt.Sprintf(format, v...))
}

func (l *tviewLogger) Error(v ...interface{}) {
	fmt.Fprintln(l.Writer, fmt.Sprintf("[%s::b]ERROR:[-::-]", util.ColortoHexString(activeTheme.ErrorColor)), fmt.Sprint(v...))
	log.Println("[ERROR]", fmt.Sprint(v...))
}

func (l *tviewLogger) Infof(format string, v ...interface{}) {
	l.Info(fmt.Sprintf(format, v...))
}

func (l *tviewLogger) Info(v ...interface{}) {
	fmt.Fprintln(l.Writer, fmt.Sprintf("[%s::b]INFO:[-::-]", util.ColortoHexString(activeTheme.InfoColor)), fmt.Sprint(v...))
	log.Println("[INFO]", fmt.Sprint(v...))
}
