package util

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/pkg/errors"
)

type Prefs struct {
	Instances   map[string]PrefsInstance `json:"instances,omitempty"`
	Users       map[string]PrefsUser     `json:"users,omitempty"`
	DefaultUser string                   `json:"default_user,omitempty"`
}

type PrefsInstance struct {
	ClientID string `json:"client_id"`
}

type PrefsUser struct {
	Instance string `json:"instance"`
}

// prefsDir is the path to the directory containing all femoji preference files.
var prefsDir string

// prefsPath is the path to the file within that directory that stores all of our prefs.
var prefsPath string

func init() {
	prefsDir = filepath.Join(xdg.ConfigHome, "dev.solistar.femoji")
	prefsPath = filepath.Join(prefsDir, "prefs.json")
}

// LoadPrefs returns preferences from disk or an empty prefs object if there are none stored or accessible.
func LoadPrefs() (*Prefs, error) {
	f, err := os.Open(prefsPath)
	if err != nil {
		return &Prefs{
			Instances: map[string]PrefsInstance{},
			Users:     map[string]PrefsUser{},
		}, nil
	}
	defer func() { _ = f.Close() }()

	var prefs Prefs
	err = json.NewDecoder(f).Decode(&prefs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	if prefs.Instances == nil {
		prefs.Instances = map[string]PrefsInstance{}
	}
	if prefs.Users == nil {
		prefs.Users = map[string]PrefsUser{}
	}

	return &prefs, nil
}

// SavePrefs creates on-disk preferences or overwrites existing ones.
func SavePrefs(prefs *Prefs) error {
	err := os.MkdirAll(prefsDir, 0o755)
	if err != nil {
		return errors.WithStack(err)
	}

	f, err := os.Create(prefsPath)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() { _ = f.Close() }()

	encoder := json.NewEncoder(f)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(prefs)
	if err != nil {
		return errors.WithStack(err)
	}

	return nil
}

var PrefNotFound = errors.New("preference value not found")

func getPrefValue(get func(prefs *Prefs) (string, bool)) (string, error) {
	prefs, err := LoadPrefs()
	if err != nil {
		return "", err
	}

	value, exists := get(prefs)
	if !exists {
		return "", errors.WithStack(PrefNotFound)
	}

	return value, nil
}

func setPrefValue(set func(prefs *Prefs)) error {
	prefs, err := LoadPrefs()
	if err != nil {
		return err
	}

	set(prefs)

	err = SavePrefs(prefs)
	if err != nil {
		return err
	}

	return nil
}

func GetDefaultUser() (string, error) {
	return getPrefValue(func(prefs *Prefs) (string, bool) {
		if prefs.DefaultUser == "" {
			return "", false
		}
		return prefs.DefaultUser, true
	})
}

func SetDefaultUser(user string) error {
	return setPrefValue(func(prefs *Prefs) {
		prefs.DefaultUser = user
	})
}

func GetInstanceClientID(instance string) (string, error) {
	return getPrefValue(func(prefs *Prefs) (string, bool) {
		prefsInstance, exists := prefs.Instances[instance]
		if !exists {
			return "", false
		}
		return prefsInstance.ClientID, true
	})
}

func SetInstanceClientID(instance string, clientID string) error {
	return setPrefValue(func(prefs *Prefs) {
		prefsInstance := prefs.Instances[instance]
		prefsInstance.ClientID = clientID
		prefs.Instances[instance] = prefsInstance
	})
}

func GetUserInstance(user string) (string, error) {
	return getPrefValue(func(prefs *Prefs) (string, bool) {
		prefsUser, exists := prefs.Users[user]
		if !exists {
			return "", false
		}
		return prefsUser.Instance, true
	})
}

func SetUserInstance(user string, instance string) error {
	return setPrefValue(func(prefs *Prefs) {
		prefsUser := prefs.Users[user]
		prefsUser.Instance = instance
		prefs.Users[user] = prefsUser
	})
}
