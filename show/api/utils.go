package main

import (
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strings"
)

func checkErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func closeFile(f *os.File) {
	err := f.Close()
	checkErr(err)
}

func getPath(path string) string {
	if filepath.IsAbs(path) {
		return path
	} else if strings.HasPrefix(path, "~") {
		currentUser, err := user.Current()
		checkErr(err)
		return strings.Replace(path, "~", currentUser.HomeDir, 1)
	} else {
		execPath, err := os.Executable()
		checkErr(err)
		return filepath.Join(filepath.Dir(execPath), path)
	}
}
