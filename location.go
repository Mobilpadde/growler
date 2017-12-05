package growler

import (
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
)

// Query is a struct for query data
type Query struct {
	Query  string
	Result string
}

// Location is a struct for holding the necessary information about a url
type Location struct {
	Source   string
	Protocol string
	Host     string
	Port     int
	Path     []string
	Query    []*Query
	Hash     string
}

var destructor = regexp.MustCompile("((https?)\\:[\\/]{2}([a-zA-Z0-9\\.\\-]+(\\:\\d+)?)([a-zA-Z0-9\\/\\-]+)?(\\?([a-zA-Z0-9\\=\\&]+))?(\\#[a-zA-Z0-9\\=\\/]+)?)")

func deconstructURL(url string) *Location {
	ctx := destructor.FindStringSubmatch(url)

	port, err := strconv.Atoi(ctx[4])
	if err != nil {
		port = 80
	}

	loc := &Location{
		Source:   url,
		Protocol: ctx[2],
		Host:     ctx[3],
		Port:     port,
		Path:     strings.Split(ctx[5], "/"),
		// Auery
		Hash: ctx[8],
	}

	queries := strings.Split(ctx[7], "&")
	lenQ := len(queries)

	if lenQ > 1 {
		for i := 0; i < lenQ; i++ {
			query := strings.Split(queries[i], "=")

			if len(query) > 1 {
				loc.Query = append(loc.Query, &Query{
					Query:  query[0],
					Result: query[1],
				})
			} else {
				loc.Query = append(loc.Query, &Query{Query: query[0]})
			}
		}
	} else {
		query := strings.Split(ctx[7], "=")

		if len(query) > 1 {
			loc.Query = append(loc.Query, &Query{
				Query:  query[0],
				Result: query[1],
			})
		} else {
			loc.Query = append(loc.Query, &Query{Query: query[0]})
		}
	}

	return loc
}

func (growl *Growler) find(rc io.ReadCloser) ([]string, bool) {
	b, err := ioutil.ReadAll(rc)

	if err == nil {
		str := string(b)

		linksAll := destructor.FindAllString(str, -1)
		var links []string

		for i := 0; i < len(linksAll); i++ {
			links = append(links, linksAll[i])
		}

		return links, len(growl.match.FindString(str)) > 0
	}

	return make([]string, 0), false
}
