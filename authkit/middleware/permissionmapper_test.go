package middleware

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultPermissionMapper(t *testing.T) {
	customMapper := &DefaultPermissionMapper{
		DefaultScopes: []string{"core", "some-other"},
		ScopePrefix:   "xxx.",
		RootResName:   "_root_",
		Mapping: map[Route]DefaultPermission{
			Route{
				Method: "GET",
				Path:   "/aa/bb/cc",
			}: DefaultPermission{
				Resource: "res:cc-bb-aa",
				Action:   "view",
				Scopes:   []string{"first-scope", "second-scope"},
			},
			Route{
				Method: "POST",
				Path:   "/aa/bb/cc",
			}: DefaultPermission{
				Resource: "res:some_resource",
				Action:   "edit",
				Scopes:   []string{"test-scope"},
			},
		},
	}

	cases := []struct {
		name   string
		method string
		path   string
		perm   *DefaultPermission
		pm     *DefaultPermissionMapper
	}{
		{
			name:   "Empty method",
			method: "",
			path:   "/aa/bb",
			pm:     &DefaultPermissionMapper{},
		},
		{
			name:   "Empty path",
			method: "GET",
			path:   "",
			pm:     &DefaultPermissionMapper{},
		},
		{
			name:   "Root path",
			method: "GET",
			path:   "/",
			perm: &DefaultPermission{
				Resource: "rn:root",
				Action:   "get",
				Scopes:   []string{"root.get"},
			},
			pm: &DefaultPermissionMapper{},
		},
		{
			name:   "GET",
			method: "GET",
			path:   "/aa",
			perm: &DefaultPermission{
				Resource: "rn:aa",
				Action:   "get",
				Scopes:   []string{"aa.get"},
			},
			pm: &DefaultPermissionMapper{},
		},
		{
			name:   "POST",
			method: "POST",
			path:   "/aa/bb/cc/",
			perm: &DefaultPermission{
				Resource: "rn:aa:bb:cc",
				Action:   "post",
				Scopes:   []string{"aa.bb.cc.post"},
			},
			pm: &DefaultPermissionMapper{},
		},
		{
			name:   "DELETE",
			method: "DELETE",
			path:   "/aa%20zz/bb-dd/cc+xx/ee%2Bff/gg_hh",
			perm: &DefaultPermission{
				Resource: "rn:aa+zz:bb-dd:cc+xx:ee+ff:gg_hh",
				Action:   "delete",
				Scopes:   []string{"aa+zz.bb-dd.cc+xx.ee+ff.gg_hh.delete"},
			},
			pm: &DefaultPermissionMapper{},
		},
		{
			name:   "PUT",
			method: "PUT",
			path:   "/aa/bb/cc/",
			perm: &DefaultPermission{
				Resource: "rn:aa:bb:cc",
				Action:   "put",
				Scopes:   []string{"aa.bb.cc.put"},
			},
			pm: &DefaultPermissionMapper{},
		},
		{
			name:   "With default scopes and prefix",
			method: "GET",
			path:   "/aa/bb/cc/",
			perm: &DefaultPermission{
				Resource: "rn:aa:bb:cc",
				Action:   "get",
				Scopes:   []string{"core", "some-other", "custom.aa.bb.cc.get"},
			},
			pm: &DefaultPermissionMapper{
				DefaultScopes: []string{"core", "some-other"},
				ScopePrefix:   "custom.",
				RootResName:   "_root_",
			},
		},
		{
			name:   "Custom root",
			method: "GET",
			path:   "/",
			perm: &DefaultPermission{
				Resource: "rn:_root_",
				Action:   "get",
				Scopes:   []string{"core", "some-other", "xxx._root_.get"},
			},
			pm: &DefaultPermissionMapper{
				DefaultScopes: []string{"core", "some-other"},
				ScopePrefix:   "xxx.",
				RootResName:   "_root_",
			},
		},
		{
			name:   "Custom mapping GET",
			method: "GET",
			path:   "/aa/bb/cc",
			perm: &DefaultPermission{
				Resource: "res:cc-bb-aa",
				Action:   "view",
				Scopes:   []string{"first-scope", "second-scope"},
			},
			pm: customMapper,
		},
		{
			name:   "Custom mapping POST",
			method: "POST",
			path:   "/aa/bb/cc",
			perm: &DefaultPermission{
				Resource: "res:some_resource",
				Action:   "edit",
				Scopes:   []string{"test-scope"},
			},
			pm: customMapper,
		},
	}

	for _, c := range cases {
		c := c
		t.Run(c.name, func(st *testing.T) {
			st.Parallel()
			assert := assert.New(st)
			p, err := c.pm.RequiredPermissioin(c.method, c.path)
			if c.perm == nil {
				assert.Error(err)
				assert.Nil(p)
			} else {
				assert.NoError(err)
				assert.Equal(c.perm, p)
				assert.IsType(c.perm, p)
			}
		})
	}
}
