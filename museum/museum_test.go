package museum

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/pghq/go-museum/museum/internal"
)

func TestNew(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		app, err := New()
		assert.Nil(t, err)
		assert.NotNil(t, app)
	})

	t.Run("MonitorError", func(t *testing.T) {
		_ = os.Setenv("SENTRY_DSN", "https://localhost")
		defer os.Clearenv()
		_, err := New()
		assert.NotNil(t, err)
	})
}

func TestApp_Health(t *testing.T) {
	app, err := New()
	assert.Nil(t, err)

	t.Run("NoError", func(t *testing.T) {
		assert.NotNil(t, app.Health())
	})
}

func TestApp_Router(t *testing.T) {
	app, _ := New()

	t.Run("NoError", func(t *testing.T) {
		assert.NotNil(t, app.Router())
	})
}

func TestEnvironment(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		o := Environment("test")
		assert.NotNil(t, o)
	})
}

func TestEnvironmentOption_Apply(t *testing.T) {
	o := Environment("test")

	t.Run("NoError", func(t *testing.T) {
		conf := internal.AppConfig{}
		o.Apply(&conf)
		assert.Equal(t, "test", conf.Environment)
	})

	t.Run("AppOption", func(t *testing.T) {
		app, _ := New(o)
		assert.Equal(t, app.environment, "test")
	})
}

func TestVersion(t *testing.T) {
	t.Run("NoError", func(t *testing.T) {
		o := Version("1.0.0")
		assert.NotNil(t, o)
	})
}

func TestVersionOption_Apply(t *testing.T) {
	o := Version("1.0.0")

	t.Run("BadVersionError", func(t *testing.T) {
		o := Version("")
		_, err := New(o)
		assert.NotNil(t, err)
	})

	t.Run("NoError", func(t *testing.T) {
		conf := internal.AppConfig{}
		o.Apply(&conf)
		assert.Equal(t, "1.0.0", conf.Version)
	})

	t.Run("AppOption", func(t *testing.T) {
		app, _ := New(o)
		assert.Equal(t, app.version.String(), "1.0.0")
	})
}
