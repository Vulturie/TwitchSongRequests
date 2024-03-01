package db

import "github.com/vulturie/twitchsongrequests/pkg/preferences"

type PreferenceStore interface {
	GetPreference(string) (*preferences.Preference, error)
	AddPreference(*preferences.Preference) error
	UpdatePreference(*preferences.Preference) error
	DeletePreference(string) error
}
