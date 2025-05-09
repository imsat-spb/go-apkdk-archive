package archive

import "strings"

type RequestItem struct {
	request string
	item    string
}

func (item *RequestItem) AddToBuilder(builder *strings.Builder) {
	builder.WriteString(item.request)
	builder.WriteString("\n")
	builder.WriteString(item.item)
	builder.WriteString("\n")
}
