package files

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

// fakeSmb serves SMB users and the SMB switch of one share.
type fakeSmb struct {
	mu       sync.Mutex
	users    map[string]map[string]any
	bodies   map[string]map[string]any
	smbOn    bool
	shareGet int
	removed  []string
}

func newFakeSmb(t *testing.T) (*fakeSmb, *shareAPI) {
	t.Helper()
	f := &fakeSmb{users: map[string]map[string]any{}, bodies: map[string]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return f, &shareAPI{api: &rawapi.API{Client: client}}
}

func (f *fakeSmb) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.bodies[r.Method+" "+r.URL.Path] = body
	reply := func(code int, data any) {
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": data})
	}
	switch {
	case r.Method == http.MethodPost && r.URL.Path == "/apps/v2/files/smb-users":
		f.users["u-1"] = map[string]any{
			"identifier": "u-1", "username": body["username"], "dcIdentifier": "dc-1", "shareCount": 0,
			"passwordChangedAt": "2026-11-01T00:00:00Z", "createdAt": "2026-11-01T00:00:00Z",
		}
		reply(http.StatusCreated, f.users["u-1"])
	case r.Method == http.MethodGet && r.URL.Path == "/apps/v2/files/smb-users":
		if r.URL.RawQuery != "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		items := []any{}
		for _, u := range f.users {
			items = append(items, u)
		}
		reply(http.StatusOK, items)
	case r.Method == http.MethodPut && r.URL.Path == "/apps/v2/files/smb-users/u-1/password":
		f.users["u-1"]["passwordChangedAt"] = "2026-11-02T00:00:00Z"
		reply(http.StatusOK, f.users["u-1"])
	case r.Method == http.MethodDelete && r.URL.Path == "/apps/v2/files/smb-users/u-1":
		delete(f.users, "u-1")
		f.removed = append(f.removed, "u-1")
		reply(http.StatusOK, map[string]any{"identifier": "u-1", "deleted": true})
	case r.Method == http.MethodPut && r.URL.Path == "/apps/v2/files/shares/sh-1/smb":
		f.smbOn, _ = body["enabled"].(bool)
		reply(http.StatusOK, map[string]any{"identifier": "sh-1", "state": "ready"})
	case r.Method == http.MethodGet && r.URL.Path == "/apps/v2/files/shares/sh-1":
		f.shareGet++
		reply(http.StatusOK, map[string]any{
			"identifier": "sh-1", "state": "ready", "protocols": map[string]any{"nfs": true, "smb": f.smbOn},
		})
	default:
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": true, "message": "refused"})
	}
}

func smbUserSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	(&smbUserResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if d := resp.Schema.ValidateImplementation(context.Background()); d.HasError() {
		t.Fatalf("schema: %v", d)
	}
	return resp
}

func smbUserState(t *testing.T, sc resource.SchemaResponse, m smbUserModel) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: sc.Schema}
	if d := st.Set(context.Background(), m); d.HasError() {
		t.Fatalf("set state: %v", d)
	}
	return st
}

func TestSmbUserLifecycle(t *testing.T) {
	ctx := context.Background()
	fake, api := newFakeSmb(t)
	r := &smbUserResource{api: api}
	sc := smbUserSchema(t)

	plan := smbUserModel{
		Username: types.StringValue("alice"), Password: types.StringValue("Correct-horse-9"),
		DcIdentifier: types.StringValue("dc-1"), VPCID: types.Int64Value(42), ShareIdentifier: types.StringNull(),
		Identifier: types.StringUnknown(), ShareCount: types.Int64Unknown(), PasswordChangedAt: types.StringUnknown(), CreatedAt: types.StringUnknown(),
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: sc.Schema, Raw: smbUserState(t, sc, plan).Raw}}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create: %v", createResp.Diagnostics)
	}
	var created smbUserModel
	createResp.State.Get(ctx, &created)
	body := fake.bodies["POST /apps/v2/files/smb-users"]
	if created.Identifier.ValueString() != "u-1" || body["vpcId"] != float64(42) || body["dcIdentifier"] != "dc-1" || body["password"] != "Correct-horse-9" {
		t.Fatalf("create wrong: state %+v body %v", created, body)
	}
	if _, has := body["shareIdentifier"]; has {
		t.Errorf("shareIdentifier must not be sent with dc and vpc: %v", body)
	}
	if created.Password.ValueString() != "Correct-horse-9" {
		t.Errorf("password must stay in state as configured")
	}

	// Read finds the user in the account's list and keeps the password and the VPC id.
	readResp := resource.ReadResponse{State: smbUserState(t, sc, created)}
	r.Read(ctx, resource.ReadRequest{State: smbUserState(t, sc, created)}, &readResp)
	if readResp.Diagnostics.HasError() {
		t.Fatalf("read: %v", readResp.Diagnostics)
	}
	var read smbUserModel
	readResp.State.Get(ctx, &read)
	if read.VPCID.ValueInt64() != 42 || read.Password.ValueString() != "Correct-horse-9" {
		t.Fatalf("read dropped configured values: %+v", read)
	}

	// A new password goes to the password route and nothing else.
	next := created
	next.Password = types.StringValue("Another-pass-77")
	upResp := resource.UpdateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Update(ctx, resource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: sc.Schema, Raw: smbUserState(t, sc, next).Raw},
		State: smbUserState(t, sc, created),
	}, &upResp)
	if upResp.Diagnostics.HasError() {
		t.Fatalf("update: %v", upResp.Diagnostics)
	}
	if fake.bodies["PUT /apps/v2/files/smb-users/u-1/password"]["password"] != "Another-pass-77" {
		t.Errorf("password body = %v", fake.bodies)
	}
	var updated smbUserModel
	upResp.State.Get(ctx, &updated)
	if updated.PasswordChangedAt.ValueString() != "2026-11-02T00:00:00Z" || updated.Password.ValueString() != "Another-pass-77" {
		t.Errorf("update state wrong: %+v", updated)
	}

	delResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: smbUserState(t, sc, updated)}, &delResp)
	if delResp.Diagnostics.HasError() || len(fake.removed) != 1 {
		t.Fatalf("delete: %v removed %v", delResp.Diagnostics, fake.removed)
	}
}

func TestSmbUserCreateByShareSendsNoVpc(t *testing.T) {
	ctx := context.Background()
	fake, api := newFakeSmb(t)
	r := &smbUserResource{api: api}
	sc := smbUserSchema(t)
	plan := smbUserModel{
		Username: types.StringValue("bob"), Password: types.StringValue("Correct-horse-9"),
		DcIdentifier: types.StringUnknown(), VPCID: types.Int64Null(), ShareIdentifier: types.StringValue("sh-1"),
		Identifier: types.StringUnknown(), ShareCount: types.Int64Unknown(), PasswordChangedAt: types.StringUnknown(), CreatedAt: types.StringUnknown(),
	}
	resp := resource.CreateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: sc.Schema, Raw: smbUserState(t, sc, plan).Raw}}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("create: %v", resp.Diagnostics)
	}
	body := fake.bodies["POST /apps/v2/files/smb-users"]
	if body["shareIdentifier"] != "sh-1" {
		t.Errorf("shareIdentifier missing: %v", body)
	}
	if _, has := body["vpcId"]; has {
		t.Errorf("vpcId must not be sent with a share: %v", body)
	}
	if _, has := body["dcIdentifier"]; has {
		t.Errorf("dcIdentifier must not be sent with a share: %v", body)
	}
}

func TestSmbUserReadRemovesAGoneUser(t *testing.T) {
	ctx := context.Background()
	_, api := newFakeSmb(t)
	r := &smbUserResource{api: api}
	sc := smbUserSchema(t)
	st := smbUserState(t, sc, smbUserModel{
		Identifier: types.StringValue("u-gone"), Username: types.StringValue("x"), Password: types.StringValue("Correct-horse-9"),
		DcIdentifier: types.StringValue("dc-1"), VPCID: types.Int64Value(1), ShareIdentifier: types.StringNull(),
		ShareCount: types.Int64Null(), PasswordChangedAt: types.StringNull(), CreatedAt: types.StringNull(),
	})
	resp := resource.ReadResponse{State: st}
	r.Read(ctx, resource.ReadRequest{State: st}, &resp)
	if resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
		t.Fatalf("a user missing from the list must leave the state: %v", resp.Diagnostics)
	}
}

func smbShareSchema(t *testing.T) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	(&smbShareResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if d := resp.Schema.ValidateImplementation(context.Background()); d.HasError() {
		t.Fatalf("schema: %v", d)
	}
	return resp
}

func smbShareState(t *testing.T, sc resource.SchemaResponse, m smbShareModel) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: sc.Schema}
	if d := st.Set(context.Background(), m); d.HasError() {
		t.Fatalf("set state: %v", d)
	}
	return st
}

func TestSmbShareLifecycle(t *testing.T) {
	ctx := context.Background()
	pollInterval, waitTimeout = time.Millisecond, 2*time.Second
	fake, api := newFakeSmb(t)
	r := &smbShareResource{api: api}
	sc := smbShareSchema(t)

	plan := smbShareModel{
		ShareIdentifier: types.StringValue("sh-1"), Enabled: types.BoolValue(true),
		Users: []smbGrantModel{{Identifier: types.StringValue("u-1"), Mode: types.StringValue("rw")}},
	}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: tfsdk.Plan{Schema: sc.Schema, Raw: smbShareState(t, sc, plan).Raw}}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create: %v", createResp.Diagnostics)
	}
	body := fake.bodies["PUT /apps/v2/files/shares/sh-1/smb"]
	users, _ := body["users"].([]any)
	if body["enabled"] != true || len(users) != 1 || fake.shareGet == 0 {
		t.Fatalf("create body %v, share reads %d (must wait for ready)", body, fake.shareGet)
	}

	// Drift: SMB switched off elsewhere shows as enabled=false at the next read.
	fake.smbOn = false
	var created smbShareModel
	createResp.State.Get(ctx, &created)
	readResp := resource.ReadResponse{State: smbShareState(t, sc, created)}
	r.Read(ctx, resource.ReadRequest{State: smbShareState(t, sc, created)}, &readResp)
	var read smbShareModel
	readResp.State.Get(ctx, &read)
	if read.Enabled.ValueBool() || len(read.Users) != 1 {
		t.Fatalf("read must report the switch and keep the configured users: %+v", read)
	}

	// Delete turns SMB off with no users.
	delResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: smbShareState(t, sc, created)}, &delResp)
	if delResp.Diagnostics.HasError() {
		t.Fatalf("delete: %v", delResp.Diagnostics)
	}
	off := fake.bodies["PUT /apps/v2/files/shares/sh-1/smb"]
	offUsers, ok := off["users"].([]any)
	if !ok || off["enabled"] != false || len(offUsers) != 0 {
		t.Errorf("delete body = %v", off)
	}
}

func validateSmbShare(t *testing.T, m smbShareModel) resource.ValidateConfigResponse {
	t.Helper()
	sc := smbShareSchema(t)
	cfg := tfsdk.Config{Schema: sc.Schema, Raw: smbShareState(t, sc, m).Raw}
	var resp resource.ValidateConfigResponse
	(&smbShareResource{}).ValidateConfig(context.Background(), resource.ValidateConfigRequest{Config: cfg}, &resp)
	return resp
}

func TestSmbShareValidateConfig(t *testing.T) {
	user := smbGrantModel{Identifier: types.StringValue("u-1"), Mode: types.StringValue("rw")}
	if r := validateSmbShare(t, smbShareModel{ShareIdentifier: types.StringValue("sh-1"), Enabled: types.BoolValue(true), Users: []smbGrantModel{}}); !r.Diagnostics.HasError() {
		t.Error("an enabled share with no user must be refused")
	}
	if r := validateSmbShare(t, smbShareModel{ShareIdentifier: types.StringValue("sh-1"), Enabled: types.BoolValue(true), Users: []smbGrantModel{user, user}}); !r.Diagnostics.HasError() {
		t.Error("a user granted twice must be refused")
	}
	if r := validateSmbShare(t, smbShareModel{ShareIdentifier: types.StringValue("sh-1"), Enabled: types.BoolValue(false), Users: []smbGrantModel{}}); r.Diagnostics.HasError() {
		t.Errorf("a disabled share needs no user: %v", r.Diagnostics)
	}
	if r := validateSmbShare(t, smbShareModel{ShareIdentifier: types.StringValue("sh-1"), Enabled: types.BoolValue(true), Users: []smbGrantModel{user}}); r.Diagnostics.HasError() {
		t.Errorf("one user is enough: %v", r.Diagnostics)
	}
}

func TestSmbUserValidateConfigNamesOneVPC(t *testing.T) {
	sc := smbUserSchema(t)
	check := func(m smbUserModel) bool {
		cfg := tfsdk.Config{Schema: sc.Schema, Raw: smbUserState(t, sc, m).Raw}
		resp := resource.ValidateConfigResponse{}
		for _, v := range (&smbUserResource{}).ConfigValidators(context.Background()) {
			v.ValidateResource(context.Background(), resource.ValidateConfigRequest{Config: cfg}, &resp)
		}
		return resp.Diagnostics.HasError()
	}
	base := smbUserModel{
		Username: types.StringValue("a-user"), Password: types.StringValue("Correct-horse-9"), Identifier: types.StringNull(),
		ShareCount: types.Int64Null(), PasswordChangedAt: types.StringNull(), CreatedAt: types.StringNull(),
		DcIdentifier: types.StringNull(), VPCID: types.Int64Null(), ShareIdentifier: types.StringNull(),
	}
	none := base
	byShare := base
	byShare.ShareIdentifier = types.StringValue("sh-1")
	byVPC := base
	byVPC.DcIdentifier, byVPC.VPCID = types.StringValue("dc-1"), types.Int64Value(7)
	half := base
	half.DcIdentifier = types.StringValue("dc-1")
	both := byVPC
	both.ShareIdentifier = types.StringValue("sh-1")
	if check(byShare) || check(byVPC) {
		t.Error("a share, or dc and vpc, must be accepted")
	}
	if !check(none) || !check(half) || !check(both) {
		t.Error("no VPC, half a VPC and both forms must be refused")
	}
}
