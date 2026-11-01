package control

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/vpsie/govpsie"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

func apiFor(t *testing.T, body map[string]any) *rawapi.API {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(body)
	}))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return &rawapi.API{Client: client}
}

func TestFetchPlatformsReadsTheListAndNullableFields(t *testing.T) {
	api := apiFor(t, map[string]any{"error": false, "data": map[string]any{
		"available": true,
		"platforms": []any{map[string]any{"identifier": "p1", "name": "vc", "connectorType": "vmware-vcenter", "region": nil, "vmCount": 12}},
	}})
	available, _, items, err := fetch(context.Background(), api, platformsPath, func(a answer[platform]) []platform { return a.Data.Platforms })
	if err != nil || !available || len(items) != 1 {
		t.Fatalf("available=%v items=%d err=%v", available, len(items), err)
	}
	if items[0].VMCount != 12 || items[0].Region != nil || optString(items[0].Region).IsNull() != true {
		t.Errorf("platform read wrong: %+v", items[0])
	}
}

func TestFetchReportsAnOfferingThatSaysNo(t *testing.T) {
	api := apiFor(t, map[string]any{"error": false, "data": map[string]any{"available": false, "reason": "notLicensed"}})
	available, reason, items, err := fetch(context.Background(), api, collectorsPath, func(a answer[collector]) []collector { return a.Data.Collector })
	if err != nil {
		t.Fatal(err)
	}
	if available || reason != "notLicensed" || len(items) != 0 {
		t.Errorf("available=%v reason=%q items=%d", available, reason, len(items))
	}
}

func TestFetchCollectors(t *testing.T) {
	api := apiFor(t, map[string]any{"error": false, "data": map[string]any{
		"collectors": []any{map[string]any{"identifier": "c1", "name": "edge", "state": "enrolled", "platformCount": 2}},
	}})
	available, _, items, err := fetch(context.Background(), api, collectorsPath, func(a answer[collector]) []collector { return a.Data.Collector })
	if err != nil || !available || len(items) != 1 || items[0].PlatformCount != 2 {
		t.Fatalf("available=%v items=%+v err=%v", available, items, err)
	}
}
