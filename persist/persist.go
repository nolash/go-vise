package persist

type Persister interface {
	Serialize() ([]byte, error)
	Deserialize(b []byte) error
	Save(key string) error
	Load(key string) error
}

