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

// fakePlatform serves /apps/v2/protect/policies from memory and records the
// last body it received per method.
type fakePlatform struct {
	mu       sync.Mutex
	policies map[string]map[string]any
	bodies   map[string]map[string]any
	nextID   int
}

func newFake(t *testing.T) (*fakePlatform, *policyAPI) {
	t.Helper()
	f := &fakePlatform{policies: map[string]map[string]any{}, bodies: map[string]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return f, &policyAPI{api: &rawapi.API{Client: client}}
}

func (f *fakePlatform) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var body map[string]any
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&body)
	}
	f.bodies[r.Method] = body
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
	const base = "/apps/v2/protect/policies"
	id := strings.TrimPrefix(strings.TrimPrefix(r.URL.Path, base), "/")
	switch {
	case r.Method == http.MethodPost && id == "":
		f.nextID++
		pid := fmt.Sprintf("pol-%d", f.nextID)
		p := map[string]any{
			"identifier": pid, "name": body["name"], "tier": body["tier"], "state": "active",
			"datacenter": map[string]any{"identifier": body["datacenter"], "name": "dc"},
			"plan":       map[string]any{"identifier": nil, "name": nil},
			"rpoMinutes": body["rpoMinutes"], "retention": body["retention"],
			"immutabilityDays": 7, "consistency": "app", "verifyMode": "off",
			"rtoTargetMinutes": body["rtoTargetMinutes"], "version": 1, "createdAt": "2026-11-01T00:00:00Z",
		}
		if v, ok := body["immutabilityDays"]; ok {
			p["immutabilityDays"] = v
		}
		f.policies[pid] = p
		reply(http.StatusCreated, p)
	case r.Method == http.MethodGet && id == "":
		items := []any{}
		for _, p := range f.policies {
			items = append(items, p)
		}
		reply(http.StatusOK, map[string]any{"items": items, "total": len(items)})
	case r.Method == http.MethodGet:
		p, ok := f.policies[id]
		if !ok {
			fail(http.StatusNotFound)
			return
		}
		reply(http.StatusOK, p)
	case r.Method == http.MethodPatch:
		p, ok := f.policies[id]
		if !ok {
			fail(http.StatusNotFound)
			return
		}
		if body["version"] != float64(p["version"].(int)) {
			fail(http.StatusConflict)
			return
		}
		for k, v := range body {
			if k != "version" {
				p[k] = v
			}
		}
		p["version"] = p["version"].(int) + 1
		reply(http.StatusOK, p)
	case r.Method == http.MethodDelete:
		p, ok := f.policies[id]
		if !ok || body["confirmName"] != p["name"] {
			fail(http.StatusBadRequest)
			return
		}
		delete(f.policies, id)
		reply(http.StatusOK, map[string]any{"identifier": id, "state": "removed"})
	default:
		fail(http.StatusNotFound)
	}
}

func baseModel() policyModel {
	return policyModel{
		Name: types.StringValue("nightly"), Tier: types.StringValue("backup"), Datacenter: types.StringValue("dc-1"),
		RPOMinutes: types.Int64Value(1440),
		Retention:  &retentionModel{KeepRecentHours: types.Int64Value(24), KeepDaily: types.Int64Value(7), KeepWeekly: types.Int64Value(4)},
		// Unset optionals are unknown in a plan.
		Plan: types.StringUnknown(), ImmutabilityDays: types.Int64Unknown(), Consistency: types.StringUnknown(),
		VerifyMode: types.StringUnknown(), RTOTargetMinutes: types.Int64Null(),
		Identifier: types.StringUnknown(), State: types.StringUnknown(), Version: types.Int64Unknown(), CreatedAt: types.StringUnknown(),
	}
}

func schemaOf(t *testing.T, r resource.Resource) resource.SchemaResponse {
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

func stateOf(t *testing.T, sc resource.SchemaResponse, m policyModel) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: sc.Schema}
	if d := st.Set(context.Background(), m); d.HasError() {
		t.Fatalf("set state: %v", d)
	}
	return st
}

func planOf(t *testing.T, sc resource.SchemaResponse, m policyModel) tfsdk.Plan {
	t.Helper()
	st := stateOf(t, sc, m)
	return tfsdk.Plan{Schema: sc.Schema, Raw: st.Raw}
}

func TestPolicyLifecycle(t *testing.T) {
	ctx := context.Background()
	fake, api := newFake(t)
	r := &policyResource{api: api}
	sc := schemaOf(t, r)

	createResp := resource.CreateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: planOf(t, sc, baseModel())}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create: %v", createResp.Diagnostics)
	}
	var created policyModel
	createResp.State.Get(ctx, &created)
	if created.Identifier.ValueString() != "pol-1" || created.ImmutabilityDays.ValueInt64() != 7 || created.Version.ValueInt64() != 1 {
		t.Fatalf("created state wrong: %+v", created)
	}
	if _, sent := fake.bodies[http.MethodPost]["immutabilityDays"]; sent {
		t.Error("an unset immutability_days must be left out so the platform default applies")
	}
	if _, sent := fake.bodies[http.MethodPost]["plan"]; sent {
		t.Error("an unset plan must be left out")
	}

	// Update: change the RPO; the version read goes with the PATCH.
	next := created
	next.RPOMinutes = types.Int64Value(60)
	next.Version = types.Int64Unknown()
	upResp := resource.UpdateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: planOf(t, sc, next), State: stateOf(t, sc, created)}, &upResp)
	if upResp.Diagnostics.HasError() {
		t.Fatalf("update: %v", upResp.Diagnostics)
	}
	var updated policyModel
	upResp.State.Get(ctx, &updated)
	if updated.RPOMinutes.ValueInt64() != 60 || updated.Version.ValueInt64() != 2 {
		t.Fatalf("updated state wrong: %+v", updated)
	}
	if v, ok := fake.bodies[http.MethodPatch]["rtoTargetMinutes"]; !ok || v != nil {
		t.Errorf("an unset rto_target_minutes must be sent as null to clear it, got %v (present %v)", v, ok)
	}

	// A stale version is refused by the platform and surfaces as an error.
	stale := created // version 1 again
	staleResp := resource.UpdateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Update(ctx, resource.UpdateRequest{Plan: planOf(t, sc, next), State: stateOf(t, sc, stale)}, &staleResp)
	if !staleResp.Diagnostics.HasError() {
		t.Error("an edit on a stale version must fail")
	}

	// Delete confirms with the policy name.
	delResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: stateOf(t, sc, updated)}, &delResp)
	if delResp.Diagnostics.HasError() {
		t.Fatalf("delete: %v", delResp.Diagnostics)
	}
	if got := fake.bodies[http.MethodDelete]["confirmName"]; got != "nightly" {
		t.Errorf("confirmName = %v, want nightly", got)
	}
	if len(fake.policies) != 0 {
		t.Error("the policy must be gone")
	}
}

func TestPolicyReadDropsAGonePolicy(t *testing.T) {
	ctx := context.Background()
	_, api := newFake(t)
	r := &policyResource{api: api}
	sc := schemaOf(t, r)
	m := baseModel()
	m.Identifier = types.StringValue("pol-missing")
	m.Version = types.Int64Value(1)
	m.State = types.StringValue("active")
	m.CreatedAt = types.StringValue("x")
	m.Plan, m.ImmutabilityDays, m.Consistency, m.VerifyMode = types.StringNull(), types.Int64Value(7), types.StringValue("app"), types.StringValue("off")
	resp := resource.ReadResponse{State: stateOf(t, sc, m)}
	r.Read(ctx, resource.ReadRequest{State: stateOf(t, sc, m)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read: %v", resp.Diagnostics)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("a policy the list does not know must be removed from the state")
	}
}

func TestPolicyDeleteTreatsAnAlreadyGonePolicyAsDone(t *testing.T) {
	ctx := context.Background()
	_, api := newFake(t)
	r := &policyResource{api: api}
	sc := schemaOf(t, r)
	m := baseModel()
	m.Identifier = types.StringValue("pol-missing")
	m.Version, m.State, m.CreatedAt = types.Int64Value(1), types.StringValue("active"), types.StringValue("x")
	m.Plan, m.ImmutabilityDays, m.Consistency, m.VerifyMode = types.StringNull(), types.Int64Value(7), types.StringValue("app"), types.StringValue("off")
	resp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: stateOf(t, sc, m)}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("delete of a gone policy must succeed: %v", resp.Diagnostics)
	}
}

func TestPolicyListPagesAndFiltersByState(t *testing.T) {
	var queries []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		queries = append(queries, r.URL.RawQuery)
		items := []map[string]any{}
		if strings.Contains(r.URL.RawQuery, "offset=0") {
			for i := 0; i < rawapi.Page; i++ {
				items = append(items, map[string]any{"identifier": fmt.Sprintf("p%d", i)})
			}
		} else {
			items = append(items, map[string]any{"identifier": "last"})
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": map[string]any{"items": items}})
	}))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	api := &policyAPI{api: &rawapi.API{Client: client}}
	got, err := api.list(context.Background(), "paused")
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != rawapi.Page+1 {
		t.Fatalf("got %d policies, want %d", len(got), rawapi.Page+1)
	}
	want := []string{"state=paused&limit=100&offset=0", "state=paused&limit=100&offset=100"}
	if strings.Join(queries, "|") != strings.Join(want, "|") {
		t.Errorf("queries = %v, want %v", queries, want)
	}
}
