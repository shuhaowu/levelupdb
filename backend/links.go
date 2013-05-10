/*
 * This file is part of levelupdb.
 *
 * levelupdb is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * levelupdb is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with levelupdb.  If not, see <http://www.gnu.org/licenses/>.
 */
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
	Bucket string
	Key string
	Tag string
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
	link.Bucket = match[2]
	link.Key = match[3]
	if len(match) == 5 {
		link.Tag = match[4]
	}
	return link
}

func QueryLinks(linksheader, bucket, tag string) []*Link {
	links := strings.Split(linksheader, ",")

	results := make([]*Link, 0, len(links))
	for _, linkstr := range links {
		link := ParseLink(linkstr)
		if (link != nil) && (bucket == "_" || link.Bucket == bucket) && (tag == "_" || link.Tag == tag) {
			results = append(results, link)
		}
	}
	return results
}

func (database *Database) GetObjectFromLink(link *Link) (*Meta, []byte, error){
	return database.GetObject(link.Bucket, link.Key)
}