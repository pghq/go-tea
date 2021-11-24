package tea

import (
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-tea/internal"
)

func TestNewApp(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		app, err := NewApp()
		assert.Nil(t, err)
		assert.NotNil(t, app)
	})

	t.Run("raises bad dsn errors", func(t *testing.T) {
		_ = os.Setenv("SENTRY_DSN", "https://localhost")
		defer os.Clearenv()
		_, err := NewApp()
		assert.NotNil(t, err)
	})
}

func TestApp_Health(t *testing.T) {
	app, err := NewApp()
	assert.Nil(t, err)

	t.Run("can create instance", func(t *testing.T) {
		assert.NotNil(t, app.Health())
	})
}

func TestApp_Router(t *testing.T) {
	app, _ := NewApp()

	t.Run("can create instance", func(t *testing.T) {
		r := app.Router()
		assert.NotNil(t, r)

		r.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/v0/health/status", nil))
	})
}

func TestEnvironment(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		o := Environment("test")
		assert.NotNil(t, o)
	})
}

func TestEnvironmentOption_Apply(t *testing.T) {
	o := Environment("test")

	t.Run("can apply", func(t *testing.T) {
		conf := internal.AppConfig{}
		o.Apply(&conf)
		assert.Equal(t, "test", conf.Environment)
	})

	t.Run("can create instance", func(t *testing.T) {
		app, _ := NewApp(o)
		assert.Equal(t, app.environment, "test")
	})
}

func TestVersion(t *testing.T) {
	t.Run("can create instance", func(t *testing.T) {
		o := Version("1.0.0")
		assert.NotNil(t, o)
	})
}

func TestVersionOption_Apply(t *testing.T) {
	o := Version("1.0.0")

	t.Run("raises bad version errors", func(t *testing.T) {
		o := Version("")
		_, err := NewApp(o)
		assert.NotNil(t, err)
	})

	t.Run("can apply", func(t *testing.T) {
		conf := internal.AppConfig{}
		o.Apply(&conf)
		assert.Equal(t, "1.0.0", conf.Version)
	})

	t.Run("can create instance", func(t *testing.T) {
		app, _ := NewApp(o)
		assert.Equal(t, app.version.String(), "1.0.0")
	})
}
