package growler

import (
	"io/ioutil"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func (growl *Growler) isAllowed(host string, path []string) bool {
	pathRegex := regexp.MustCompile("((" + strings.Join(path, ")|(") + "jklol))")

	for k := range growl.ignore {
		if pathRegex.MatchString(k) {
			return false
		}
	}

	res, err := http.Get(host + "/robots.txt")

	if err != nil {
		return true
	}

	defer res.Body.Close()
	b, err := ioutil.ReadAll(res.Body)
	if err == nil {
		return true
	}

	str := string(b)
	split := strings.Split("\n", str)

	wait := regexp.MustCompile("(Crawl-delay\\:\\s(\\d+))")
	disallow := regexp.MustCompile("(Disallow\\:\\s(.+))")
	// hostRegex := regexp.MustCompile(host)

	allow := true
	for i := range split {
		if wait.MatchString(split[i]) {
			delayStr := wait.FindStringSubmatch(split[i])[2]
			delay, _ := strconv.Atoi(delayStr) // missing err

			growl.Lock()
			growl.wait[host] = time.Duration(delay) * time.Millisecond
			growl.Unlock()
		} else if disallow.MatchString(split[i]) {
			disallowed := disallow.FindAllStringSubmatch(split[i], -1)

			for i := range disallowed {
				growl.ignore[host+disallowed[i][2]] = true
			}

			if pathRegex.MatchString(split[i]) {
				allow = false
			}
		}
	}

	return allow
}
