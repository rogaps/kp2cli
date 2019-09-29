package terminal

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync/atomic"

	"github.com/peterh/liner"
)

type Term struct {
	Line           *liner.State
	prompt         string
	Cmds           []Command
	historyFile    string
	running        uint32
	interruptCount uint32
}

type TermConfig struct {
	WordCompleter liner.WordCompleter
}

func NewTerm(config *TermConfig) *Term {
	t := &Term{
		Line:    liner.NewLiner(),
		prompt:  "kp2cli> ",
		running: 1,
	}
	if config.WordCompleter != nil {
		t.Line.SetWordCompleter(config.WordCompleter)
	} else {
		t.Line.SetWordCompleter(t.defaultWordCompleter)
	}
	t.Line.SetTabCompletionStyle(liner.TabPrints)
	t.Line.SetCtrlCAborts(true)
	return t
}

func (t *Term) SetWordCompleter(f liner.WordCompleter) {
	t.Line.SetWordCompleter(f)
}

func (t *Term) SetCompleter(f liner.Completer) {
	t.Line.SetCompleter(f)
}

func (t *Term) defaultWordCompleter(line string, pos int) (head string, completions []string, tail string) {
	words := LineSplit(line)
	if len(words) > 1 {
		cmdName, args := words[0], words[1:]
		cmd := t.findCmd(strings.TrimSpace(cmdName))
		if cmd.Completer != nil {
			head, completions, tail = cmd.Completer(strings.Join(args, ""), pos-len(cmdName))
			head = cmdName + head
			return
		}
	}
	for _, cmd := range t.Cmds {
		if strings.HasPrefix(cmd.Name, strings.ToLower(line)) {
			completions = append(completions, cmd.Name+" ")
		}
	}
	return
}

func (t *Term) SetHistoryFile(path string) {
	t.historyFile = path
}

func (t *Term) SetCommands(cmds ...Command) {
	t.Cmds = cmds
}

func (t *Term) SetPrompt(prompt string) {
	t.prompt = prompt
}

func (t *Term) Stop() {
	atomic.CompareAndSwapUint32(&t.running, 1, 0)
}

func (t *Term) Close() {
	t.Line.Close()
}

func (t *Term) Run() {
	defer t.Close()

	f, err := os.Open(t.historyFile)
	if err != nil {
		f, err = os.Create(t.historyFile)
		if err != nil {
			fmt.Printf("Unable to open history file: %v. History will not be saved for this session.", err)
		}
	}

	t.Line.ReadHistory(f)
	f.Close()

	for atomic.LoadUint32(&t.running) == 1 {
		line, err := t.promptForInput()

		switch err {
		case nil:
		case io.EOF:
			fmt.Println("exit")
			t.handleEOF(line, err)
		case liner.ErrPromptAborted:
			t.interruptCount++
			t.handleInterrupt(line, err)
		default:
			t.handleError(line, err)
		}

		if err := t.callCmd(line); err != nil {
			t.handleCmdError(err)
		}
	}
}

func (t *Term) handleEOF(line string, err error) {
	fmt.Println("exit")
	t.handleExit()
	t.Stop()
}

func (t *Term) handleInterrupt(line string, err error) {
	if t.interruptCount >= 2 {
		fmt.Println("interrupted")
		t.handleExit()
		t.Stop()
	} else {
		fmt.Println("Press ctrl-c once more to exit")
	}
}

func (t *Term) handleError(line string, err error) {
	fmt.Println("prompt for input failed")
	t.handleExit()
	os.Exit(1)
}

func (t *Term) handleCmdError(err error) {
	fmt.Fprintln(os.Stderr, err.Error())
}

func (t *Term) promptForInput() (string, error) {
	l, err := t.Line.Prompt(t.prompt)
	if err != nil {
		return l, err
	}
	l = strings.TrimSuffix(l, "\n")
	if l != "" {
		t.Line.AppendHistory(l)
	}
	return l, nil
}

func (t *Term) handleExit() {
	if f, err := os.OpenFile(t.historyFile, os.O_RDWR, 0666); err == nil {
		_, err = t.Line.WriteHistory(f)
		if err != nil {
			fmt.Println("readline history error:", err)
		}
		f.Close()
	}
}

func (t *Term) findCmd(line string) *Command {
	for _, cmd := range t.Cmds {
		if cmd.match(line) {
			return &cmd
		}
	}
	return &noCmdAvailable
}

func (t *Term) callCmd(line string) error {

	words := LineSplit(line)
	if len(words) > 0 {
		cmdName := strings.TrimSpace(words[0])
		if len(cmdName) > 0 {
			cmd := t.findCmd(cmdName)
			ctx := &Context{
				Cmd:  cmd,
				Args: strings.Join(words[1:], ""),
			}

			return cmd.CmdFn(t, ctx)
		}
	}
	return nil
}
