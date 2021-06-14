package fshandlers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
)

type UploadHandler struct {
	ServerAddr string
	ServeDir   string
}

func (h *UploadHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	file, header, err := r.FormFile("file")
	if err != nil {
		log.Print("cannot read file")
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		http.Error(w, "Unable to read file", http.StatusBadRequest)
		return
	}

	filePath := h.ServeDir + "/" + header.Filename

	err = ioutil.WriteFile(filePath, data, 0777)
	if err != nil {
		log.Println(err)
		http.Error(w, "Unable to save file", http.StatusInternalServerError)
		return
	}

	fileLink := "http://" + h.ServerAddr + "/" + header.Filename
	fmt.Fprintln(w, fileLink)
}

type FileData struct {
	Name string
	Ext  string
	Size int64
}

type ListHandler struct {
	ServeDir string
}

func (h *ListHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodGet {
		http.Error(w, "GET-method expected", http.StatusBadRequest)
		return
	}

	files, err := ioutil.ReadDir(h.ServeDir)
	if err != nil {
		log.Printf("cannot read file from directory %s", h.ServeDir)
		http.Error(w, "Unable to read directory", http.StatusBadRequest)
		return
	}

	ext := r.FormValue("ext")
	var filesData []FileData
	for _, someFile := range files {
		if !someFile.IsDir() {
			fileAttr := FileData{
				Name: someFile.Name(),
				Ext:  filepath.Ext(someFile.Name()),
				Size: someFile.Size(),
			}
			if ext == "" || fileAttr.Ext == ext {
				filesData = append(filesData, fileAttr)
			}
		}
	}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(filesData)
	if err != nil {
		log.Printf("cannot convert data into json %v", err)
		http.Error(w, "Unable to read directory", http.StatusBadRequest)
		return
	}
}
