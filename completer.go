package main

import (
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"

	"github.com/mitchellh/go-homedir"
	"github.com/rogaps/kp2cli/terminal"
)

func filenameCompleter(line string, pos int) (head string, completions []string, tail string) {
	var err error
	var quoteFound bool
	var word string

	words := terminal.LineSplit(line[:pos])
	wordslen := len(words)
	head = strings.Join(words[:wordslen-1], "")
	word = words[wordslen-1]
	tail = line[pos:]
	wordPos := len(head)

	oq, oqsize := utf8.DecodeRuneInString(word)
	if strings.IndexRune(terminal.QuoteChars, oq) >= 0 {
		quoteFound = true
		head = line[:wordPos+oqsize]
		word = line[wordPos+oqsize : pos]
		cq, cqsize := utf8.DecodeLastRuneInString(word)
		if cq == oq {
			word = word[:len(word)-cqsize]
		}
	}

	dir, match := filepath.Split(terminal.UnescapeString(word))
	if match == "." {
		completions = append(completions, "./", "../")
	} else if match == ".." {
		completions = append(completions, "../")
	}

	var f *os.File
	if dir == "" {
		f, err = os.Open(".")
	} else {
		expandedDir, err := homedir.Expand(dir)
		if err != nil {
			return
		}
		f, err = os.Open(expandedDir)
	}
	if err != nil {
		return
	}
	entries, err := f.Readdir(0)
	for _, entry := range entries {
		if strings.HasPrefix(entry.Name(), match) {
			var sep string
			if entry.IsDir() {
				sep = string(os.PathSeparator)
			} else {
				if len(tail) == 0 && !quoteFound {
					sep = " "
				}
			}
			path := entry.Name() + sep
			completions = append(completions, path)
		}
	}
	head += terminal.EscapeString(dir, quoteFound)

	for i, complet := range completions {
		completions[i] = terminal.EscapeString(complet, quoteFound)
	}

	return
}

func groupSplit(path string) (group string, entry string) {
	for i := len(path); i > 0; {
		r, size := utf8.DecodeLastRuneInString(path[0:i])
		if r == '/' {
			return path[:i], path[i:]
		}
		i -= size
	}
	return "", path
}

func groupCompleter(line string, pos int) (head string, completions []string, tail string) {
	var err error
	var quoteFound bool
	var word string

	words := terminal.LineSplit(line[:pos])
	wordslen := len(words)
	head = strings.Join(words[:wordslen-1], "")
	word = words[wordslen-1]
	tail = line[pos:]
	wordPos := len(head)

	oq, oqsize := utf8.DecodeRuneInString(word)
	if strings.IndexRune(terminal.QuoteChars, oq) >= 0 {
		quoteFound = true
		head = line[:wordPos+oqsize]
		word = line[wordPos+oqsize : pos]
	}

	groupPath, match := groupSplit(terminal.UnescapeString(word))
	if match == "." {
		completions = append(completions, "./", "../")
	} else if match == ".." {
		completions = append(completions, "../")
	}

	group, err := travel(cwd, groupPath)
	if err != nil {
		return
	}
	for _, subGroup := range group.Group().Groups {
		name := subGroup.Name
		if strings.HasPrefix(name, match) {
			completions = append(completions, name+"/")
		}
	}

	head += terminal.EscapeString(groupPath, quoteFound)

	for i, complet := range completions {
		completions[i] = terminal.EscapeString(complet, quoteFound)
	}
	return
}

func entryCompleter(line string, pos int) (head string, completions []string, tail string) {
	var err error
	var quoteFound bool
	var word string

	words := terminal.LineSplit(line[:pos])
	wordslen := len(words)
	head = strings.Join(words[:wordslen-1], "")
	word = words[wordslen-1]
	tail = line[pos:]
	wordPos := len(head)

	oq, oqsize := utf8.DecodeRuneInString(word)
	if strings.IndexRune(terminal.QuoteChars, oq) >= 0 {
		quoteFound = true
		head = line[:wordPos+oqsize]
		word = line[wordPos+oqsize : pos]
	}

	groupPath, match := groupSplit(terminal.UnescapeString(word))
	if match == "." {
		completions = append(completions, "./", "../")
	} else if match == ".." {
		completions = append(completions, "../")
	}

	group, err := travel(cwd, groupPath)
	if err != nil {
		return
	}
	for _, subGroup := range group.Group().Groups {
		name := subGroup.Name
		if strings.HasPrefix(name, match) {
			completions = append(completions, name+"/")
		}
	}

	for _, entry := range group.Group().Entries {
		title := entry.GetTitle()
		if strings.HasPrefix(title, match) {
			completions = append(completions, title+" ")
		}
	}

	head += terminal.EscapeString(groupPath, quoteFound)

	for i, complet := range completions {
		completions[i] = terminal.EscapeString(complet, quoteFound)
	}
	return
}
