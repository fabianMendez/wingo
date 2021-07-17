package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/google/uuid"
)

const notificationsdir = "notifications"

type notificationSetting struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
	Email       string `json:"email"`
}

func loadNotificationSettings() ([]notificationSetting, error) {
	fentries, err := os.ReadDir(notificationsdir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	settingsMutex := new(sync.Mutex)
	settings := []notificationSetting{}

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
			var setting notificationSetting
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

var githubStorage = githubSessionStorage{
	token: "ghtoken",
	owner: "fabianMendez",
	repo:  "wingo",
}

func saveNotificationSetting(setting notificationSetting) error {
	uid := uuid.New()
	b, err := json.Marshal(setting)
	if err != nil {
		return err
	}

	fname := filepath.Join(notificationsdir, uid.String()+".json")
	return githubStorage.Write(fname, b, "add notification")
}
