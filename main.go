package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type appTheme struct {
	BackgroundColor     tcell.Color
	BorderColor         tcell.Color
	CategoryNodeColor   tcell.Color
	FolderNodeColor     tcell.Color
	ItemNodeColor       tcell.Color
	ActionNodeColor     tcell.Color
	LoadingColor        tcell.Color
	LiveColor           tcell.Color
	UpdateColor         tcell.Color
	NoContentColor      tcell.Color
	InfoColor           tcell.Color
	ErrorColor          tcell.Color
	TerminalAccentColor tcell.Color
	TerminalTextColor   tcell.Color
}

type viewerSession struct {
	cfg config

	// tview
	app        *tview.Application
	textWindow *tview.TextView
	tree       *tview.TreeView
}

var activeTheme appTheme

func main() {
	cfg, err := loadConfig()
	if err != nil {
		log.Fatal("Could not open config: ", err)
	}

	logFile, err := configureLogging(cfg)
	if err != nil {
		log.Fatal(err)
	}
	defer logFile.Close()

	session := newSession(cfg)

	go session.checkLive()
	go session.CheckUpdate()

	// set vod types nodes
	go func() {
		nodes, err := session.getVodTypeNodes()
		if err != nil {
			session.logError(err)
		} else {
			appendNodes(session.tree.GetRoot(), nodes...)
			session.app.Draw()
		}
	}()

	session.tree.GetRoot().AddChild(session.getCollectionsNode())

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
}

func newSession(cfg config) (session *viewerSession) {
	// set defaults
	session = &viewerSession{}
	session.cfg = cfg
	cfg.Theme.apply()

	session.app = tview.NewApplication()

	// build base tree
	root := tview.NewTreeNode("Categories").
		SetSelectable(false)
	session.tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root).
		SetTopLevel(1)

	// set full race weekends node
	fullSessions := tview.NewTreeNode("Full Race Weekends").
		SetColor(activeTheme.CategoryNodeColor)
	fullSessions.SetSelectedFunc(session.withBlink(fullSessions, func() {
		fullSessions.SetSelectedFunc(nil)
		seasons, err := session.getSeasonNodes()
		if err != nil {
			session.logError(err)
		} else {
			appendNodes(fullSessions, seasons...)
		}
	}))
	session.tree.GetRoot().AddChild(fullSessions)
	session.tree.SetSelectedFunc(session.toggleVisibility)
	// flex containing everything
	flex := tview.NewFlex()
	if session.cfg.HorizontalLayout {
		flex.SetDirection(tview.FlexRow)
	}
	// debug window
	session.textWindow = tview.NewTextView().SetWordWrap(false).SetWrap(false)
	session.textWindow.SetDynamicColors(true)
	session.textWindow.SetBorder(true)
	session.textWindow.SetChangedFunc(func() {
		session.app.Draw()
	})
	flex.AddItem(session.tree, 0, cfg.TreeRatio, true)
	flex.AddItem(session.textWindow, 0, cfg.OutputRatio, false)
	go func() {
		err := session.app.SetRoot(flex, true).Run()
		if err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	return
}

func (session *viewerSession) checkLive() {
	for {
		session.logInfo("checking for live session")
		isLive, liveNode, err := session.getLiveNode()
		if err != nil {
			session.logError("error looking for live session: ", err)
			if session.cfg.LiveRetryTimeout <= 0 {
				return
			}
		} else if isLive {
			insertNodeAtTop(session.tree.GetRoot(), liveNode)
			if session.app != nil {
				session.app.Draw()
			}
			return
		} else if session.cfg.LiveRetryTimeout <= 0 {
			session.logInfo("no live session found")
			return
		} else {
			session.logInfo("no live session found")
		}
		time.Sleep(time.Second * time.Duration(session.cfg.LiveRetryTimeout))
	}
}

func (session *viewerSession) CheckUpdate() {
	if !session.cfg.CheckUpdate {
		return
	}
	node, err := session.getUpdateNode()
	if err != nil {
		session.logInfo(err)
	} else {
		session.logInfo("Newer version found!")
		if re, ok := node.GetReference().(release); ok {
			fmt.Fprintln(session.textWindow, "\n[blue::bu]"+re.Name+"[-::-]\n")
			fmt.Fprintln(session.textWindow, re.Body)
		}
		insertNodeAtTop(session.tree.GetRoot(), node)
		session.app.Draw()
	}
}
