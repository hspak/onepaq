package main

import (
	"errors"
	"fmt"

	"github.com/hspak/opvault"
)

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
