package protect

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

// fakeProtect serves /apps/v2/protect/vms from memory.
type fakeProtect struct {
	mu     sync.Mutex
	vms    map[string]map[string]any
	bodies map[string]map[string]any
	paths  []string
	nextID int
}

func newFakeProtect(t *testing.T) (*fakeProtect, *serverAPI) {
	t.Helper()
	f := &fakeProtect{vms: map[string]map[string]any{}, bodies: map[string]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return f, &serverAPI{api: &rawapi.API{Client: client}}
}

func (f *fakeProtect) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var body map[string]any
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	f.bodies[r.Method] = body
	f.paths = append(f.paths, r.Method+" "+r.URL.Path)
	reply := func(code int, data any) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": data})
	}
	fail := func(code int) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": true, "message": "refused"})
	}
	const base = "/apps/v2/protect/vms"
	rest := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, base), "/")
	id, action, _ := strings.Cut(rest, "/")
	switch {
	case r.Method == http.MethodPost && id == "":
		f.nextID++
		vid := fmt.Sprintf("pvm-%d", f.nextID)
		refs, _ := body["vms"].([]any)
		if len(refs) != 1 {
			fail(http.StatusBadRequest)
			return
		}
		ref, _ := refs[0].(map[string]any)["ref"].(string)
		v := map[string]any{
			"identifier": vid, "name": "vm-" + ref, "tier": "backup", "state": "seeding", "pausedBy": nil,
			"policy":  map[string]any{"identifier": body["policy"], "name": "nightly", "state": "active"},
			"source":  map[string]any{"identifier": body["source"], "name": "vcenter", "type": "vmware"},
			"version": float64(0), "createdAt": "2026-11-02T00:00:00Z",
		}
		f.vms[vid] = v
		reply(http.StatusCreated, map[string]any{"items": []any{v}})
	case r.Method == http.MethodGet && id == "":
		items := []any{}
		for _, v := range f.vms {
			items = append(items, v)
		}
		reply(http.StatusOK, map[string]any{"items": items, "total": len(items)})
	case r.Method == http.MethodGet:
		v, ok := f.vms[id]
		if !ok {
			fail(http.StatusNotFound)
			return
		}
		reply(http.StatusOK, v)
	case r.Method == http.MethodPost && action == "pause":
		v, ok := f.vms[id]
		if !ok {
			fail(http.StatusNotFound)
			return
		}
		v["state"], v["pausedBy"] = "paused", "tenant"
		reply(http.StatusOK, v)
	case r.Method == http.MethodPost && action == "resume":
		v, ok := f.vms[id]
		if !ok {
			fail(http.StatusNotFound)
			return
		}
		v["state"], v["pausedBy"] = "protected", nil
		reply(http.StatusOK, v)
	case r.Method == http.MethodDelete:
		v, ok := f.vms[id]
		if !ok || body["confirmName"] != v["name"] {
			fail(http.StatusBadRequest)
			return
		}
		v["state"] = "removing"
		reply(http.StatusOK, map[string]any{"identifier": id, "state": "removing", "removeAfter": "2026-11-09T00:00:00Z"})
	default:
		fail(http.StatusNotFound)
	}
}

func serverModel() serverModelT {
	return serverModelT{
		Policy: types.StringValue("pol-1"), Source: types.StringValue("src-1"), VMRef: types.StringValue("vm-42"),
		Paused:     types.BoolUnknown(),
		Identifier: types.StringUnknown(), Name: types.StringUnknown(), Tier: types.StringUnknown(),
		State: types.StringUnknown(), PausedBy: types.StringUnknown(), CreatedAt: types.StringUnknown(),
	}
}

func serverSchema(t *testing.T, r resource.Resource) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", resp.Diagnostics)
	}
	if d := resp.Schema.ValidateImplementation(context.Background()); d.HasError() {
		t.Fatalf("schema implementation: %v", d)
	}
	return resp
}

func serverState(t *testing.T, sc resource.SchemaResponse, m serverModelT) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: sc.Schema}
	if d := st.Set(context.Background(), m); d.HasError() {
		t.Fatalf("set state: %v", d)
	}
	return st
}

func serverPlan(t *testing.T, sc resource.SchemaResponse, m serverModelT) tfsdk.Plan {
	return tfsdk.Plan{Schema: sc.Schema, Raw: serverState(t, sc, m).Raw}
}

func TestProtectedServerLifecycle(t *testing.T) {
	ctx := context.Background()
	fake, api := newFakeProtect(t)
	r := &serverResource{api: api}
	sc := serverSchema(t, r)

	createResp := resource.CreateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: serverPlan(t, sc, serverModel())}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create: %v", createResp.Diagnostics)
	}
	var created serverModelT
	createResp.State.Get(ctx, &created)
	if created.Identifier.ValueString() != "pvm-1" || created.Name.ValueString() != "vm-vm-42" || created.Paused.ValueBool() {
		t.Fatalf("created state wrong: %+v", created)
	}
	if fake.bodies[http.MethodPost]["policy"] != "pol-1" || fake.bodies[http.MethodPost]["source"] != "src-1" {
		t.Errorf("add body wrong: %v", fake.bodies[http.MethodPost])
	}

	// paused = true pauses it; paused = false resumes it.
	pausedPlan := created
	pausedPlan.Paused = types.BoolValue(true)
	upResp := resource.UpdateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: serverPlan(t, sc, pausedPlan), State: serverState(t, sc, created)}, &upResp)
	if upResp.Diagnostics.HasError() {
		t.Fatalf("pause: %v", upResp.Diagnostics)
	}
	var paused serverModelT
	upResp.State.Get(ctx, &paused)
	if !paused.Paused.ValueBool() || paused.State.ValueString() != "paused" || paused.PausedBy.ValueString() != "tenant" {
		t.Fatalf("paused state wrong: %+v", paused)
	}
	resumePlan := paused
	resumePlan.Paused = types.BoolValue(false)
	resumeResp := resource.UpdateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: serverPlan(t, sc, resumePlan), State: serverState(t, sc, paused)}, &resumeResp)
	if resumeResp.Diagnostics.HasError() {
		t.Fatalf("resume: %v", resumeResp.Diagnostics)
	}
	var resumed serverModelT
	resumeResp.State.Get(ctx, &resumed)
	if resumed.Paused.ValueBool() || resumed.State.ValueString() != "protected" {
		t.Fatalf("resumed state wrong: %+v", resumed)
	}

	// Delete confirms with the server name read from the state.
	delResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: serverState(t, sc, resumed)}, &delResp)
	if delResp.Diagnostics.HasError() {
		t.Fatalf("delete: %v", delResp.Diagnostics)
	}
	if got := fake.bodies[http.MethodDelete]["confirmName"]; got != "vm-vm-42" {
		t.Errorf("confirmName = %v, want vm-vm-42", got)
	}
}

func TestProtectedServerReadDropsGoneAndRemovingServers(t *testing.T) {
	ctx := context.Background()
	fake, api := newFakeProtect(t)
	r := &serverResource{api: api}
	sc := serverSchema(t, r)
	m := serverModel()
	m.Identifier, m.Name, m.Tier = types.StringValue("pvm-missing"), types.StringValue("x"), types.StringValue("backup")
	m.State, m.PausedBy, m.CreatedAt, m.Paused = types.StringValue("protected"), types.StringNull(), types.StringValue("x"), types.BoolValue(false)

	resp := resource.ReadResponse{State: serverState(t, sc, m)}
	r.Read(ctx, resource.ReadRequest{State: serverState(t, sc, m)}, &resp)
	if resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
		t.Fatalf("a server the list does not know must leave the state: %v", resp.Diagnostics)
	}

	fake.vms["pvm-9"] = map[string]any{"identifier": "pvm-9", "state": "removing", "policy": map[string]any{}, "source": map[string]any{}}
	m.Identifier = types.StringValue("pvm-9")
	resp = resource.ReadResponse{State: serverState(t, sc, m)}
	r.Read(ctx, resource.ReadRequest{State: serverState(t, sc, m)}, &resp)
	if resp.Diagnostics.HasError() || !resp.State.Raw.IsNull() {
		t.Fatalf("a server being removed must leave the state: %v", resp.Diagnostics)
	}
}

func TestProtectedServerDeleteOfGoneServerIsDone(t *testing.T) {
	ctx := context.Background()
	_, api := newFakeProtect(t)
	r := &serverResource{api: api}
	sc := serverSchema(t, r)
	m := serverModel()
	m.Identifier, m.Name, m.Tier = types.StringValue("pvm-missing"), types.StringValue("x"), types.StringValue("backup")
	m.State, m.PausedBy, m.CreatedAt, m.Paused = types.StringValue("protected"), types.StringNull(), types.StringValue("x"), types.BoolValue(false)
	resp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: serverState(t, sc, m)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("delete of a gone server must succeed: %v", resp.Diagnostics)
	}
}

func TestProtectedServerAddFailsWhenPlatformReturnsNoServer(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": map[string]any{"items": []any{}}})
	}))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	api := &serverAPI{api: &rawapi.API{Client: client}}
	if _, err := api.add(context.Background(), "p", "s", "r"); err == nil {
		t.Fatal("an empty add answer must be an error, not a silent success")
	}
}
