package config

import (
	"encoding/json"
	"path"
	"runtime"
	"strings"
	"time"
)

type System struct {
	Address string `json:"address"`
	Port    string `json:"port"`

	SessionsCleanInterval string `json:"sessions_clean_interval"`
	SessionsLiveTime      string `json:"sessions_live_time"`
}

func ProjectRootPath() string {
	_, currentFile, _, _ := runtime.Caller(1)
	// currentFile is pkg/config/config.go
	root := path.Dir(path.Dir(path.Dir(currentFile)))
	return root
}

func ParseConfig() (*System, error) {
	//root := ProjectRootPath()
	//configPath := path.Join(root, "data", "web_config.json")
	//f, err := os.Open(configPath)
	//if err != nil {
	//	return nil, err
	//}
	configJSON := `{
	  "address": "127.0.0.1",
	  "port": "8080",
	  "sessions_clean_interval": "10s",
	  "sessions_live_time": "1h"
	}`
	decoder := json.NewDecoder(strings.NewReader(configJSON))
	var conf System
	err := decoder.Decode(&conf)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(conf.SessionsCleanInterval)
	if err != nil {
		return nil, err
	}
	_, err = time.ParseDuration(conf.SessionsLiveTime)
	if err != nil {
		return nil, err
	}
	return &conf, nil
}
