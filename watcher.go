package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

//
var watcher *fsnotify.Watcher

// main
func watch(dir string, out *string) {
	go func() {
		// creates a new file watcher
		watcher, _ = fsnotify.NewWatcher()
		defer watcher.Close()

		// starting at the root of the project, walk each file/directory searching for
		// directories
		if err := filepath.Walk(dir, watchDir); err != nil {
			fmt.Println("ERROR", err)
		}

		done := make(chan bool)

		for {
			select {
			case <-watcher.Events:
				{
					json := getCombinedJson()
					if "" != json {
						*out = json
					}
				}
			case <-watcher.Errors:
				{
					json := getCombinedJson()
					if "" != json {
						*out = json
					}
				}
			}
		}
		<-done
	}()
}

// watchDir gets run as a walk func, searching for directories to add watchers to
func watchDir(path string, fi os.FileInfo, err error) error {

	// since fsnotify can watch all the files in a directory, watchers only need
	// to be added to each nested directory
	if fi.Mode().IsDir() {
		return watcher.Add(path)
	}

	return nil
}
