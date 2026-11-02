package protect

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestProtectedServersDataSourceListsAndFilters(t *testing.T) {
	f, api := newFakeProtect(t)
	f.vms["a"] = map[string]any{
		"identifier": "a", "name": "web-1", "tier": "backup", "state": "protected", "pausedBy": nil,
		"policy": map[string]any{"identifier": "pol-1", "name": "nightly"},
		"source": map[string]any{"identifier": "src-1", "name": "vcenter"}, "createdAt": "2026-11-01T00:00:00Z",
	}
	f.vms["b"] = map[string]any{
		"identifier": "b", "name": "db-1", "tier": "dr", "state": "paused", "pausedBy": "tenant",
		"policy": map[string]any{"identifier": "pol-2", "name": "dr"},
		"source": map[string]any{"identifier": "src-1", "name": "vcenter"}, "createdAt": "2026-11-01T01:00:00Z",
	}
	ds := &protectedServersDataSource{api: api}

	var sc datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &sc)
	read := func(state types.String) []protectedServerListModel {
		t.Helper()
		cfg := tfsdk.Config{Schema: sc.Schema, Raw: tfsdk.State{Schema: sc.Schema}.Raw}
		cfgState := tfsdk.State{Schema: sc.Schema, Raw: cfg.Raw}
		if diags := cfgState.Set(context.Background(), protectedServersDataSourceModel{State: state, Servers: nil}); diags.HasError() {
			t.Fatal(diags)
		}
		resp := datasource.ReadResponse{State: tfsdk.State{Schema: sc.Schema, Raw: cfgState.Raw}}
		ds.Read(context.Background(), datasource.ReadRequest{Config: tfsdk.Config{Schema: sc.Schema, Raw: cfgState.Raw}}, &resp)
		if resp.Diagnostics.HasError() {
			t.Fatal(resp.Diagnostics)
		}
		var out protectedServersDataSourceModel
		if diags := resp.State.Get(context.Background(), &out); diags.HasError() {
			t.Fatal(diags)
		}
		return out.Servers
	}

	if all := read(types.StringNull()); len(all) != 2 {
		t.Fatalf("want 2 servers, got %d", len(all))
	}
	paused := read(types.StringValue("paused"))
	if len(paused) != 1 || paused[0].Identifier.ValueString() != "b" || paused[0].PausedBy.ValueString() != "tenant" ||
		paused[0].PolicyIdentifier.ValueString() != "pol-2" || paused[0].SourceIdentifier.ValueString() != "src-1" {
		t.Fatalf("unexpected paused list: %+v", paused)
	}
}
