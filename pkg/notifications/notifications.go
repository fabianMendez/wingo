package notifications

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fabianMendez/wingo/pkg/date"
	"github.com/fabianMendez/wingo/pkg/storage"
	"github.com/google/uuid"
)

const notificationsdir = "notifications"

type Setting struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
	Email       string `json:"email"`
	Confirmed   bool   `json:"confirmed"`
}

func LoadAllSettings() ([]Setting, error) {
	fentries, err := os.ReadDir(notificationsdir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	settingsMutex := new(sync.Mutex)
	settings := []Setting{}

	wg := new(sync.WaitGroup)

	for _, fentry := range fentries {
		if fentry.IsDir() {
			continue
		}

		wg.Add(1)
		go func(fentry fs.DirEntry) {
			defer wg.Done()

			fname := filepath.Join(notificationsdir, fentry.Name())
			f, err := os.Open(fname)
			if err != nil {
				return
			}

			defer f.Close()
			var setting Setting
			err = json.NewDecoder(f).Decode(&setting)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				return
			}

			settingsMutex.Lock()
			settings = append(settings, setting)
			settingsMutex.Unlock()
		}(fentry)
	}
	wg.Wait()

	return settings, nil
}

var githubStorage = storage.GithubStorage{
	Token: "ghtoken",
	Owner: "fabianMendez",
	Repo:  "wingo",
}

func SaveSetting(setting Setting) (string, error) {
	uid := uuid.New()
	b, err := json.Marshal(setting)
	if err != nil {
		return "", err
	}

	fname := filepath.Join(notificationsdir, uid.String()+".json")
	err = githubStorage.Write(fname, b, "add notification")
	if err != nil {
		return "", fmt.Errorf("could not save setting: %w", err)
	}

	return uid.String(), nil
}

func UpdateSetting(uid string, setting Setting) error {
	b, err := json.Marshal(setting)
	if err != nil {
		return err
	}

	fname := filepath.Join(notificationsdir, uid+".json")
	err = githubStorage.Write(fname, b, "update notification")
	if err != nil {
		return fmt.Errorf("could not save setting: %w", err)
	}

	return nil
}

func DeleteSetting(uid string) error {
	fname := filepath.Join(notificationsdir, uid+".json")
	err := githubStorage.Delete(fname, "delete notification")
	if err != nil {
		return fmt.Errorf("could not delete setting: %w", err)
	}

	return nil
}

func GetSetting(uid string) (Setting, error) {
	var setting Setting
	fname := filepath.Join(notificationsdir, uid+".json")

	content, err := githubStorage.Read(fname)
	if err != nil {
		return setting, fmt.Errorf("could not read setting: %w", err)
	}

	err = json.Unmarshal(content, &setting)
	if err != nil {
		return setting, fmt.Errorf("could not decode setting: %w", err)
	}

	return setting, nil
}

func FilterConfirmed(settings []Setting) []Setting {
	filtered := []Setting{}
	for _, setting := range settings {
		if setting.Confirmed {
			filtered = append(filtered, setting)
		}
	}
	return filtered
}

func FilterBetweenDates(subscriptions []Setting, start, end time.Time) []Setting {
	filtered := []Setting{}
	for _, sub := range subscriptions {
		d, err := date.Parse(sub.Date)
		if err != nil {
			continue
		}
		if (start.Before(*d) || start.Equal(*d)) && end.After(*d) {
			filtered = append(filtered, sub)
		}
	}
	return filtered
}

func GroupByRoute(subs []Setting) map[string]map[string][]Setting {
	grouped := map[string]map[string][]Setting{}

	for _, sub := range subs {
		if grouped[sub.Origin] == nil {
			grouped[sub.Origin] = map[string][]Setting{}
		}
		if grouped[sub.Destination] == nil {
			grouped[sub.Origin][sub.Destination] = []Setting{}
		}
		grouped[sub.Origin][sub.Destination] = append(grouped[sub.Origin][sub.Destination], sub)
	}

	return grouped
}
