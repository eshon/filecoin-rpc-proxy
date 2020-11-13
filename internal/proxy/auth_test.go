package proxy

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/protofire/filecoin-rpc-proxy/internal/testhelpers"

	"github.com/protofire/filecoin-rpc-proxy/internal/auth"

	"github.com/protofire/filecoin-rpc-proxy/internal/logger"

	"github.com/stretchr/testify/require"
)

const testMethod = "test"

func TestServerAuxiliaryFunc(t *testing.T) {

	conf, err := testhelpers.GetConfig("http://test.com", testMethod)
	require.NoError(t, err)
	server, err := FromConfig(conf)
	require.NoError(t, err)
	handler := PrepareRoutes(conf, logger.Log, server)

	s := httptest.NewServer(handler)
	defer s.Close()

	paths := []string{"healthz", "ready", "metrics"}

	for idx := range paths {
		path := paths[idx]
		t.Run(fmt.Sprintf("test_%s", path), func(t *testing.T) {
			resp, err := http.Get(fmt.Sprintf("%s/%s", s.URL, path))
			require.NoError(t, err)
			require.Equal(t, 200, resp.StatusCode)
		})
	}
}

func TestServerJWTAuthFunc401(t *testing.T) {

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Kind", "application/json")
		w.WriteHeader(http.StatusOK)
	}))
	defer backend.Close()

	conf, err := testhelpers.GetConfig(backend.URL, testMethod)
	require.NoError(t, err)
	server, err := FromConfig(conf)
	require.NoError(t, err)
	handler := PrepareRoutes(conf, logger.Log, server)

	s := httptest.NewServer(handler)
	defer s.Close()

	resp, err := http.Get(fmt.Sprintf("%s/%s", s.URL, "/test"))
	require.NoError(t, err)
	require.Equal(t, 401, resp.StatusCode)

}

func TestServerJWTAuthFunc(t *testing.T) {

	conf, err := testhelpers.GetConfig("", testMethod)
	require.NoError(t, err)
	jwtToken, err := auth.NewJWT(conf.JWTSecret, conf.JWTAlgorithm, []string{"admin"})
	require.NoError(t, err)

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Content-Kind", "application/json")
		w.WriteHeader(http.StatusOK)
		assert.Equal(t, r.Header.Get("Authorization"), fmt.Sprintf("Bearer %s", jwtToken))
	}))
	defer backend.Close()

	conf.ProxyURL = backend.URL

	server, err := FromConfig(conf)
	require.NoError(t, err)
	handler := PrepareRoutes(conf, logger.Log, server)
	frontend := httptest.NewServer(handler)
	defer frontend.Close()

	url := fmt.Sprintf("%s/%s", frontend.URL, "test")

	req, err := http.NewRequest("GET", url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", jwtToken))

	resp, err := (&http.Client{}).Do(req)
	require.NoError(t, err)
	require.Equal(t, 200, resp.StatusCode)

}
