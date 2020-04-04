package handler

import (
	"testing"
)

func TestPatchURI(t *testing.T) {
	t.Log(patchURI("http:/google.com"))
	t.Log(patchURI("https:/google.com"))
	t.Log(patchURI("http://google.com"))
	t.Log(patchURI("https://google.com"))
}
