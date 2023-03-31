package resource


type Fetcher interface {
	Get(symbol string) (string, error)
	Render(symbol string, values map[string]string) (string, error)
}
