package main

import (
	"testing"

	"github.com/rivo/tview"
	"github.com/stretchr/testify/assert"
)

func TestMetadata(t *testing.T) {
	node := tview.NewTreeNode("Testing").SetReference(NodeMetadata{nodeType: MiscNode})
	metadata, err := getMetadata(node)
	assert.NoError(t, err)
	assert.Equal(t, MiscNode, metadata.nodeType)

	_, err = getMetadata(nil)
	assert.EqualError(t, err, "node is nil")

	_, err = getMetadata(tview.NewTreeNode("Testing"))
	assert.EqualError(t, err, "Node has reference of unexpected type <nil>")

	_, err = getMetadata(tview.NewTreeNode("Testing").SetReference(123))
	assert.EqualError(t, err, "Node has reference of unexpected type int")
}
