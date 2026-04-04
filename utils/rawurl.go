package utils

import (
	"fmt"
	"net"
	neturl "net/url"
	"strings"
)

type RawURL struct {
	Raw       string
	Scheme    string
	Authority string
	Host      string
	Path      string
	RawQuery  string
	Fragment  string
}

func ParseRawURL(input string) (RawURL, error) {
	if input == "" {
		return RawURL{}, fmt.Errorf("empty URL")
	}

	parsedURL, err := neturl.Parse(input)
	if err != nil {
		return RawURL{}, err
	}
	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return RawURL{}, fmt.Errorf("invalid URL: %s", input)
	}

	scheme := parsedURL.Scheme
	schemeSplit := strings.SplitN(input, "://", 2)
	if len(schemeSplit) != 2 || schemeSplit[1] == "" {
		return RawURL{}, fmt.Errorf("invalid URL: %s", input)
	}
	remainder := schemeSplit[1]
	authorityEnd := len(remainder)
	for i, ch := range remainder {
		if ch == '/' || ch == '?' || ch == '#' {
			authorityEnd = i
			break
		}
	}

	authority := remainder[:authorityEnd]
	if authority == "" {
		return RawURL{}, fmt.Errorf("invalid URL authority: %s", input)
	}

	pathAndMore := remainder[authorityEnd:]
	fragment := ""
	if idx := strings.IndexByte(pathAndMore, '#'); idx >= 0 {
		fragment = pathAndMore[idx+1:]
		pathAndMore = pathAndMore[:idx]
	}

	rawQuery := ""
	path := pathAndMore
	if idx := strings.IndexByte(pathAndMore, '?'); idx >= 0 {
		rawQuery = pathAndMore[idx+1:]
		path = pathAndMore[:idx]
	}

	if path == "" {
		path = "/"
	}

	return RawURL{
		Raw:       input,
		Scheme:    scheme,
		Authority: authority,
		Host:      parsedURL.Host,
		Path:      path,
		RawQuery:  rawQuery,
		Fragment:  fragment,
	}, nil
}

func (u RawURL) String() string {
	return u.Raw
}

func (u RawURL) Base() string {
	return u.Scheme + "://" + u.Authority
}

func (u RawURL) RequestURI() string {
	requestURI := u.Path
	if requestURI == "" {
		requestURI = "/"
	}
	if u.RawQuery != "" {
		requestURI += "?" + u.RawQuery
	}
	return requestURI
}

func (u RawURL) WithRequestURI(requestURI string) RawURL {
	next := u
	next.Raw = u.Base() + requestURI
	next.Path = requestURI
	next.RawQuery = ""
	next.Fragment = ""

	if idx := strings.IndexByte(requestURI, '?'); idx >= 0 {
		next.Path = requestURI[:idx]
		next.RawQuery = requestURI[idx+1:]
	}

	if next.Path == "" {
		next.Path = "/"
	}

	return next
}

func (u RawURL) WithPath(path string) RawURL {
	requestURI := path
	if u.RawQuery != "" {
		requestURI += "?" + u.RawQuery
	}
	return u.WithRequestURI(requestURI)
}

func (u RawURL) AppendToPath(suffix string) RawURL {
	return u.WithPath(u.Path + suffix)
}

func (u RawURL) DirPath() string {
	path := u.Path
	if path == "" || path == "/" {
		return "/"
	}

	trimmed := strings.TrimSuffix(path, "/")
	if trimmed == "" {
		return "/"
	}

	lastSlash := strings.LastIndex(trimmed, "/")
	if lastSlash <= 0 {
		return "/"
	}

	return trimmed[:lastSlash]
}

func (u RawURL) ParsedForDisplay() *neturl.URL {
	port := u.Port()
	hostForURL := u.Hostname()
	if strings.Contains(hostForURL, ":") && !strings.HasPrefix(hostForURL, "[") {
		hostForURL = "[" + hostForURL + "]"
	}
	if port != "" {
		hostForURL = net.JoinHostPort(hostForURL, port)
	}

	return &neturl.URL{
		Scheme:   u.Scheme,
		Host:     hostForURL,
		Path:     u.Path,
		RawPath:  u.Path,
		RawQuery: u.RawQuery,
	}
}

func (u RawURL) Hostname() string {
	return stripPort(u.Host)
}

func (u RawURL) Port() string {
	return portOnly(u.Host)
}

func stripPort(host string) string {
	if strings.HasPrefix(host, "[") {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			return strings.Trim(parsedHost, "[]")
		}
		return strings.Trim(host, "[]")
	}

	if strings.Count(host, ":") == 1 {
		if parsedHost, _, err := net.SplitHostPort(host); err == nil {
			return parsedHost
		}
	}

	return host
}

func portOnly(host string) string {
	if strings.HasPrefix(host, "[") || strings.Count(host, ":") == 1 {
		if _, port, err := net.SplitHostPort(host); err == nil {
			return port
		}
	}
	return ""
}
