package tests

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Gilmardealcantara/shortener/db"
	"github.com/Gilmardealcantara/shortener/pkg/handlers"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	ctx := context.Background()
	cfg := SetupContainers(ctx)

	db.InitPostgres(ctx, cfg.PostgresDSN)
	defer db.ClosePostgres(ctx)

	db.InitRedis(ctx, cfg.RedisDSN)
	defer db.CloseRedis(ctx)

	testServer := httptest.NewServer(SetupMuxServer())
	slog.Info("TestRedirect", "testServer", testServer.URL)

	// Create a custom client that blocks redirect-following
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}

	t.Run("Post", func(t *testing.T) {
		payload := `{"long_url": "http://pudim.com.br"}`
		resp, err := http.Post(testServer.URL+"/shorten", "text/plain", strings.NewReader(payload))
		if err != nil {
			t.Fatal(err)
		}
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		body, _ := io.ReadAll(resp.Body)
		var data handlers.ShortenerResponse
		json.Unmarshal(body, &data)
		assert.Regexp(t, testServer.URL+`.*$`, data.ShortURL)

		t.Run("Geting the shortened URL"+data.ShortURL, func(t *testing.T) {
			resp, err := client.Get(data.ShortURL)
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, http.StatusFound, resp.StatusCode)
			assert.Equal(t, "http://pudim.com.br", resp.Header.Get("Location"))
		})

		t.Run("Get Not Found", func(t *testing.T) {
			resp, err := client.Get(testServer.URL + "/xyz")
			if err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	})
}
