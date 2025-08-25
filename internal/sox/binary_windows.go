package sox

import (
	"embed"
	"os"
)

//go:embed sox-14.4.2-win32/*
var embedfs embed.FS

var foldername = "sox-14.4.2-win32"
var binaryname = "soxstatic/sox.exe"

func getPath() (pathname string, err error) {
	// check if soxstatic folder exists
	_, err = os.ReadDir("soxstatic")
	if err == nil {
		return binaryname, nil
	}
	err = os.Mkdir("soxstatic", 0755)
	if err != nil {
		return
	}

	// list files in the embed.FS
	fis, err := embedfs.ReadDir(foldername)
	if err != nil {
		return
	}
	for _, fi := range fis {
		// copy files from embed.FS to soxstatic folder
		var b []byte
		b, err = embedfs.ReadFile(foldername + "/" + fi.Name())
		if err != nil {
			return
		}
		err = os.WriteFile("soxstatic/"+fi.Name(), b, 0755)
		if err != nil {
			return
		}
	}
	return binaryname, nil
}
