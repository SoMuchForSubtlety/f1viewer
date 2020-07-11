package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/SoMuchForSubtlety/keyring"
	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"gopkg.in/alecthomas/kingpin.v2"
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

	ring      keyring.Keyring
	username  string
	password  string
	authtoken string
	// tview
	app        *tview.Application
	textWindow *tview.TextView
	tree       *tview.TreeView
}

var (
	version = "dev"
	commit  = ""
	date    = ""
)

func main() {
	app := kingpin.New("f1viewer", "a TUI client for F1TV")
	app.Version(buildVersion(version, commit, date))
	app.VersionFlag.Short('v')
	app.HelpFlag.Short('h')

	kingpin.MustParse(app.Parse(os.Args[1:]))

	session, logfile, err := newSession()
	defer logfile.Close()
	if err != nil {
		fmt.Println("[ERROR]", err)
		log.Fatal(err)
	}
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

func newSession() (*viewerSession, *os.File, error) {
	var err error
	session := &viewerSession{}

	session.cfg, err = loadConfig()
	if err != nil {
		return nil, nil, fmt.Errorf("Could not open config: %w", err)
	}

	logFile, err := configureLogging(session.cfg)
	if err != nil {
		return nil, nil, err
	}

	err = session.openRing()
	if err != nil {
		session.logError(fmt.Errorf("Could not access credential store: %w", err))
	}

	err = session.loadCredentials()
	if err != nil {
		session.logError(err)
	}

	session.app = tview.NewApplication()
	session.app.EnableMouse(true)

	root := tview.NewTreeNode("Categories").SetSelectable(false)
	root.AddChild(session.getFullSessionsNode())
	session.tree = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root).
		SetTopLevel(1)

	session.textWindow = tview.NewTextView().
		SetWordWrap(false).
		SetWrap(false).
		SetDynamicColors(true).
		SetChangedFunc(func() { session.app.Draw() })
	session.textWindow.SetBorder(true)

	session.tree.SetSelectedFunc(session.toggleVisibility)

	token, err := session.login()
	if err != nil {
		session.initUIWithForm()
	} else {
		session.logInfo("logged in!")
		session.authtoken = token
		session.initUI()
	}

	return session, logFile, nil
}

func (session *viewerSession) initUIWithForm() {

	form := tview.NewForm().
		AddInputField("username", session.username, 30, nil, session.updateUsername).
		AddPasswordField("password", "", 30, '*', session.updatePassword).
		AddButton("test", session.testAuth).
		AddButton("save", session.closeForm)

	formTreeFlex := tview.NewFlex()
	if !session.cfg.HorizontalLayout {
		formTreeFlex.SetDirection(tview.FlexRow)
	}

	if session.cfg.HorizontalLayout {
		formTreeFlex.
			AddItem(form, 50, 0, true).
			AddItem(session.tree, 0, 1, false)
	} else {
		formTreeFlex.
			AddItem(form, 7, 0, true).
			AddItem(session.tree, 0, 1, false)
	}

	masterFlex := tview.NewFlex()
	if session.cfg.HorizontalLayout {
		masterFlex.SetDirection(tview.FlexRow)
	}

	masterFlex.
		AddItem(formTreeFlex, 0, session.cfg.TreeRatio, true).
		AddItem(session.textWindow, 0, session.cfg.OutputRatio, false)

	session.app.SetRoot(masterFlex, true)
}

func (session *viewerSession) initUI() {
	flex := tview.NewFlex().
		AddItem(session.tree, 0, session.cfg.TreeRatio, true).
		AddItem(session.textWindow, 0, session.cfg.OutputRatio, false)

	if session.cfg.HorizontalLayout {
		flex.SetDirection(tview.FlexRow)
	}

	session.app.SetRoot(flex, true)
}

func (session *viewerSession) closeForm() {
	session.testAuth()
	err := session.saveCredentials()
	if err != nil {
		session.logError(err)
	}
	session.initUI()
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

func buildVersion(version, commit, date string) string {
	result := fmt.Sprintf("Version:     %s", version)
	if commit != "" {
		result += fmt.Sprintf("\nGit commit:  %s", commit)
	}
	if date != "" {
		result += fmt.Sprintf("\nBuilt:       %s", date)
	}
	return result
}
