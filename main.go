package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
)

type viewerSession struct {
	con config

	abortWritingInfo chan bool

	// cache
	episodeMap      map[string]episode
	episodeMapMutex sync.RWMutex
	teamMapMutex    sync.RWMutex
	driverMapMutex  sync.RWMutex

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

	err = session.loadCollections()
	if err != nil {
		session.logError(err)
	}

	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, os.Kill)

	<-c
}

func newSession(cfg config) (session *viewerSession) {
	// set defaults
	session = &viewerSession{}
	session.con = cfg
	cfg.Theme.apply()

	// cache
	session.episodeMap = make(map[string]episode)

	session.app = tview.NewApplication()

	// build base tree
	root := tview.NewTreeNode("Categories").
		SetSelectable(false)
	session.tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	// set full race weekends node
	fullSessions := tview.NewTreeNode("Full Race Weekends").
		SetColor(tview.Styles.SecondaryTextColor)
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
	session.tree.SetSelectedFunc(session.nodeSelected)
	// flex containing everything
	flex := tview.NewFlex()
	flex.SetDirection(tview.FlexRow)
	// debug window
	session.textWindow = tview.NewTextView().SetWordWrap(false).SetWrap(false)
	session.textWindow.SetDynamicColors(true)
	session.textWindow.SetBorder(true)
	session.textWindow.SetChangedFunc(func() {
		session.app.Draw()
	})
	flex.AddItem(session.tree, 0, 1, true)
	flex.AddItem(session.textWindow, 0, 1, false)
	go func() {
		session.app.SetRoot(flex, true).Run()
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
		} else if isLive {
			insertNodeAtTop(session.tree.GetRoot(), liveNode)
			if session.app != nil {
				session.app.Draw()
			}
			return
		} else if session.con.LiveRetryTimeout < 0 {
			session.logInfo("no live session found")
			return
		} else {
			session.logInfo("no live session found")
		}
		time.Sleep(time.Second * time.Duration(session.con.LiveRetryTimeout))
	}
}

func (session *viewerSession) CheckUpdate() {
	if !session.con.CheckUpdate {
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

func (session *viewerSession) nodeSelected(node *tview.TreeNode) {
	children := node.GetChildren()
	if len(children) > 0 {
		// Collapse if visible, expand if collapsed.
		node.SetExpanded(!node.IsExpanded())
	}
}

func (session *viewerSession) loadCollections() error {
	node := tview.NewTreeNode("Collections").SetColor(tview.Styles.SecondaryTextColor).SetExpanded(false)
	list, err := getCollectionList()
	if err != nil {
		return err
	}
	for _, coll := range list.Objects {
		child := tview.NewTreeNode(coll.Title)
		collID := coll.Self
		child.SetSelectedFunc(session.withBlink(child, func() {
			child.SetSelectedFunc(nil)
			var nodes []*tview.TreeNode
			nodes, err = session.getCollectionContent(collID)
			if err != nil {
				session.logError(err)
			} else if len(nodes) > 0 {
				appendNodes(child, nodes...)
			} else {
				child.AddChild(tview.NewTreeNode("no content").SetColor(tcell.ColorRed))
			}
		}))
		node.AddChild(child)
	}
	session.tree.GetRoot().AddChild(node)
	return nil
}
