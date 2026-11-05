package servergroup

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

// newGroupResource serves GET /apps/v2/vm/group with one group and returns a
// resource wired to it.
func newGroupResource(t *testing.T) *serverGroupResource {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet || r.URL.Path != "/apps/v2/vm/group" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": map[string]any{
			"total": 1,
			"rows": []any{map[string]any{
				"identifier": "grp-1", "group_name": "web", "group_description": "web tier",
				"is_distributed": true, "created_on": "2026-11-05T00:00:00Z",
			}},
		}})
	}))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return &serverGroupResource{client: client}
}

func readGroup(t *testing.T, r *serverGroupResource, in serverGroupResourceModel) serverGroupResourceModel {
	t.Helper()
	ctx := context.Background()
	var sc resource.SchemaResponse
	r.Schema(ctx, resource.SchemaRequest{}, &sc)
	st := tfsdk.State{Schema: sc.Schema}
	if d := st.Set(ctx, in); d.HasError() {
		t.Fatalf("set state: %v", d)
	}
	resp := resource.ReadResponse{State: st}
	r.Read(ctx, resource.ReadRequest{State: st}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read: %v", resp.Diagnostics)
	}
	var out serverGroupResourceModel
	if d := resp.State.Get(ctx, &out); d.HasError() {
		t.Fatalf("get state: %v", d)
	}
	return out
}

// After `terraform import` only the identifier is in the state. Read must fill
// in the configurable attributes, otherwise the next plan sees them as unset
// and replaces (destroys) the imported group.
func TestServerGroupReadFillsAttributesAfterImport(t *testing.T) {
	r := newGroupResource(t)
	out := readGroup(t, r, serverGroupResourceModel{
		Identifier:       types.StringValue("grp-1"),
		GroupName:        types.StringNull(),
		GroupDescription: types.StringNull(),
		IsDistributed:    types.BoolNull(),
		CreatedOn:        types.StringNull(),
	})
	if out.GroupName.ValueString() != "web" || out.GroupDescription.ValueString() != "web tier" ||
		out.IsDistributed.IsNull() || !out.IsDistributed.ValueBool() || out.CreatedOn.ValueString() != "2026-11-05T00:00:00Z" {
		t.Fatalf("imported state not filled from the API: %+v", out)
	}
}

// A managed group keeps its configured values on refresh: the attributes are
// RequiresReplace, so a normalised API value must not cause a replacement.
func TestServerGroupReadKeepsConfiguredValues(t *testing.T) {
	r := newGroupResource(t)
	out := readGroup(t, r, serverGroupResourceModel{
		Identifier:       types.StringValue("grp-1"),
		GroupName:        types.StringValue("web"),
		GroupDescription: types.StringValue("Web Tier"),
		IsDistributed:    types.BoolValue(true),
		CreatedOn:        types.StringValue("old"),
	})
	if out.GroupDescription.ValueString() != "Web Tier" {
		t.Fatalf("configured description overwritten: %q", out.GroupDescription.ValueString())
	}
	if out.CreatedOn.ValueString() != "2026-11-05T00:00:00Z" {
		t.Fatalf("created_on not refreshed: %q", out.CreatedOn.ValueString())
	}
}
