package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gdamore/tcell"
	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestFullSessionSeasons(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"objects":[{"uid":"race_12345","year":2019,"has_content":true,"name":"2019 Season"},{"uid":"race_6789","year":2020,"has_content":false,"name":"2020 Season"}]}`)
	}))
	defer server.Close()
	endpoint = server.URL + "/"

	expectedUnselected := `
Full Race Weekends  ┌──────────────────┐
                    │                  │
                    │                  │
                    │                  │
                    └──────────────────┘`
	expectedSelected := `
Full Race Weekends  ┌──────────────────┐
└──2019 Season      │                  │
                    │                  │
                    │                  │
                    └──────────────────┘`

	screen, session := newTestApp(40, 5)
	go session.app.Run()

	node := session.getFullSessionsNode()
	assert.Equal(t, node.GetText(), "Full Race Weekends")
	assert.Empty(t, node.GetChildren())

	session.tree.GetRoot().AddChild(node)
	session.tree.SetCurrentNode(node)

	time.Sleep(time.Millisecond * 10)
	assert.Equal(t, expectedUnselected, toTextScreen(screen))

	handler := session.tree.InputHandler()
	handler(tcell.NewEventKey(tcell.KeyEnter, ' ', tcell.ModNone), func(p tview.Primitive) {})
	time.Sleep(time.Millisecond * 500)
	assert.Equal(t, expectedSelected, toTextScreen(screen))
}
