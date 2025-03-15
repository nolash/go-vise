package engine

import (
	"fmt"
)

// Config globally defines behavior of all components driven by the engine.
type Config struct {
	// OutputSize sets the maximum size of output from a single rendered page. If set to 0, no size limit is imposed.
	OutputSize uint32
	// SessionId is used to segment the context of state and application data retrieval and storage.
	SessionId string
	// Root is the node name of the bytecode entry point.
	Root string
	// FlagCount is used to set the number of user-defined signal flags used in the execution state.
	FlagCount uint32
	// CacheSize determines the total allowed cumulative cache size for a single SessionId storage segment. If set to 0, no size limit is imposed.
	CacheSize uint32
	// Language determines the ISO-639-3 code of the default translation language. If not set, no language translations will be looked up.
	Language string
	// StateDebug activates string translations of flags in output logs if set
	StateDebug bool
	// MenuSeparator sets the string to use for separating menu selectors and menu descriptors in the renderer
	MenuSeparator string
	// ResetOnEmptyInput purges cache and restart state execution at root on empty input
	ResetOnEmptyInput bool
}

// String implements the string interface.
func (c Config) String() string {
	return fmt.Sprintf("sessionid '%s', rootpath '%s', flagcount %d, language '%s'", c.SessionId, c.Root, c.FlagCount, c.Language)
}
