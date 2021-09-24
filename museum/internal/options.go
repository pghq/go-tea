package internal

// AppOption should provide an option to modify the application
type AppOption interface {
	Apply(conf *AppConfig)
}
