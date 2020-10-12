package sessions

import (
	"bytes"
	"math/rand"
	"strconv"
	"testing"
)

func TestTemplate(t *testing.T) {
	templateParams := map[string]interface{}{
		"SessionID": strconv.Itoa(rand.Int()),
	}
	var buf bytes.Buffer
	err := tmpl.Execute(&buf, templateParams)
	if err != nil {
		t.Error(err)
	}
}
