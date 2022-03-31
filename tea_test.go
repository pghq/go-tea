package tea

import (
	"os"
	"testing"

	"github.com/pghq/go-tea/trail"
)

func TestMain(m *testing.M) {
	trail.Testing()
	os.Exit(m.Run())
}
