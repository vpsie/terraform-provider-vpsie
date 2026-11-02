package files

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

func TestSmbUsersDataSourceListsUsersWithoutSecrets(t *testing.T) {
	f, api := newFakeSmb(t)
	f.users["u-1"] = map[string]any{
		"identifier": "u-1", "username": "alice", "dcIdentifier": "dc-1", "shareCount": 2,
		"passwordChangedAt": "2026-11-01T00:00:00Z", "createdAt": "2026-10-30T00:00:00Z",
	}
	f.users["u-2"] = map[string]any{
		"identifier": "u-2", "username": "bob", "dcIdentifier": "dc-1", "shareCount": nil,
		"passwordChangedAt": "2026-11-01T00:00:00Z", "createdAt": "2026-10-31T00:00:00Z",
	}
	ds := &smbUsersDataSource{api: api}

	var sc datasource.SchemaResponse
	ds.Schema(context.Background(), datasource.SchemaRequest{}, &sc)
	raw := tfsdk.State{Schema: sc.Schema}
	if diags := raw.Set(context.Background(), smbUsersDataSourceModel{}); diags.HasError() {
		t.Fatal(diags)
	}
	resp := datasource.ReadResponse{State: tfsdk.State{Schema: sc.Schema, Raw: raw.Raw}}
	ds.Read(context.Background(), datasource.ReadRequest{Config: tfsdk.Config{Schema: sc.Schema, Raw: raw.Raw}}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatal(resp.Diagnostics)
	}
	var out smbUsersDataSourceModel
	if diags := resp.State.Get(context.Background(), &out); diags.HasError() {
		t.Fatal(diags)
	}
	if len(out.Users) != 2 {
		t.Fatalf("want 2 users, got %d", len(out.Users))
	}
	byName := map[string]smbUserListModel{}
	for _, u := range out.Users {
		byName[u.Username.ValueString()] = u
	}
	if byName["alice"].ShareCount.ValueInt64() != 2 || !byName["bob"].ShareCount.IsNull() {
		t.Fatalf("share counts wrong: %+v", byName)
	}
}
