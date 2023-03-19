package main

import "time"

type Result struct {
	Status           bool   `json:"status"`
	Message          string `json:"message,omitempty"`
	Error            string `json:"error,omitempty"`
	HostUrl          string `json:"hostUrl,omitempty"`
	PathServer       string `json:"pathServer,omitempty"`
	Size             int64  `json:"size,omitempty"`
	FilenameUploaded string `json:"filenameUploaded,omitempty"`
}

type File struct {
	Filename string    `json:"filename"`
	Size     int64     `json:"size"`
	ModTime  time.Time `json:"modTime"`
}

type Files struct {
	Status       bool   `json:"status"`
	Message      string `json:"message,omitempty"`
	Error        string `json:"error,omitempty"`
	Items        int    `json:"items"`
	ItemsRemoved int    `json:"itemsRemoved"`
	Size         int64  `json:"size"`
	SizeRemoved  int64  `json:"sizeRemoved"`
	Files        []File `json:"Files"`
}
