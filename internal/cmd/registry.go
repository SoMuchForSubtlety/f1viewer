// +build windows

package cmd

import (
	"runtime"

	"golang.org/x/sys/windows/registry"
)

func checkRegistry(c Command) (Command, bool) {
	regPath := c.registry
	if runtime.GOARCH == "386" {
		regPath = c.registry32
	}

	if regPath == "" {
		return c, false
	}

	result, err := registry.OpenKey(registry.LOCAL_MACHINE, regPath, registry.QUERY_VALUE)
	if err != nil {
		return c, false
	}

	path, _, err := result.GetStringValue("InstallDir")
	if err != nil {
		return c, false
	}
	c.Command[0] = path + "\\" + c.Command[0]

	return c, true
}
