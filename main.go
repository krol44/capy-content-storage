package main

import (
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	log "github.com/sirupsen/logrus"
	"golang.org/x/exp/slices"
	"io"
	_ "io/ioutil"
	"math/rand"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var rootDir = "/files"
var rootDirRemoved = "/files-removed"

func main() {
	setup()
	limitUploadMB := os.Getenv("LIMIT_UPLOAD_MB")

	log.Info("Service is running...")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var jsonResult []byte
		lum, _ := strconv.Atoi(limitUploadMB)

		_, isToken := r.Header["Token"]
		if isToken == false || r.Header["Token"][0] != os.Getenv("TOKEN") {
			jsonResult, _ = json.Marshal(Result{Status: false, Message: "no auth token", Error: ""})
			w.WriteHeader(http.StatusBadRequest)
			result(w, jsonResult)
			return
		}

		storageMask, isStorage := r.Header["Storage"]
		if isStorage == false || r.Header["Storage"][0] == "" {
			jsonResult, _ = json.Marshal(Result{Status: false, Message: "no storage mask", Error: ""})
			w.WriteHeader(http.StatusBadRequest)
			result(w, jsonResult)
			return
		}

		if strings.Contains(r.URL.Path, "/files") {
			files, err := getFiles(r)
			if err != nil {
				jsonResult, _ = json.Marshal(Result{Status: false, Message: "error retrieving the list of files",
					Error: err.Error()})
				result(w, jsonResult)
				return
			}
			result(w, files)
			return
		}

		if strings.Contains(r.URL.Path, "/remove") {
			err := remove(r)
			if err != nil {
				jsonResult, _ = json.Marshal(Result{Status: false, Message: "error removing",
					Error: err.Error()})
				result(w, jsonResult)
				return
			}
			jsonResult, _ = json.Marshal(Result{Status: true})
			result(w, jsonResult)
			return
		}

		if strings.Contains(r.URL.Path, "/upload") {
			file, handler, err := r.FormFile("file")
			if err != nil {
				jsonResult, _ = json.Marshal(Result{Status: false, Message: "error retrieving the file",
					Error: err.Error()})
				result(w, jsonResult)
				return
			}
			if handler.Size > int64(lum<<20) {
				jsonResult, _ = json.Marshal(Result{Status: false, Message: "max file size - " + limitUploadMB + "MB",
					Error: "request body too large"})
				result(w, jsonResult)
				return
			}
			defer func(file multipart.File) {
				err := file.Close()
				if err != nil {
					return
				}
			}(file)

			str, err := handlerFile(storageMask[0], file, handler)

			if err != nil {
				jsonResult, _ = json.Marshal(Result{Status: false, Message: str, Error: err.Error()})
				w.WriteHeader(http.StatusServiceUnavailable)
				result(w, jsonResult)
			} else {
				jsonResult, _ = json.Marshal(Result{
					Status:           true,
					HostUrl:          os.Getenv("HOST_URL"),
					PathServer:       str,
					Size:             handler.Size,
					FilenameUploaded: handler.Filename,
				})
				result(w, jsonResult)
				return
			}

			jsonResult, _ = json.Marshal(Result{Status: true, Message: "Service is running..."})
			result(w, jsonResult)
		}
	})

	err := http.ListenAndServe(":8017", nil)
	if err != nil {
		log.Fatalln(err)
	}
}

func result(w http.ResponseWriter, jsonResult []byte) {
	w.Header().Set("Content-Type", "application/json")

	_, err := fmt.Fprintf(w, string(jsonResult))
	if err != nil {
		log.Error(err.Error())
		return
	}
}

func getFiles(r *http.Request) ([]byte, error) {
	rdWith := rootDir + "/"
	var (
		files        []File
		items        int
		itemsRemoved int
		size         int64
		sizeRemoved  int64
		getAllPaths  bool
	)

	all, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	var result map[string][]string
	err = json.Unmarshal(all, &result)
	if err != nil {
		return nil, err
	}

	if len(result["paths"]) == 1 && result["paths"][0] == "all" {
		getAllPaths = true
	}

	err = filepath.Walk(rootDirRemoved,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			itemsRemoved++
			sizeRemoved = sizeRemoved + info.Size()
			return nil
		})
	if err != nil {
		return nil, err
	}

	err = filepath.Walk(rootDir,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			items++
			size = size + info.Size()

			ptr := strings.TrimPrefix(p, rdWith)
			idx := slices.Index(result["paths"], ptr)

			if idx != -1 || getAllPaths == true {
				files = append(files, File{
					ptr,
					info.Size(),
					info.ModTime(),
				})
			}
			return nil
		})
	if err != nil {
		return nil, err
	}

	jsonResult, _ := json.Marshal(Files{
		Status:       true,
		Items:        items,
		ItemsRemoved: itemsRemoved,
		Size:         size,
		SizeRemoved:  sizeRemoved,
		Files:        files,
	})

	return jsonResult, nil
}

func remove(r *http.Request) error {
	all, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}

	var result map[string]string
	err = json.Unmarshal(all, &result)
	if err != nil {
		return err
	}

	pathRem, is := result["path"]

	if is == false {
		return errors.New("no path")
	}

	old := rootDir + "/" + pathRem
	new := rootDirRemoved + "/" + pathRem

	read, err := os.Open(old)
	if err != nil {
		return err
	}
	defer read.Close()

	err = os.MkdirAll(path.Dir(new), os.ModePerm)
	if err != nil {
		return err
	}

	write, err := os.Create(rootDirRemoved + "/" + pathRem)
	if err != nil {
		log.Error(err)
		return err
	}
	defer write.Close()

	_, err = io.Copy(write, read)
	if err != nil {
		return err
	}

	err = os.Remove(old)
	if err != nil {
		return err
	}

	return nil
}

func handlerFile(storageMask string, file multipart.File, handler *multipart.FileHeader) (string, error) {
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", err
	}

	ext := path.Ext(handler.Filename)
	if ext == "" {
		return "no extension", errors.New("no ext")
	}

	rand.Seed(time.Now().UnixNano())
	b := make([]byte, 20+2)
	rand.Read(b)
	nameFile := genMd5(strconv.FormatInt(handler.Size, 10) + handler.Filename +
		fmt.Sprintf("%x", b)[2:20+2])

	middlePath := "/" + string(nameFile[0]) + "/" + string(nameFile[1]) + string(nameFile[2])
	fullPath := rootDir + "/" + storageMask + middlePath

	err = os.MkdirAll(fullPath, os.ModePerm)
	if err != nil {
		return "can't create dirs", err
	}

	newNameWithExt := "/" + nameFile + ext

	if _, err := os.Stat(fullPath + newNameWithExt); !errors.Is(err, os.ErrNotExist) {
		log.Error("collision - " + fullPath + newNameWithExt)
		return "fatal, path and filename are existing", errors.New("collision")
	}

	openFile, err := os.OpenFile(fullPath+newNameWithExt, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0777)
	defer openFile.Close()
	if err != nil {
		return "can't open file", err
	}

	_, err = openFile.Write(fileBytes)
	if err != nil {
		return "can't write file", err
	}

	return storageMask + middlePath + newNameWithExt, nil
}

func genMd5(string string) string {
	h := md5.New()
	_, err := io.WriteString(h, string)
	if err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func setup() {
	if os.Getenv("DEV") == "true" {
		rootDir = rootDir[1:]
		rootDirRemoved = rootDirRemoved[1:]
	}

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
		CallerPrettyfier: func(f *runtime.Frame) (string, string) {
			filename := path.Base(f.File)
			return fmt.Sprintf("%s()", f.Function), fmt.Sprintf(" %s:%d", filename, f.Line)
		},
	})
	if l, err := log.ParseLevel("debug"); err == nil {
		log.SetLevel(l)
		log.SetReportCaller(l == log.DebugLevel)
		log.SetOutput(os.Stdout)
	}
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}
