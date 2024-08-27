package engine

// Config globally defines behavior of all components driven by the engine.
type Config struct {
	OutputSize uint32 // Maximum size of output from a single rendered page
	SessionId string
	Root string
	FlagCount uint32
	CacheSize uint32
	Language string
}
