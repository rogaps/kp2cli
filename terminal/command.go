package terminal

import (
	"fmt"

	"github.com/peterh/liner"
)

type CommandFunc func(*Term, *Context) error

type Command struct {
	Name      string
	Help      string
	CmdFn     CommandFunc
	Completer liner.WordCompleter
}

func (cmd *Command) match(cmdStr string) bool {
	if cmdStr == cmd.Name {
		return true
	}
	return false
}

type Context struct {
	Cmd  *Command
	Args string
}

var errCmdNotAvailable = fmt.Errorf("command not available")
var noCmdAvailable = Command{CmdFn: func(t *Term, ctx *Context) error {
	return errCmdNotAvailable
}}
