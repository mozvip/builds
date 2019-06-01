package version

import (
	"testing"
)

func TestVersionCompare1(t *testing.T) {
	version1 := Version{StringVersion:"12.0"}
	version2 := Version{StringVersion:"12.1"}

	if version1.After(&version2) {
		t.Fail()
	}
}

func TestVersionCompare2(t *testing.T) {
	version1 := Version{StringVersion:"1.14"}
	version2 := Version{StringVersion:"1.20"}

	if version1.After(&version2) {
		t.Fail()
	}
}

func TestVersionCompare3(t *testing.T) {
	version1 := Version{FloatVersion:9.1}
	version2 := Version{FloatVersion:12.2}

	if version1.After(&version2) {
		t.Fail()
	}
}