package backend

import (
	"regexp"
	"strings"
)

var OldLinkRegexp *regexp.Regexp
var NewLinkRegexp *regexp.Regexp

func InitializeLinkRegexp() {
	OldLinkRegexp = regexp.MustCompile(`</([^/]+)/([^/]+)/([^/]+)>; ?riaktag="([^"]+)"`)
	NewLinkRegexp = regexp.MustCompile(`</(buckets)/([^/]+)/keys/([^/]+)>; ?riaktag="([^"]+)"`)
}

type Link struct {
	bucket string
	key string
	tag string
}

func ParseLink(linkstr string) *Link {
	linkstr = strings.Trim(linkstr, " ")
	match := OldLinkRegexp.FindStringSubmatch(linkstr)
	if match == nil || len(match) < 4 {
		match = NewLinkRegexp.FindStringSubmatch(linkstr)
		if match == nil || len(match) < 4 {
			return nil
		}
	}

	link := new(Link)
	link.bucket = match[2]
	link.key = match[3]
	if len(match) == 5 {
		link.tag = match[4]
	}
	return link
}

func QueryLinks(linksheader, bucket, tag string) []*Link {
	links := strings.Split(linksheader, ",")

	results := make([]*Link, 0, len(links))
	for _, linkstr := range links {
		link := ParseLink(linkstr)
		if (link != nil) && (bucket == "_" || link.bucket == bucket) && (tag == "_" || link.tag == tag) {
			results = append(results, link)
		}
	}
	return results
}