package files

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

// fakeFiles serves one share. A change leaves the share in a transitional
// state for `busy` further reads, as the platform's reconciler does.
type fakeFiles struct {
	mu      sync.Mutex
	share   map[string]any
	busy    int
	bodies  map[string]map[string]any
	deleted bool
}

func newFake(t *testing.T) (*fakeFiles, *shareAPI) {
	t.Helper()
	pollInterval, waitTimeout = time.Millisecond, 2*time.Second
	f := &fakeFiles{bodies: map[string]map[string]any{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	return f, &shareAPI{api: &rawapi.API{Client: client}}
}

func (f *fakeFiles) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var body map[string]any
	_ = json.NewDecoder(r.Body).Decode(&body)
	f.bodies[r.Method] = body
	reply := func(code int, data any) {
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": data})
	}
	fail := func(code int) {
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(map[string]any{"error": true, "message": "refused"})
	}
	path := strings.TrimPrefix(r.URL.Path, "/apps/v2/files/shares")
	switch {
	case r.Method == http.MethodPost && path == "":
		f.share = map[string]any{
			"identifier": "sh-1", "name": body["name"], "dcIdentifier": body["dcIdentifier"], "sizeGb": body["sizeGb"],
			"state": "creating", "createdAt": "2026-11-01T00:00:00Z", "access": body["access"],
		}
		f.busy = 2
		reply(http.StatusAccepted, f.share)
	case r.Method == http.MethodGet && path == "":
		items := []any{}
		if f.share != nil && !f.deleted {
			items = append(items, f.share)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": items, "total": len(items)})
	case r.Method == http.MethodGet && path == "/sh-1":
		if f.share == nil || f.deleted {
			fail(http.StatusNotFound)
			return
		}
		if f.busy > 0 {
			f.busy--
		} else if f.share["state"] != "ready" && f.share["state"] != "deleting" {
			f.share["state"] = "ready"
			f.share["gateway"] = map[string]any{"identifier": "gw", "vpcIp": "10.0.0.9", "state": "ready"}
			f.share["mount"] = map[string]any{"nfs4": "mount -t nfs 10.0.0.9:/files/data /mnt/data"}
		} else if f.share["state"] == "deleting" {
			f.deleted = true
			fail(http.StatusNotFound)
			return
		}
		reply(http.StatusOK, f.share)
	case r.Method == http.MethodPatch && path == "/sh-1":
		f.share["sizeGb"], f.share["state"], f.busy = body["sizeGb"], "resizing", 1
		reply(http.StatusAccepted, f.share)
	case r.Method == http.MethodPut && path == "/sh-1/access":
		f.share["access"] = body["access"]
		reply(http.StatusOK, f.share)
	case r.Method == http.MethodDelete && path == "/sh-1":
		if body["confirmName"] != f.share["name"] {
			fail(http.StatusBadRequest)
			return
		}
		f.share["state"], f.busy = "deleting", 1
		reply(http.StatusAccepted, f.share)
	default:
		fail(http.StatusNotFound)
	}
}

func schemaOf(t *testing.T) resource.SchemaResponse {
	t.Helper()
	var resp resource.SchemaResponse
	(&shareResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if d := resp.Schema.ValidateImplementation(context.Background()); d.HasError() {
		t.Fatalf("schema: %v", d)
	}
	return resp
}

func stateOf(t *testing.T, sc resource.SchemaResponse, m shareModel) tfsdk.State {
	t.Helper()
	st := tfsdk.State{Schema: sc.Schema}
	if d := st.Set(context.Background(), m); d.HasError() {
		t.Fatalf("set state: %v", d)
	}
	return st
}

func planModel() shareModel {
	return shareModel{
		Name: types.StringValue("data"), DcIdentifier: types.StringValue("dc-1"), VPCID: types.Int64Value(42), SizeGb: types.Int64Value(100),
		Access: []accessModel{
			{CIDR: types.StringValue("10.0.0.0/24"), VMIdentifier: types.StringNull(), Mode: types.StringValue("rw")},
			{VMIdentifier: types.StringValue("vm-1"), CIDR: types.StringNull(), Mode: types.StringValue("ro")},
		},
		Identifier: types.StringUnknown(), State: types.StringUnknown(), GatewayIP: types.StringUnknown(),
		MountNFS4: types.StringUnknown(), CreatedAt: types.StringUnknown(),
	}
}

func TestShareLifecycle(t *testing.T) {
	ctx := context.Background()
	fake, api := newFake(t)
	r := &shareResource{api: api}
	sc := schemaOf(t)

	plan := tfsdk.Plan{Schema: sc.Schema, Raw: stateOf(t, sc, planModel()).Raw}
	createResp := resource.CreateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Create(ctx, resource.CreateRequest{Plan: plan}, &createResp)
	if createResp.Diagnostics.HasError() {
		t.Fatalf("create: %v", createResp.Diagnostics)
	}
	var created shareModel
	createResp.State.Get(ctx, &created)
	if created.State.ValueString() != "ready" || created.GatewayIP.ValueString() != "10.0.0.9" || created.VPCID.ValueInt64() != 42 {
		t.Fatalf("create must wait for ready and keep vpc_id: %+v", created)
	}
	if got := fake.bodies[http.MethodPost]["vpcId"]; got != float64(42) {
		t.Errorf("vpcId sent = %v", got)
	}

	// Grow and change access in one apply.
	next := created
	next.SizeGb = types.Int64Value(200)
	next.Access = created.Access[:1]
	upResp := resource.UpdateResponse{State: tfsdk.State{Schema: sc.Schema}}
	r.Update(ctx, resource.UpdateRequest{
		Plan:  tfsdk.Plan{Schema: sc.Schema, Raw: stateOf(t, sc, next).Raw},
		State: stateOf(t, sc, created),
	}, &upResp)
	if upResp.Diagnostics.HasError() {
		t.Fatalf("update: %v", upResp.Diagnostics)
	}
	var updated shareModel
	upResp.State.Get(ctx, &updated)
	if updated.SizeGb.ValueInt64() != 200 || updated.State.ValueString() != "ready" || len(updated.Access) != 1 {
		t.Fatalf("updated state wrong: %+v", updated)
	}
	if fake.bodies[http.MethodPatch]["sizeGb"] != float64(200) {
		t.Errorf("resize body = %v", fake.bodies[http.MethodPatch])
	}

	// Delete confirms with the name and waits until the share is gone.
	delResp := resource.DeleteResponse{}
	r.Delete(ctx, resource.DeleteRequest{State: stateOf(t, sc, updated)}, &delResp)
	if delResp.Diagnostics.HasError() {
		t.Fatalf("delete: %v", delResp.Diagnostics)
	}
	if fake.bodies[http.MethodDelete]["confirmName"] != "data" || !fake.deleted {
		t.Errorf("delete body %v, deleted %v", fake.bodies[http.MethodDelete], fake.deleted)
	}
}

func TestWaitReadyStopsOnErrorState(t *testing.T) {
	_, api := newFake(t)
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"error": false, "data": map[string]any{"identifier": "sh-1", "state": "error", "errorCode": "FILE_SHARE_FAILED"}})
	}))
	t.Cleanup(srv.Close)
	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL + "/"); err != nil {
		t.Fatal(err)
	}
	api.api = &rawapi.API{Client: client}
	_, err := api.waitReady(context.Background(), "sh-1")
	if err == nil || !strings.Contains(err.Error(), "FILE_SHARE_FAILED") {
		t.Fatalf("a share in error must end the wait with its code, got %v", err)
	}
}

func TestOrderLikeKeepsConfiguredOrder(t *testing.T) {
	a := accessModel{CIDR: types.StringValue("10.0.0.0/24"), VMIdentifier: types.StringNull(), Mode: types.StringValue("rw")}
	b := accessModel{VMIdentifier: types.StringValue("vm-1"), CIDR: types.StringNull(), Mode: types.StringValue("ro")}
	got := orderLike([]accessModel{a, b}, []accessModel{b, a})
	if !got[0].CIDR.Equal(a.CIDR) || !got[1].VMIdentifier.Equal(b.VMIdentifier) {
		t.Errorf("order not kept: %+v", got)
	}
	extra := accessModel{CIDR: types.StringValue("192.168.0.0/16"), VMIdentifier: types.StringNull(), Mode: types.StringValue("rw")}
	if got := orderLike([]accessModel{a}, []accessModel{extra, a}); len(got) != 2 || !got[1].CIDR.Equal(extra.CIDR) {
		t.Errorf("an entry the configuration lacks must follow: %+v", got)
	}
}
