package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/hspak/opvault"
)

type entry struct {
	username string
	password string
}

// Supports both login and password types
func NewItems(items []*opvault.Item) (map[string]*entry, error) {
	entrymap := make(map[string]*entry)
	for _, item := range items {
		if !item.Trashed() {
			l := &entry{
				username: "N/A",
				password: "N/A",
			}
			detail, err := item.Detail()
			if err != nil {
				return nil, err
			}
			if detail.Password() == "" {
				for _, field := range detail.Fields() {
					if field.Name() == "username" {
						l.username = field.Value()
					} else if field.Name() == "password" {
						l.password = field.Value()
					}
				}
			} else {
				l.password = detail.Password()
			}
			entrymap[strings.ToLower(item.Title())] = l
		}
	}
	return entrymap, nil
}

func queryItems(profile *opvault.Profile) ([]string, error) {
	var titles []string
	items, err := profile.Items()
	if err != nil {
		return titles, err
	}
	for _, item := range items {
		if !item.Trashed() {
			titles = append(titles, item.Title())
		}
	}
	return titles, nil
}

func checkProfiles(vault *opvault.Vault, profileName string) error {
	names, err := vault.ProfileNames()
	if err != nil {
		return err
	}
	if len(names) < 1 {
		return errors.New("error: no profiles found")
	}
	validProfileName := false
	for _, name := range names {
		if name == profileName {
			validProfileName = true
			break
		}
	}
	if !validProfileName {
		msg := fmt.Sprintf("error: no profile named \"%s\" found", profileName)
		return errors.New(msg)
	}
	return nil
}
