package middleware

import (
	"fmt"
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

func NewPattern(p string) (Pattern, error) {
	i := strings.IndexByte(p, '*')
	if i == -1 {
		return eqPattern{p}, nil
	}
	if p == "*" {
		return anyPattern, nil
	}
	il := strings.LastIndexByte(p, '*')
	if i == il {
		if i == 0 {
			return suffixPattern{p[i+1:]}, nil
		}
		if il == len(p)-1 {
			return prefixPattern{p[0:i]}, nil
		}
		return presufPattern{pre: p[0:i], suf: p[i+1:]}, nil
	}
	return nil, fmt.Errorf("not support multiple wildcard pattern: %v", p)
}
