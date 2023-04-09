package render

type Renderer interface {
	Map(key string) error
	Render(sym string, values map[string]string, idx uint16) (string, error)
	Reset()
}
