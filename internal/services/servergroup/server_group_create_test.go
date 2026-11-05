package servergroup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

// newCreateResource serves a group list that already holds a group named "web"
// (listed first); POST /apps/v2/vm/group/new appends a second group with the
// same name and the identifier "grp-new".
func newCreateResource(t *testing.T) *serverGroupResource {
	t.Helper()
	var mu sync.Mutex
	rows := []any{map[string]any{
		"identifier": "grp-old", "group_name": "web", "group_description": "old",
		"is_distributed": false, "created_on": "2026-01-01T00:00:00Z",
	}}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/apps/v2/vm/group/new":
			rows = append(rows, map[string]any{
				"identifier": "grp-new", "group_name": "web", "group_description": "new",
				"is_distributed": true, "created_on": "2026-11-05T00:00:00Z",
			})
			_ = json.NewEncoder(w).Encode(map[string]any{"error": false})
		case r.Method == http.MethodGet && r.URL.Path == "/apps/v2/vm/group":
			_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": map[string]any{
				"total": len(rows), "rows": rows,
			}})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return &serverGroupResource{client: client}
}

// Create must record the group it just created, not the first group that has
// the same name.
func TestServerGroupCreateUsesNewGroupWhenNameExists(t *testing.T) {
	r := newCreateResource(t)
	ctx := context.Background()
	var sc resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sc)
	plan := tfsdk.Plan{Schema: sc.Schema}
	if d := plan.Set(ctx, serverGroupResourceModel{
		Identifier:       types.StringUnknown(),
		GroupName:        types.StringValue("web"),
		GroupDescription: types.StringValue("new"),
		IsDistributed:    types.BoolValue(true),
		CreatedOn:        types.StringUnknown(),
	}); d.HasError() {
		t.Fatalf("set plan: %v", d)
	}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("create: %v", resp.Diagnostics)
	}
	var out serverGroupResourceModel
	if d := resp.State.Get(ctx, &out); d.HasError() {
		t.Fatalf("get state: %v", d)
	}
	if out.Identifier.ValueString() != "grp-new" {
		t.Fatalf("adopted the wrong group: identifier %q, want grp-new", out.Identifier.ValueString())
	}
	if out.CreatedOn.ValueString() != "2026-11-05T00:00:00Z" {
		t.Fatalf("created_on from the wrong group: %q", out.CreatedOn.ValueString())
	}
}
