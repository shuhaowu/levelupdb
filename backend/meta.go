package backend

import (
	"net/http"
	"strings"
)

type Meta struct {
	Indexes     [][2]string       `json:"I"`
	Links       string            `json:"L"`
	Meta        map[string]string `json:"M"`
	ContentType string            `json:"C"`
}

func MetaFromRequest(req *http.Request) (*Meta, error) {
	meta := new(Meta)

	meta.Links = req.Header.Get("Link")
	meta.ContentType = req.Header.Get("Content-Type")
	meta.Meta = make(map[string]string)
	for headerKey, headerValue := range req.Header {
		headerValueLength := len(headerValue)
		if strings.HasPrefix(headerKey, "X-Riak-Index-") && headerValueLength > 0 {
			indexKey := strings.ToLower(headerKey[13:]) // case insenstive because go convert the first character into caps?
			index := [2]string{indexKey, headerValue[0]}
			meta.Indexes = append(meta.Indexes, index)
		}

		if strings.HasPrefix(headerKey, "X-Riak-Meta-") && headerValueLength > 0 {
			metaKey := strings.ToLower(headerKey[12:]) // same reason as above.
			meta.Meta[metaKey] = headerValue[0]
		}
	}

	return meta, nil
}

func (meta *Meta) ToHeaders(headers http.Header) {
	if meta.Links != "" {
		headers.Add("Link", meta.Links)
	}
	headers.Add("Content-Type", meta.ContentType)
	for _, index := range meta.Indexes {
		headers.Add("X-Riak-Index-"+index[0], index[1])
	}

	for k, v := range meta.Meta {
		headers.Add("X-Riak-Meta-"+k, v)
	}
	headers.Add("X-Riak-Vclock", "Yay02966e9d038d6332eea23012217f8c4b521eaf92==")
}