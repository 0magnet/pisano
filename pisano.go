// Package main pisano.go
/*
pisano — the designs that fall out of reducing an integer sequence modulo m
*/
package main

import (
	"github.com/0magnet/pisano/cmd/pisano/commands"
	"github.com/0magnet/pisano/pkg/flags"
)

func init() {
	flags.InitFlags(commands.RootCmd, true)
}

func main() {
	commands.Execute()
}
