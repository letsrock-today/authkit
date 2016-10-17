package middleware

import (
	"errors"
	"net/url"
	"strings"
	"unicode"
)

type (

	// DefaultPermission is a permission descriptor for Hydra-backed TokenValidator.
	// Fields of this struct should be passed along with token to Hydra (or similar) API.
	DefaultPermission struct {
		Resource string
		Action   string
		Scopes   []string
	}

	// Route used as a key to provide explicite mapping.
	Route struct {
		Method string
		Path   string
	}

	// DefaultPermissionMapper maps method and path of request to DefaultPermission.
	// See unit tests for additional info.
	DefaultPermissionMapper struct {

		// DefaultScopes added to calculated scope.
		DefaultScopes []string

		// ScopePrefix prefixes calculated scope name.
		ScopePrefix string

		// RootResName used for root path ("/").
		RootResName string

		// DefaultScopes, ScopePrefix and RootResName are ignored when explicit
		// mapping exists for route.
		Mapping map[Route]DefaultPermission
	}
)

// RequiredPermissioin maps method and path to Resource, Action and Scopes
// according to following rules.
// Resource is created from path with "rn:" prefix and slashes replaced to colons.
// Root resource replaced with "rn:root" or with "rn:" + RootResName.
// Scopes added to slice of DefaultScopes.
// Scope names created from path with slashes replaced with dots.
// Every created scope name is prefixed with ScopePrefix and postfixed with
// lower-cased http method name.
// Action is a lower-cased method name.
// See tests for examples.
func (m DefaultPermissionMapper) RequiredPermissioin(method, path string) (interface{}, error) {
	if method == "" || path == "" {
		return nil, errors.New("illegal arguments")
	}
	// exact mapping has the most priority
	if m.Mapping != nil {
		r := Route{Method: method, Path: path}
		if p, ok := m.Mapping[r]; ok {
			return &p, nil
		}
	}
	a := strings.ToLower(method)
	r := strings.Trim(path, "/")
	if r == "" {
		r = m.RootResName
		if r == "" {
			r = "root"
		}
	} else {
		var err error
		r, err = url.QueryUnescape(r)
		if err != nil {
			return nil, err
		}
		r = strings.Map(func(r rune) rune {
			switch r {
			case ' ':
				return '+'
			case '/':
				return ':'
			}
			return unicode.ToLower(r)
		}, r)
	}
	s := m.ScopePrefix + strings.Replace(r, ":", ".", -1) + "." + a
	r = "rn:" + r
	p := &DefaultPermission{
		Resource: r,
		Action:   a,
		Scopes:   append(m.DefaultScopes, s),
	}
	return p, nil
}
