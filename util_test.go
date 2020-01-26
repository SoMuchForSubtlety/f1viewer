package main

import (
	"errors"
	"runtime"
	"testing"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestSanitizeFileName(t *testing.T) {
	title := `file name: "with" <illegal> \/cahracters|`
	var target string
	title = sanitizeFileName(title)
	if runtime.GOOS == "windows" {
		target = `file name with illegal cahracters`
	} else {
		target = `file name: "with" <illegal> \ cahracters|`
	}

	assert.Equal(t, target, title)
}

var colorPairs = []struct {
	hex  string
	name string
}{
	{hex: "#bc8f8f", name: "rosybrown"},
	{hex: "#fff5ee", name: "seashell"},
	{hex: "#00ff7f", name: "springgreen"},
	{hex: "#ffe4c4", name: "bisque"},
	{hex: "#2f4f4f", name: "darkslategrey"},
	{hex: "#b8860b", name: "darkgoldenrod"},
	{hex: "#e0ffff", name: "lightcyan"},
	{hex: "#66cdaa", name: "mediumaquamarine"},
	{hex: "#ffdab9", name: "peachpuff"},
	{hex: "#f4a460", name: "sandybrown"},
	{hex: "#d8bfd8", name: "thistle"},
	{hex: "#d3d3d3", name: "lightgrey"},
	{hex: "#808080", name: "gray"},
	{hex: "#a52a2a", name: "brown"},
	{hex: "#e9967a", name: "darksalmon"},
	{hex: "#dda0dd", name: "plum"},
	{hex: "#708090", name: "slategray"},
	{hex: "#ffffff", name: "white"},
}

func TestColorToHexString(t *testing.T) {
	for _, s := range colorPairs {
		hex := colortoHexString(tcell.GetColor(s.name))
		assert.Equal(t, s.hex, hex)
	}
}

func TestHexStringToColor(t *testing.T) {
	for _, s := range colorPairs {
		color := hexStringToColor(s.hex)
		assert.Equal(t, color.Hex(), tcell.GetColor(s.name).Hex())
	}
}

func TestWithBlink(t *testing.T) {
	// TODO add check for colors
	originalScreen := `
node title┌────────┐
          │        │
          │        │
          │        │
          └────────┘`

	loadingScreen := `
loading...┌────────┐
          │        │
          │        │
          │        │
          └────────┘`

	originalText := "node title"
	originalColor := tcell.ColorViolet
	node := tview.NewTreeNode(originalText)
	node.SetColor(originalColor)

	simScreen, s := newTestApp(20, 5)
	s.tree.GetRoot().AddChild(node)
	go s.app.Run()

	go s.withBlink(node, func() {
		time.Sleep(time.Millisecond * 200)
	})()

	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, loadingScreen, toTextScreen(simScreen))

	time.Sleep(time.Millisecond * 500)

	assert.Equal(t, originalScreen, toTextScreen(simScreen))
	assert.Equal(t, originalColor, node.GetColor())
	assert.Equal(t, originalText, node.GetText())
}

func TestGetYearAndRace(t *testing.T) {
	// TODO add checks for post 2020 events
	year, race, err := getYearAndRace("1914_ITA_FP2_F1TV")
	assert.Nil(t, err)
	assert.Equal(t, "2019", year)
	assert.Equal(t, "14", race)

	year, race, err = getYearAndRace("9414_ABC")
	assert.Nil(t, err)
	assert.Equal(t, "1994", year)
	assert.Equal(t, "14", race)

	year, race, err = getYearAndRace("2018_TEST")
	assert.Nil(t, err)
	assert.Equal(t, "2018", year)
	assert.Equal(t, "0", race)

	_, _, err = getYearAndRace("abcde")
	assert.EqualError(t, err, "not a valid YearRaceID")

	_, _, err = getYearAndRace("123")
	assert.EqualError(t, err, "not long enough")
}

func TestLog(t *testing.T) {
	expectedInfo := `
               ┌─────────────┐
               │INFO: info   │
               │             │
               │             │
               └─────────────┘`

	expectedError := `
               ┌─────────────┐
               │INFO: info   │
               │ERROR: test  │
               │             │
               └─────────────┘`

	simScreen, s := newTestApp(30, 5)
	go s.app.Run()
	s.logInfo("info")
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, expectedInfo, toTextScreen(simScreen))
	s.logError(errors.New("test"))
	time.Sleep(time.Millisecond * 100)
	assert.Equal(t, expectedError, toTextScreen(simScreen))
}

func toTextScreen(screen tcell.SimulationScreen) string {
	content := "\n"
	contents, width, _ := screen.GetContents()
	var cursor int
	for _, cell := range contents {
		if cursor >= width {
			content += "\n"
			cursor = 0
		}
		content += string(cell.Bytes)
		cursor++
	}
	return content
}

func newTestApp(x, y int) (tcell.SimulationScreen, viewerSession) {
	simScreen := tcell.NewSimulationScreen("UTF-8")
	simScreen.Init()
	simScreen.SetSize(x, y)

	app := tview.NewApplication()
	app.SetScreen(simScreen)

	text := tview.NewTextView().
		SetWordWrap(false).
		SetWrap(false).
		SetDynamicColors(true).
		SetChangedFunc(func() {
			app.Draw()
		})

	text.SetBorder(true)

	tree := tview.NewTreeView().
		SetRoot(tview.NewTreeNode("root")).
		SetTopLevel(1)

	flex := tview.NewFlex()
	flex.AddItem(tree, 0, 1, true)
	flex.AddItem(text, 0, 1, false)

	app.SetRoot(flex, true)

	return simScreen, viewerSession{tree: tree, app: app, textWindow: text}
}
