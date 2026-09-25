package main

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gilmardealcantara/shortener/db"
	"github.com/Gilmardealcantara/shortener/pkg/config"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	cfg := config.New()
	db.InitPostgres(context.Background(), cfg.PostgresDSN)
	defer db.ClosePostgres(context.Background())

	db.InitRedis(context.Background())
	defer db.CloseRedis(context.Background())

	testServer := httptest.NewServer(getServer(cfg))

	t.Run("Get", func(t *testing.T) {
		resp, err := http.Get(testServer.URL + "/xyz")
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		assert.Equal(t, "xyz", string(body))
	})

	t.Run("Post", func(t *testing.T) {
		payload := `{"long_url": "http://pudim.com.br"}`
		resp, err := http.Post(testServer.URL+"/shorten", "text/plain", strings.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var data ShortenerResponse
		json.Unmarshal(body, &data)
		assert.Regexp(t, `^http://localhost:8080/.*$`, data.ShortURL)
	})
}
