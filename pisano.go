// Package main pisano.go
/*
pisano — the designs that fall out of reducing an integer sequence modulo m
*/
package main

import (
	"github.com/0magnet/calvin/clihelp"
	"github.com/0magnet/pisano/cmd/pisano/commands"
)

func init() {
	clihelp.Init(commands.RootCmd, "pisano", true)
}

func main() {
	commands.Execute()
}
