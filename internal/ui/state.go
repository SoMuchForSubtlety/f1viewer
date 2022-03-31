package ui

import (
	"errors"
	"fmt"
	"time"

	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/cmd"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/config"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/creds"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/github"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
	"github.com/SoMuchForSubtlety/f1viewer/v2/pkg/f1tv/v2"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
	"github.com/zalando/go-keyring"
)

// TODO: rework
var activeTheme = struct {
	CategoryNodeColor   tcell.Color
	FolderNodeColor     tcell.Color
	ItemNodeColor       tcell.Color
	ActionNodeColor     tcell.Color
	LoadingColor        tcell.Color
	LiveColor           tcell.Color
	MultiCommandColor   tcell.Color
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
	MultiCommandColor:   tcell.ColorAquaMarine,
	UpdateColor:         tcell.ColorDarkRed,
	NoContentColor:      tcell.ColorOrangeRed,
	InfoColor:           tcell.ColorGreen,
	ErrorColor:          tcell.ColorRed,
	TerminalAccentColor: tcell.ColorGreen,
	TerminalTextColor:   tview.Styles.PrimaryTextColor,
}

type UIState struct {
	version string
	cfg     config.Config

	app *tview.Application

	textWindow *tview.TextView
	treeView   *tview.TreeView

	LiveNode *tview.TreeNode

	logger util.Logger

	// TODO: replace activeTheme
	// theme config.Theme

	v2 *f1tv.F1TV

	cmd *cmd.Store

	liveSessions map[string]struct{}
}

func NewUI(cfg config.Config, version string) *UIState {
	ui := UIState{
		version:      version,
		cfg:          cfg,
		v2:           f1tv.NewF1TV(version),
		liveSessions: make(map[string]struct{}),
	}
	ui.applyTheme(cfg.Theme)

	ui.app = tview.NewApplication()
	ui.app.EnableMouse(cfg.EnableMouse)

	root := tview.NewTreeNode("Categories").SetSelectable(false)

	ui.treeView = tview.NewTreeView().
		SetRoot(root).
		SetCurrentNode(root).
		SetTopLevel(1)

	// refresh supported nodes on 'r' key press or quit on 'q'
	ui.treeView.SetInputCapture(ui.TreeInputHanlder)

	ui.textWindow = tview.NewTextView().
		SetWordWrap(false).
		SetWrap(cfg.TerminalWrap).
		SetDynamicColors(true).
		SetChangedFunc(func() { ui.app.Draw() })
	ui.textWindow.SetBorder(true)

	ui.treeView.SetSelectedFunc(ui.toggleVisibility)

	ui.logger = ui.Logger()

	ui.cmd = cmd.NewStore(cfg.CustomPlaybackOptions, cfg.MultiCommand, cfg.Lang, ui.logger, activeTheme.TerminalAccentColor)

	err := ui.loginWithStoredCredentials()
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			ui.logger.Errorf("could not get credentials: %s", err.Error())
		}
		ui.initUIWithForm()
	} else {
		ui.logger.Info("logged in!")
		ui.initUI()
	}

	homepageContent := tview.NewTreeNode("homepage").
		SetColor(activeTheme.CategoryNodeColor).
		SetReference(&NodeMetadata{nodeType: CategoryNode, metadata: cmd.MetaData{}}).
		SetExpanded(true)
	homepageContent.SetSelectedFunc(ui.withBlink(homepageContent, func() {
		homepageContent.SetSelectedFunc(nil)
		appendNodes(homepageContent, ui.getPageNodes(f1tv.PAGE_HOMEPAGE)...)
	}, nil))

	appendNodes(root,
		ui.pageNode(f1tv.PAGE_HOMEPAGE, "Homepage"),
		ui.pageNode(f1tv.PAGE_SEASON_2022, "2022 Season"),
		ui.pageNode(f1tv.PAGE_ARCHIVE, "Archive"),
		ui.pageNode(f1tv.PAGE_DOCUMENTARIES, "Documentaries"),
		ui.pageNode(f1tv.PAGE_SHOWS, "Shows"),
	)

	return &ui
}

func (ui *UIState) pageNode(id f1tv.PageID, title string) *tview.TreeNode {
	node := tview.NewTreeNode(title).
		SetColor(activeTheme.FolderNodeColor).
		SetReference(&NodeMetadata{nodeType: CategoryNode, metadata: cmd.MetaData{}}).
		SetExpanded(true)
	node.SetSelectedFunc(ui.withBlink(node, func() {
		node.SetSelectedFunc(nil)
		appendNodes(node, ui.getPageNodes(id)...)
	}, nil))

	return node
}

func (ui *UIState) Stop() {
	ui.app.Stop()
}

func (ui *UIState) Run() error {
	done := make(chan error)
	go func() {
		done <- ui.app.Run()
	}()

	go ui.checkLive()
	go ui.loadUpdate()

	logOutNode := tview.NewTreeNode("Log Out").
		SetReference(&NodeMetadata{nodeType: ActionNode}).
		SetColor(activeTheme.ActionNodeColor)
	logOutNode.SetSelectedFunc(ui.logout)

	ui.treeView.GetRoot().AddChild(logOutNode)

	return <-done
}

func (s *UIState) logout() {
	if err := creds.RemoveCredentials(); err != nil {
		s.logger.Error(err)
	}
	s.initUIWithForm()
}

func (s *UIState) loginWithStoredCredentials() (err error) {
	var username, password string

	if s.cfg.UseEnvironmentCredentials {
		username, password, err = creds.LoadEnvCredentials()
	} else {
		username, password, err = creds.LoadCredentials()
	}

	if err != nil {
		return err
	}

	return s.login(username, password)
}

func (s *UIState) login(username, pw string) error {
	err := s.v2.Authenticate(username, pw, s.logger)
	return err
}

func (s *UIState) initUIWithForm() {
	username, _, _ := creds.LoadCredentials()
	pw := ""
	form := tview.NewForm().
		AddInputField("email", username, 30, nil, func(text string) { username = text }).
		AddPasswordField("password", "", 30, '*', func(text string) { pw = text }).
		AddButton("test", func() {
			err := s.login(username, pw)
			if err == nil {
				s.logger.Info("credentials accepted")
			} else {
				s.logger.Error(err)
			}
		}).
		AddButton("save", func() { s.closeForm(username, pw) })

	formTreeFlex := tview.NewFlex()
	if !s.cfg.HorizontalLayout {
		formTreeFlex.SetDirection(tview.FlexRow)
	}

	if s.cfg.HorizontalLayout {
		formTreeFlex.
			AddItem(form, 50, 0, true).
			AddItem(s.treeView, 0, 1, false)
	} else {
		formTreeFlex.
			AddItem(form, 7, 0, true).
			AddItem(s.treeView, 0, 1, false)
	}

	masterFlex := tview.NewFlex()
	if s.cfg.HorizontalLayout {
		masterFlex.SetDirection(tview.FlexRow)
	}

	masterFlex.
		AddItem(formTreeFlex, 0, s.cfg.TreeRatio, true).
		AddItem(s.textWindow, 0, s.cfg.OutputRatio, false)

	s.app.SetRoot(masterFlex, true)
}

func (s *UIState) initUI() {
	flex := tview.NewFlex().
		AddItem(s.treeView, 0, s.cfg.TreeRatio, true).
		AddItem(s.textWindow, 0, s.cfg.OutputRatio, false)

	if s.cfg.HorizontalLayout {
		flex.SetDirection(tview.FlexRow)
	}

	s.app.SetRoot(flex, true)
}

func (s *UIState) closeForm(username, pw string) {
	if err := s.login(username, pw); err != nil {
		s.logger.Error(err)
	}
	if err := creds.SaveCredentials(username, pw); err != nil {
		s.logger.Error(err)
	}
	s.initUI()
}

func (s *UIState) withBlink(node *tview.TreeNode, fn func(), after func()) func() {
	return func() {
		done := make(chan struct{})
		go func() {
			fn()
			done <- struct{}{}
		}()
		go func() {
			s.blinkNode(node, done)
			if after != nil {
				after()
			}
		}()
	}
}

func (s *UIState) blinkNode(node *tview.TreeNode, done chan struct{}) {
	originalText := node.GetText()
	originalColor := node.GetColor()
	color1 := originalColor
	color2 := activeTheme.LoadingColor
	node.SetText("loading...")

	ticker := time.NewTicker(200 * time.Millisecond)
	for {
		select {
		case <-done:
			node.SetText(originalText)
			node.SetColor(originalColor)
			s.app.Draw()
			return
		case <-ticker.C:
			node.SetColor(color2)
			s.app.Draw()
			color1, color2 = color2, color1
		}
	}
}

func (ui *UIState) addLiveNode(newNode *tview.TreeNode) {
	if ui.LiveNode != nil {
		ui.treeView.GetRoot().RemoveChild(ui.LiveNode)
	}

	// newNode if nil if the previous session is no longer live and there is no new live session
	if newNode != nil {
		ui.LiveNode = newNode
		insertNodeAtTop(ui.treeView.GetRoot(), newNode)
	}
	ui.app.Draw()
}

func (ui *UIState) loadUpdate() {
	release, new, err := github.CheckUpdate(ui.version)
	if err != nil {
		ui.logger.Error("failed to check for update: ", err)
	}
	if !new {
		return
	}

	ui.logger.Info("New version found!")
	ui.logger.Info(release.TagName)
	fmt.Fprintln(ui.logger, "\n[blue::bu]"+release.Name+"[-::-]")
	fmt.Fprintln(ui.logger, release.Body+"\n")

	updateNode := tview.NewTreeNode("UPDATE AVAILABLE").
		SetColor(activeTheme.UpdateColor).
		SetExpanded(false)
	getUpdateNode := tview.NewTreeNode("download update").
		SetColor(activeTheme.ActionNodeColor).
		SetSelectedFunc(func() {
			err := util.Open("https://github.com/SoMuchForSubtlety/f1viewer/releases/latest")
			if err != nil {
				ui.logger.Error(err)
			}
		})

	appendNodes(updateNode, getUpdateNode)
	insertNodeAtTop(ui.treeView.GetRoot(), updateNode)
}
