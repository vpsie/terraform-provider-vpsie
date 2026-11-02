package migrate

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/vpsie/govpsie"
)

func testAPI(t *testing.T, handler http.HandlerFunc) *sourceAPI {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return &sourceAPI{client: client}
}

func TestCreateBodyMergesSettingsAndCredentials(t *testing.T) {
	body, err := createBody("vmware-esxi", "lab", "dc-1",
		map[string]string{"host": "10.0.0.5", "port": "443"},
		map[string]string{"password": "s3cret"})
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"type": "vmware-esxi", "name": "lab", "datacenterIdentifier": "dc-1",
		"host": "10.0.0.5", "port": "443", "password": "s3cret",
	}
	for k, v := range want {
		if body[k] != v {
			t.Errorf("body[%s] = %q, want %q", k, body[k], v)
		}
	}
}

func TestCreateBodyRefusesReservedAndDuplicateKeys(t *testing.T) {
	if _, err := createBody("t", "n", "dc", map[string]string{"type": "x"}, nil); err == nil {
		t.Error("a settings key named type must be refused")
	}
	if _, err := createBody("t", "n", "dc", map[string]string{"requestIp": "1.1.1.1"}, nil); err == nil {
		t.Error("a settings key named requestIp must be refused")
	}
	_, err := createBody("t", "n", "dc", map[string]string{"password": "a"}, map[string]string{"password": "b"})
	if err == nil || !strings.Contains(err.Error(), "password") {
		t.Errorf("a key in both settings and credentials must be refused, got %v", err)
	}
}

func TestCreateAndGetSource(t *testing.T) {
	api := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/apps/v2/migrate/sources":
			var got map[string]any
			_ = json.NewDecoder(r.Body).Decode(&got)
			if got["password"] != "s3cret" || got["type"] != "vmware-esxi" {
				t.Errorf("unexpected create body: %v", got)
			}
			_, _ = w.Write([]byte(`{"error":false,"data":{"identifier":"abc-1","type":"vmware-esxi","name":"lab","host":"10.0.0.5","username":"root","datacenterIdentifier":"dc-1","status":"active","createdAt":"2026-10-31T00:00:00Z"}}`))
		case r.Method == http.MethodGet && r.URL.Path == "/apps/v2/migrate/sources/abc-1":
			_, _ = w.Write([]byte(`{"error":false,"data":{"identifier":"abc-1","type":"vmware-esxi","name":"lab","host":"10.0.0.5","username":"root","datacenterIdentifier":"dc-1","status":"active","createdAt":"2026-10-31T00:00:00Z"}}`))
		default:
			http.NotFound(w, r)
		}
	})
	src, err := api.create(context.Background(), map[string]string{"type": "vmware-esxi", "password": "s3cret"})
	if err != nil {
		t.Fatal(err)
	}
	if src.Identifier != "abc-1" || src.Status != "active" || src.Host != "10.0.0.5" {
		t.Errorf("unexpected source: %+v", src)
	}
	got, err := api.get(context.Background(), "abc-1")
	if err != nil || got == nil || got.Name != "lab" {
		t.Errorf("get = %+v, %v", got, err)
	}
}

func TestGetReturnsNilWhenSourceIsGone(t *testing.T) {
	api := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/v2/migrate/sources/gone" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"error":{"code":"migrationSourceNotFound"},"data":false}`))
			return
		}
		if r.URL.Path == "/apps/v2/migrate/sources" {
			_, _ = w.Write([]byte(`{"error":false,"data":{"items":[{"identifier":"other"}],"limit":100,"offset":0}}`))
			return
		}
		http.NotFound(w, r)
	})
	got, err := api.get(context.Background(), "gone")
	if err != nil || got != nil {
		t.Errorf("a source missing from the list must read as gone, got %+v, %v", got, err)
	}
}

func TestGetKeepsTheErrorWhenSourceStillListed(t *testing.T) {
	api := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/apps/v2/migrate/sources/abc" {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		_, _ = w.Write([]byte(`{"error":false,"data":{"items":[{"identifier":"abc"}],"limit":100,"offset":0}}`))
	})
	got, err := api.get(context.Background(), "abc")
	if err == nil || got != nil {
		t.Errorf("a failing read of a listed source must surface the error, got %+v, %v", got, err)
	}
}

func TestRotateAndDelete(t *testing.T) {
	var seen []string
	api := testAPI(t, func(w http.ResponseWriter, r *http.Request) {
		seen = append(seen, r.Method+" "+r.URL.Path)
		_, _ = w.Write([]byte(`{"error":false,"data":{"identifier":"abc-1"}}`))
	})
	if err := api.rotateCredentials(context.Background(), "abc-1", map[string]string{"password": "n"}); err != nil {
		t.Fatal(err)
	}
	if err := api.delete(context.Background(), "abc-1"); err != nil {
		t.Fatal(err)
	}
	want := []string{"PUT /apps/v2/migrate/sources/abc-1/credentials", "DELETE /apps/v2/migrate/sources/abc-1"}
	if strings.Join(seen, "|") != strings.Join(want, "|") {
		t.Errorf("calls = %v, want %v", seen, want)
	}
}
