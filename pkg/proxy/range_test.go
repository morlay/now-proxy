package proxy

import "testing"

func TestParseRange(t *testing.T) {
	r, _ := ParseRange("bytes=0-4194303", 4*MiB)
	t.Log(r.RangeString())
	r, _ = ParseRange("bytes=4194304-8388607", 4*MiB)
	t.Log(r.RangeString())
	r, _ = ParseRange("bytes=8388608-12582911", 4*MiB)
	t.Log(r.RangeString())
}
