package config

import (
	"path"
	"testing"
)

func TestProjectRootPath(t *testing.T) {
	root := ProjectRootPath()
	if path.Base(root) != "wolfmud_web" {
		t.Errorf("can't find project root path")
	}
}

func TestParseConfig(t *testing.T) {
	_, err := ParseConfig()
	if err != nil {
		t.Error(err)
	}
}
