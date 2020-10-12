package config

import (
	"encoding/json"
	"os"
	"path"
	"runtime"
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
	root := ProjectRootPath()
	configPath := path.Join(root, "data", "web_config.json")
	f, err := os.Open(configPath)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(f)
	var conf System
	err = decoder.Decode(&conf)
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
