package middleware

import (
	"errors"
	"net/url"
	"unicode"
)

import "strings"

type (

	// PermissionMapper used to map method and path of http request to desirable
	// permission descriptor. Permission descriptor is an interface, passed to
	// the TokenValidator. For example, in case of Hydra-backed TokenValidator,
	// permission descriptor contains resource name, action and scope.
	PermissionMapper interface {

		// RequiredPermissioin returns permission descriptor to be passed to
		// TokenValidator. It may return an error to prevent access to resource
		// without request to TokenValidator.
		RequiredPermissioin(method, path string) (interface{}, error)
	}

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

		// DefaultScopes, ScopePrefix and RootResName are ignored when explicite mapping exists for route.
		Mapping map[Route]DefaultPermission
	}
)

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
