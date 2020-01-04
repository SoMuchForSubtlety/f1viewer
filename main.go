package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"time"

	"github.com/atotto/clipboard"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
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

type viewerSession struct {
	con config

	vodTypes vodTypesStruct

	abortWritingInfo chan bool

	// cache
	episodeMap      map[string]episodeStruct
	driverMap       map[string]driverStruct
	teamMap         map[string]teamStruct
	episodeMapMutex sync.RWMutex
	teamMapMutex    sync.RWMutex
	driverMapMutex  sync.RWMutex

	// tview
	app       *tview.Application
	debugText *tview.TextView
	tree      *tview.TreeView
}

func newSession(cfg config) (session *viewerSession) {
	// set defaults
	session = &viewerSession{}
	session.con = cfg

	// cache
	session.episodeMap = make(map[string]episodeStruct)
	session.driverMap = make(map[string]driverStruct)
	session.teamMap = make(map[string]teamStruct)

	session.app = tview.NewApplication()

	// build base tree
	root := tview.NewTreeNode("VOD-Types").
		SetColor(tcell.ColorBlue).
		SetSelectable(false)
	session.tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root)

	var allSeasons allSeasonStruct
	// set full race weekends node
	fullSessions := tview.NewTreeNode("Full Race Weekends").
		SetSelectable(true).
		SetReference(allSeasons).
		SetColor(tcell.ColorYellow)
	session.tree.GetRoot().AddChild(fullSessions)
	session.tree.SetSelectedFunc(session.nodeSelected)
	// flex containing everything
	flex := tview.NewFlex()
	// debug window
	session.debugText = tview.NewTextView().SetWordWrap(false).SetWrap(false)
	session.debugText.SetDynamicColors(true)
	session.debugText.SetBorder(true).SetTitle("Info")
	session.debugText.SetChangedFunc(func() {
		session.app.Draw()
	})
	flex.AddItem(session.tree, 0, 1, true)
	flex.AddItem(session.debugText, 0, 1, false)
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
	node, err := getUpdateNode()
	if err != nil {
		session.logInfo(err)
	} else {
		session.logInfo("Newer version found!")
		if re, ok := node.GetReference().(release); ok {
			fmt.Fprintln(session.debugText, "\n[blue::bu]"+re.Name+"[-::-]\n")
			fmt.Fprintln(session.debugText, re.Body)
		}
		insertNodeAtTop(session.tree.GetRoot(), node)
		session.app.Draw()
	}
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
	signal.Notify(c, os.Interrupt)

	<-c
}

func (session *viewerSession) nodeSelected(node *tview.TreeNode) {
	reference := node.GetReference()
	children := node.GetChildren()
	if node.GetText() == "loading..." {
		// Selecting the root node or a loading node does nothing
		return
	} else if len(children) > 0 {
		// Collapse if visible, expand if collapsed.
		node.SetExpanded(!node.IsExpanded())
	} else if coll, ok := reference.(collection); ok {
		session.loadCollectionContent(coll.Self, node)
	} else if event, ok := reference.(eventStruct); ok {
		// if event (eg. Australian GP 2018) is selected from full race weekends
		done := false
		hasSessions := false
		go func() {
			sessions, err := session.getSessionNodes(event)
			if err != nil {
				session.logError(err)
				hasSessions = true
			} else {
				for _, session := range sessions {
					if session != nil && len(session.GetChildren()) > 0 {
						hasSessions = true
						node.AddChild(session)
					}
				}
			}
			done = true
		}()
		go func() {
			session.blinkNode(node, &done, tcell.ColorWhite)
			if !hasSessions {
				node.SetColor(tcell.ColorRed)
				node.SetText(node.GetText() + " - NO CONTENT AVAILABLE")
				node.SetSelectable(false)
			}
			session.app.Draw()
		}()
	} else if season, ok := reference.(seasonStruct); ok {
		// if full season is selected from full race weekends
		done := false
		go func() {
			events, err := getEventNodes(season)
			if err != nil {
				session.logError(err)
			} else {
				for _, event := range events {
					if event != nil {
						layout := "2006-01-02"
						e := event.GetReference().(eventStruct)
						t, _ := time.Parse(layout, e.StartDate)
						if t.Before(time.Now().AddDate(0, 0, 1)) {
							node.AddChild(event)
						}
					}
				}
			}
			done = true
		}()
		go session.blinkNode(node, &done, tcell.ColorWheat)
	} else if context, ok := reference.(commandContext); ok {
		go func() {
			err := session.runCustomCommand(context, node)
			if err != nil {
				session.logError(err)
			}
		}()
	} else if i, ok := reference.(int); ok {
		// if episodes for category are not loaded yet
		if i < len(session.vodTypes.Objects) {
			done := false
			go func() {
				episodes, err := session.getEpisodeNodes(session.vodTypes.Objects[i].ContentUrls)
				if err != nil {
					session.logError(err)
				} else {
					appendNodes(node, episodes...)
				}
				done = true
			}()
			go session.blinkNode(node, &done, tcell.ColorYellow)
		}
	} else if _, ok := reference.(allSeasonStruct); ok {
		done := false
		go func() {
			seasons, err := getSeasonNodes()
			if err != nil {
				session.logError(err)
			} else {
				appendNodes(node, seasons...)
				node.SetReference(seasons)
			}
			done = true
		}()
		go session.blinkNode(node, &done, tcell.ColorYellow)
	} else if node.GetText() == "Play with MPV" {
		go func() {
			url, err := getPlayableURL(reference.(string))
			if err != nil {
				session.logError(err)
				return
			}
			cmd := exec.Command("mpv", url, "--alang="+session.con.Lang, "--start=0", "--quiet")
			session.runCmd(cmd)
		}()
	} else if node.GetText() == "Download .m3u8" {
		go func() {
			node.SetColor(tcell.ColorBlue)
			urlAndTitle := reference.([]string)
			url, err := getPlayableURL(urlAndTitle[0])
			if err != nil {
				session.logError(err)
				return
			}
			_, _, err = downloadAsset(url, urlAndTitle[1])
			if err != nil {
				session.logError(err)
			}
		}()
	} else if node.GetText() == "Copy URL to clipboard" || node.GetText() == "URL copied to clipboard" {
		go func() {
			url, err := getPlayableURL(reference.(string))
			if err != nil {
				session.logError(err)
				return
			}
			err = clipboard.WriteAll(url)
			if err != nil {
				session.logError(err)
				return
			}
			node.SetText("URL copied to clipboard")
			node.SetColor(tcell.ColorBlue)
			session.app.Draw()
		}()
	} else if node.GetText() == "download update" {
		err := openbrowser("https://github.com/SoMuchForSubtlety/F1viewer/releases/latest")
		if err != nil {
			session.logError(err)
		}
	} else if node.GetText() == "don't tell me about updates" {
		session.con.CheckUpdate = false
		err := session.con.save()
		if err != nil {
			session.logError(err)
		}
		node.SetColor(tcell.ColorBlue)
		node.SetText("update notifications turned off")
	}
}

func (session *viewerSession) loadCollections() error {
	node := tview.NewTreeNode("Collections").SetSelectable(true).SetColor(tcell.ColorYellow).SetExpanded(false)
	list, err := getCollectionList()
	if err != nil {
		return err
	}
	for _, coll := range list.Objects {
		child := tview.NewTreeNode(coll.Title).SetExpanded(false).SetReference(coll)
		node.AddChild(child)
	}
	session.tree.GetRoot().AddChild(node)
	return nil
}

func (session *viewerSession) loadCollectionContent(collID string, parent *tview.TreeNode) error {
	nodes, err := session.getCollectionContent(collID)
	if err == nil {
		appendNodes(parent, nodes...)
	}
	parent.Expand()
	return err
}
