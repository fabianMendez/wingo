package main

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fabianMendez/wingo"
	"github.com/fabianMendez/wingo/pkg/date"
)

func saveToFile(filename string, v interface{}) error {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(filename, b, os.ModePerm)
}

func loadFromFile(filename string, v interface{}) error {
	f, err := os.Open(filename)
	if err != nil {
		return err
	}

	return json.NewDecoder(f).Decode(v)
}

func loadSavedFlights(dir string, savedRoutes []wingo.Route, startDate, stopDate time.Time) (flightsMap, error) {
	wg := new(sync.WaitGroup)
	flightsMutex := new(sync.Mutex)
	flights := flightsMap{}

	for _, origin := range savedRoutes {
		for _, destination := range origin.Routes {
			dirname := filepath.Join(dir, outdir, origin.Code, destination.Code)
			direntries, err := os.ReadDir(dirname)
			if err != nil {
				if os.IsNotExist(err) {
					continue
				}
				return nil, err
			}

			wg.Add(1)
			go func(origin, destination string, direntries []fs.DirEntry) {
				defer wg.Done()

				for _, dentry := range direntries {
					if !dentry.IsDir() {
						continue
					}

					flightsMutex.Lock()
					if flights[origin] == nil {
						flights[origin] = map[string]map[string][]vueloArchivado{}
					}

					if flights[origin][destination] == nil {
						flights[origin][destination] = map[string][]vueloArchivado{}
					}
					flightsMutex.Unlock()

					datepath := filepath.Join(dirname, dentry.Name())
					fentries, err := os.ReadDir(datepath)
					if err != nil {
						// return nil, err
						continue
					}

					for _, fentry := range fentries {
						var varchivado vueloArchivado

						flightpath := filepath.Join(datepath, fentry.Name())
						f, err := os.Open(flightpath)
						if err != nil {
							// return nil, err
							continue
						}
						defer f.Close()

						err = json.NewDecoder(f).Decode(&varchivado)
						if err != nil {
							// return nil, err
							continue
						}

						datestr := dentry.Name()
						date, err := date.Parse(datestr)
						if err != nil {
							continue
						}

						if date.Before(startDate) || date.After(stopDate) {
							continue
						}

						flightsMutex.Lock()
						flights[origin][destination][datestr] = append(
							flights[origin][destination][datestr],
							varchivado,
						)
						flightsMutex.Unlock()
					}
				}
			}(origin.Code, destination.Code, direntries)
		}
	}
	wg.Wait()

	return flights, nil
}
