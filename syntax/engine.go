package syntax

// NullEngine is a no-op syntax provider.
type NullEngine struct{}

func (NullEngine) Language() string                          { return "" }
func (NullEngine) NotifyEdit(Edit)                           {}
func (NullEngine) Parse([]byte) error                        { return nil }
func (NullEngine) HighlightLine(int) []Span                  { return nil }
func (NullEngine) HighlightViewport(int, int) map[int][]Span { return map[int][]Span{} }
