package archive

type updateEventInfo interface {
	getArchiveServerRequest() []*RequestItem
}

type createRequest struct {
	Index   string `json:"_index"`
	Id      string `json:"_id"`
	DocType string `json:"_type"`
}
