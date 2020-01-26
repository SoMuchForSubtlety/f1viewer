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

var activeTheme = struct {
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
}{
	CategoryNodeColor:   tcell.ColorOrange,
	FolderNodeColor:     tcell.ColorWhite,
	ItemNodeColor:       tcell.ColorLightGreen,
	ActionNodeColor:     tcell.ColorDarkCyan,
	LoadingColor:        tcell.ColorDarkCyan,
	LiveColor:           tcell.ColorRed,
	UpdateColor:         tcell.ColorDarkRed,
	NoContentColor:      tcell.ColorOrangeRed,
	InfoColor:           tcell.ColorGreen,
	ErrorColor:          tcell.ColorRed,
	TerminalAccentColor: tcell.ColorGreen,
	TerminalTextColor:   tview.Styles.PrimaryTextColor,
}

type viewerSession struct {
	cfg config

	// tview
	app        *tview.Application
	textWindow *tview.TextView
	tree       *tview.TreeView
}

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
	go func() {
		if err := session.app.Run(); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()

	go session.checkLive()
	go session.CheckUpdate()

	// set vod types nodes

	session.tree.GetRoot().AddChild(session.getCollectionsNode())
	nodes, err := session.getVodTypeNodes()
	if err != nil {
		session.logError(err)
	} else {
		appendNodes(session.tree.GetRoot(), nodes...)
		session.app.Draw()
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	<-c
}

func newSession(cfg config) *viewerSession {
	cfg.Theme.apply()
	root := tview.NewTreeNode("Categories").
		SetSelectable(false)

	session := &viewerSession{
		cfg: cfg,
		app: tview.NewApplication(),
		tree: tview.NewTreeView().
			SetRoot(root).
			SetCurrentNode(root).
			SetTopLevel(1),
		textWindow: tview.NewTextView().
			SetWordWrap(false).
			SetWrap(false).
			SetDynamicColors(true),
	}

	root.AddChild(session.getFullSessionsNode())
	session.textWindow.SetBorder(true)
	session.tree.SetSelectedFunc(session.toggleVisibility)
	session.textWindow.SetChangedFunc(func() {
		session.app.Draw()
	})

	flex := tview.NewFlex().
		AddItem(session.tree, 0, cfg.TreeRatio, true).
		AddItem(session.textWindow, 0, cfg.OutputRatio, false)

	if session.cfg.HorizontalLayout {
		flex.SetDirection(tview.FlexRow)
	}
	session.app.SetRoot(flex, true)

	return session
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
