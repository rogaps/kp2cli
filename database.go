package main

import (
	"regexp"
	"strings"

	"github.com/tobischo/gokeepasslib/v3"
	w "github.com/tobischo/gokeepasslib/v3/wrappers"
)

func mkValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{Key: key, Value: gokeepasslib.V{Content: value}}
}

func mkProtectedValue(key string, value string) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key:   key,
		Value: gokeepasslib.V{Content: value, Protected: w.NewBoolWrapper(true)},
	}
}
func newDatabase() *gokeepasslib.Database {
	rootGroup := gokeepasslib.NewGroup()
	rootGroup.Name = "Database"

	emailGroup := gokeepasslib.NewGroup()
	emailGroup.Name = "Email"

	internetGroup := gokeepasslib.NewGroup()
	internetGroup.Name = "Internet"

	rootGroup.Groups = append(rootGroup.Groups, emailGroup, internetGroup)

	db := gokeepasslib.NewDatabase()
	db.Content.Root = &gokeepasslib.RootData{
		Groups: []gokeepasslib.Group{rootGroup},
	}
	return db
}

func getEntryContent(entry gokeepasslib.Entry, key string) string {
	for i := range entry.Values {
		if strings.EqualFold(entry.Values[i].Key, key) {
			return entry.Values[i].Value.Content
		}
	}
	return ""
}

func findEntry(root workingGroup, search string) []string {
	var results []string
	group := root.Group()
	regex := regexp.MustCompile("(?i)" + search)
	for _, entry := range group.Entries {
		for _, value := range entry.Values {
			if regex.MatchString(value.Value.Content) {
				results = append(results, strings.Join([]string{root.String(), entry.GetTitle()}, "/"))
				break
			}
		}
	}
	for _, subGroup := range group.Groups {
		subWorkingGroup, err := travel(root, subGroup.Name)
		if err != nil {
			panic(err)
		}
		results = append(results, findEntry(subWorkingGroup, search)...)
	}
	return results
}
