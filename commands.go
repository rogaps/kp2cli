package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
	"text/tabwriter"
	"unsafe"

	"github.com/atotto/clipboard"
	"github.com/google/shlex"
	"github.com/mitchellh/go-homedir"
	"github.com/rogaps/kp2cli/terminal"
	"github.com/tobischo/gokeepasslib/v3"
)

var cmds = []terminal.Command{
	{Name: "cd", Help: "Change directory (path to a group)", CmdFn: cdCmd, Completer: groupCompleter},
	{Name: "close", Help: "Close the opened database", CmdFn: closeCmd, Completer: filenameCompleter},
	{Name: "exit", Help: "Exit this program", CmdFn: exitCmd},
	{Name: "find", Help: "Find entries", CmdFn: findCmd},
	{Name: "help", Help: "Print help", CmdFn: helpCmd},
	{Name: "ls", Help: "List items in the pwd or specified paths", CmdFn: lsCmd, Completer: groupCompleter},
	{Name: "open", Help: "Open a Keepass database", CmdFn: openCmd, Completer: filenameCompleter},
	{Name: "xp", Help: "Copy password to clipboard", CmdFn: xpCmd, Completer: entryCompleter},
	{Name: "xu", Help: "Copy username to clipboard", CmdFn: xuCmd, Completer: entryCompleter},
	{Name: "xx", Help: "Clear the clipboard", CmdFn: xxCmd},
}

func helpCmd(t *terminal.Term, ctx *terminal.Context) error {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
	for _, cmd := range t.Cmds {
		fmt.Fprintf(w, "%s\t%s\t\n", cmd.Name, cmd.Help)
	}
	return w.Flush()
}

func exitCmd(t *terminal.Term, ctx *terminal.Context) error {
	t.Stop()
	return errors.New("exit")
}

func openCmd(t *terminal.Term, ctx *terminal.Context) error {
	var filePath string
	var keyPath string
	var err error
	var password string
	args, err := shlex.Split(ctx.Args)
	argsLength := len(args)

	if argsLength == 0 {
		return fmt.Errorf("args == 0")
	}
	if 0 < argsLength && argsLength <= 1 {
		filePath = args[0]
	} else {
		filePath = args[0]
		keyPath = args[1]
	}

	expandedFilePath, err := homedir.Expand(filePath)
	if err != nil {
		return err
	}
	file, err := os.Open(expandedFilePath)
	if err != nil {
		return err
	}
	db := gokeepasslib.NewDatabase()
	if keyPath != "" {
		expandedKeyPath, err := homedir.Expand(keyPath)
		if err != nil {
			return err
		}
		db.Credentials, err = gokeepasslib.NewKeyCredentials(expandedKeyPath)
		if err != nil {
			return err
		}
	} else {
		password, err = t.Line.PasswordPrompt("Enter password: ")
		if err != nil {
			return err
		}
		db.Credentials = gokeepasslib.NewPasswordCredentials(password)
	}
	err = gokeepasslib.NewDecoder(file).Decode(db)
	if err != nil {
		return err
	}
	if err := db.UnlockProtectedEntries(); err != nil {
		return err
	}
	setDb(t, db)

	return nil
}

func closeCmd(t *terminal.Term, ctx *terminal.Context) error {
	setDb(t, newDatabase())
	return nil
}

type winsize struct {
	Row    uint16
	Col    uint16
	Xpixel uint16
	Ypixel uint16
}

func getColumns() (int, error) {
	ws := &winsize{}
	retCode, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TIOCGWINSZ),
		uintptr(unsafe.Pointer(ws)))

	if int(retCode) == -1 {
		return 0, errno
	}
	return int(ws.Col), nil
}

func calculateColumns(screenWidth int, items []string) (numColumns, numRows, maxWidth int) {
	for _, item := range items {
		if len(item) >= screenWidth {
			return 1, len(items), screenWidth - 1
		}
		if len(item) >= maxWidth {
			maxWidth = len(item) + 1
		}
	}

	numColumns = screenWidth / maxWidth
	numRows = len(items) / numColumns
	if len(items)%numColumns > 0 {
		numRows++
	}

	if len(items) <= numColumns {
		maxWidth = 0
	}

	return
}

func lsCmd(t *terminal.Term, ctx *terminal.Context) error {
	var items []string
	target := cwd
	if len(ctx.Args) > 0 {
		var err error
		target, err = travel(target, ctx.Args)
		if err != nil {
			return err
		}
	}

	group := target.Group()
	if len(group.Groups) > 0 {
		for _, subGroup := range group.Groups {
			items = append(items, subGroup.Name+"/")
		}
	}
	if len(group.Entries) > 0 {
		for _, entry := range group.Entries {
			items = append(items, entry.GetTitle())
		}
	}
	if len(items) > 0 {
		if screenWidth, err := getColumns(); err != nil {
			for _, item := range items {
				fmt.Println(item)
			}
		} else {
			cols, rows, maxWidth := calculateColumns(screenWidth, items)
			for i := 0; i < rows; i++ {
				for j := 0; j < cols*rows; j += rows {
					if i+j < len(items) {
						if maxWidth > 0 {
							fmt.Printf("%-*.[1]*s", maxWidth, items[i+j])
						} else {
							fmt.Printf("%v ", items[i+j])
						}
					}
				}
				fmt.Println("")
			}
		}
	}
	return nil
}

func cdCmd(t *terminal.Term, ctx *terminal.Context) error {
	if len(ctx.Args) > 0 {
		wg, _ := shlex.Split(ctx.Args)
		b, err := travel(cwd, wg[0])
		if err != nil {
			return err
		}
		setCwd(t, b)
	}
	return nil
}

func xuCmd(t *terminal.Term, ctx *terminal.Context) error {
	var selectedEntry *gokeepasslib.Entry
	args, _ := shlex.Split(ctx.Args)
	if len(args) > 0 {
		if len(args) > 0 {
			groupPath, entryTitle := groupSplit(args[0])
			wg, err := travel(cwd, groupPath)
			if err != nil {
				return err
			}
			group := wg.Group()
			for _, entry := range group.Entries {
				title := entry.GetTitle()
				if strings.TrimSpace(entryTitle) == title {
					selectedEntry = &entry
					break
				}
			}

			if selectedEntry != nil {
				if err := clipboard.WriteAll(getEntryContent(*selectedEntry, "Username")); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func xpCmd(t *terminal.Term, ctx *terminal.Context) error {
	var selectedEntry *gokeepasslib.Entry
	args, _ := shlex.Split(ctx.Args)
	if len(args) > 0 {
		if len(args) > 0 {
			groupPath, entryTitle := groupSplit(args[0])
			wg, err := travel(cwd, groupPath)
			if err != nil {
				return err
			}
			group := wg.Group()
			for _, entry := range group.Entries {
				title := entry.GetTitle()
				if strings.TrimSpace(entryTitle) == title {
					selectedEntry = &entry
					break
				}
			}

			if selectedEntry != nil {
				if err := clipboard.WriteAll(selectedEntry.GetPassword()); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func xxCmd(t *terminal.Term, ctx *terminal.Context) error {
	return clipboard.WriteAll("")
}

func findCmd(t *terminal.Term, ctx *terminal.Context) error {
	var results []string
	root := newRootGroup(db)
	args, _ := shlex.Split(ctx.Args)
	if len(args) > 0 {
		results = findEntry(root, args[0])
	}
	for _, result := range results {
		fmt.Println(result)
	}
	return nil
}
