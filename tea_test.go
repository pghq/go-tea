package tea

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	Testing()
	os.Exit(m.Run())
}
