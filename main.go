package main

import (
	"fmt"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/rogaps/kp2cli/terminal"
	"github.com/tobischo/gokeepasslib/v3"
)

var (
	db  *gokeepasslib.Database
	cwd workingGroup
)

func setCwd(t *terminal.Term, c workingGroup) {
	cwd = c
	t.SetPrompt(fmt.Sprintf("%s:%s> ", "kp2cli", cwd.String()))
}

func setDb(t *terminal.Term, d *gokeepasslib.Database) {
	db = d
	setCwd(t, newRootGroup(db))
}

func main() {
	var err error
	var historyFilePath string
	var exitCode int

	t := terminal.NewTerm(&terminal.TermConfig{})
	t.SetCommands(cmds...)
	historyFilePath, err = homedir.Expand("~/.kp2cli_history")
	if err != nil {
		historyFilePath = ".kp2cli_history"
	}
	t.SetHistoryFile(historyFilePath)
	setDb(t, newDatabase())
	setCwd(t, newRootGroup(db))

	fmt.Println("kp2cli, Keepass 2 Interactive Shell")
	fmt.Println("Type \"help\" for help.")
	fmt.Println()

	t.Run()
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(exitCode)
	}
}
