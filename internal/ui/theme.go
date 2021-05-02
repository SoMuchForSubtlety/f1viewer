package ui

import (
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/config"
	"github.com/SoMuchForSubtlety/f1viewer/v2/internal/util"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// TODO: replace with non global
func (ui *UIState) applyTheme(t config.Theme) {
	if t.TerminalTextColor != "" {
		tview.Styles.PrimaryTextColor = util.HexStringToColor(t.TerminalTextColor)
	}
	if t.CategoryNodeColor != "" {
		activeTheme.CategoryNodeColor = util.HexStringToColor(t.CategoryNodeColor)
	}
	if t.FolderNodeColor != "" {
		activeTheme.FolderNodeColor = util.HexStringToColor(t.FolderNodeColor)
	}
	if t.ItemNodeColor != "" {
		activeTheme.ItemNodeColor = util.HexStringToColor(t.ItemNodeColor)
	}
	if t.ActionNodeColor != "" {
		activeTheme.ActionNodeColor = util.HexStringToColor(t.ActionNodeColor)
	}
	if t.BackgroundColor != "" {
		tview.Styles.PrimitiveBackgroundColor = util.HexStringToColor(t.BackgroundColor)
	} else {
		tview.Styles.PrimitiveBackgroundColor = tcell.ColorDefault
	}
	if t.BorderColor != "" {
		tview.Styles.BorderColor = util.HexStringToColor(t.BorderColor)
	}
	if t.NoContentColor != "" {
		activeTheme.NoContentColor = util.HexStringToColor(t.NoContentColor)
	}
	if t.LoadingColor != "" {
		activeTheme.LoadingColor = util.HexStringToColor(t.LoadingColor)
	}
	if t.LiveColor != "" {
		activeTheme.LiveColor = util.HexStringToColor(t.LiveColor)
	}
	if t.UpdateColor != "" {
		activeTheme.UpdateColor = util.HexStringToColor(t.UpdateColor)
	}
	if t.TerminalAccentColor != "" {
		activeTheme.TerminalAccentColor = util.HexStringToColor(t.TerminalAccentColor)
	}
	if t.TerminalTextColor != "" {
		activeTheme.TerminalTextColor = util.HexStringToColor(t.TerminalTextColor)
	}
	if t.InfoColor != "" {
		activeTheme.InfoColor = util.HexStringToColor(t.InfoColor)
	}
	if t.ErrorColor != "" {
		activeTheme.ErrorColor = util.HexStringToColor(t.ErrorColor)
	}
	if t.MultiCommandColor != "" {
		activeTheme.MultiCommandColor = util.HexStringToColor(t.MultiCommandColor)
	}
}
