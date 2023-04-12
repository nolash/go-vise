package cache

// Memory defines the interface for store of a symbol mapped content store.
type Memory interface {
	Add(key string, val string, sizeLimit uint16) error
	Update(key string, val string) error
	ReservedSize(key string) (uint16, error)
	Get(key string) (string, error)
	Push() error
	Pop() error
	Reset()
}
