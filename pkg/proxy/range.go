package proxy

import (
	"errors"
	"fmt"
	"net/textproto"
	"strconv"
	"strings"
)

type Range struct {
	Start  int64
	Length int64
}

func (r Range) ContentRange(total int64) string {
	return fmt.Sprintf("bytes %d-%d/%d", r.Start, r.Start+r.Length-1, total)
}

func (r Range) RangeString() string {
	return fmt.Sprintf("bytes=%d-%d", r.Start, r.Start+r.Length-1)
}

var (
	ErrInvalid = errors.New("invalid range")
)

func ParseRange(s string, max int64) (*Range, error) {
	if s == "" {
		return nil, nil
	}
	const b = "bytes="
	if !strings.HasPrefix(s, b) {
		return nil, ErrInvalid
	}

	ra := textproto.TrimString(s[len(b):])

	i := strings.Index(ra, "-")
	if i < 0 {
		return nil, ErrInvalid
	}

	start, end := textproto.TrimString(ra[:i]), textproto.TrimString(ra[i+1:])

	r := &Range{}

	r.Start, _ = strconv.ParseInt(start, 10, 64)

	if end != "" {
		n, _ := strconv.ParseInt(end, 10, 64)
		r.Length = n - r.Start + 1
	}

	if r.Length > max {
		r.Length = max
	}

	return r, nil
}

const (
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB

	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
)
