package middleware

import (
	"regexp"
	"strings"
)

type Pattern interface {
	Match(string) bool
}

var anyPattern Pattern = anyPatternT{}

type anyPatternT struct{}

func (p anyPatternT) Match(_ string) bool {
	return true
}

type eqPattern struct {
	p string
}

func (p eqPattern) Match(s string) bool {
	return p.p == s
}

type prefixPattern struct {
	p string
}

func (p prefixPattern) Match(s string) bool {
	return strings.HasPrefix(s, p.p)
}

type suffixPattern struct {
	p string
}

func (p suffixPattern) Match(s string) bool {
	return strings.HasSuffix(s, p.p)
}

type presufPattern struct {
	pre string
	suf string
}

func (p presufPattern) Match(s string) bool {
	return strings.HasPrefix(s, p.pre) && strings.HasSuffix(s, p.suf)
}

type regexPattern struct {
	r *regexp.Regexp
}

func (p regexPattern) Match(s string) bool {
	return p.r.MatchString(s)
}

func NewWildCard(p string) Pattern {
	i := strings.IndexByte(p, '*')
	if i == -1 {
		return eqPattern{p}
	}
	if p == "*" {
		return anyPattern
	}
	// one or more "*"
	var pre, suf string
	if i != 0 {
		pre = p[0:i]
	}
	il := strings.LastIndexByte(p, '*')
	if il != len(p)-1 {
		suf = p[il+1:]
	}
	if i == il {
		// only one "*"
		if pre == "" {
			return suffixPattern{suf}
		}
		if suf == "" {
			return prefixPattern{pre}
		}
		return presufPattern{pre, suf}
	}
	ws := regexp.MustCompile(`\*+`).Split(p, 100)
	for i, w := range ws {
		ws[i] = regexp.QuoteMeta(w)
	}
	r := regexp.MustCompile("^" + strings.Join(ws, "(.*?)") + "$")
	return regexPattern{r}
}
