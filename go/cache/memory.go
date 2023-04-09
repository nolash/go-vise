package cache

type Memory interface {
	Add(key string, val string, sizeLimit uint16) error
	Update(key string, val string) error
	ReservedSize(key string) (uint16, error)
	Get() (map[string]string, error)
	Push() error
	Pop() error
	Reset()
}