package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
)

type notificationSetting struct {
	Origin      string `json:"origin"`
	Destination string `json:"destination"`
	Date        string `json:"date"`
	Email       string `json:"email"`
}

func loadNotificationsSettings() ([]notificationSetting, error) {
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
