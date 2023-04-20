package render

type Renderer interface {
	Keys() []string
	Map(key string) error
}
