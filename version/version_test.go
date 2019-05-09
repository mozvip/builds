package version

import (
	"testing"
	"time"
)

func TestVersionCompare1(t *testing.T) {
	version1 := Version{"12.0", time.Time{}}
	version2 := Version{"12.1", time.Time{}}

	if version1.After(&version2) {
		t.Fail()
	}
}

func TestVersionCompare2(t *testing.T) {
	version1 := Version{"1.14", time.Time{}}
	version2 := Version{"1.20", time.Time{}}

	if version1.After(&version2) {
		t.Fail()
	}
}