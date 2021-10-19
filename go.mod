module github.com/SoMuchForSubtlety/f1viewer/v2

go 1.16

replace (
	github.com/99designs/keyring => github.com/SoMuchForSubtlety/keyring v0.0.0-20211019205428-6ca9d53c3df8
	github.com/rivo/tview => github.com/SoMuchForSubtlety/tview v0.0.0-20210731202536-88987c7f5054
)

require (
	github.com/99designs/keyring v1.1.6
	github.com/atotto/clipboard v0.1.4
	github.com/gdamore/tcell/v2 v2.4.0
	github.com/mattn/go-runewidth v0.0.13 // indirect
	github.com/rivo/tview v0.0.0-20210624165335-29d673af0ce2
	github.com/stretchr/testify v1.7.0
	golang.org/x/sys v0.0.0-20211019181941-9d821ace8654
	golang.org/x/text v0.3.7 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)
