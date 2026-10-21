package baremetal

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

const deployPath = "/apps/v2/bare-metal/deploy"

func deployJSON(status string, deliveryHours int) string {
	return `{"success":true,"identifier":"bm-1","osPhase":"installing","ipAddress":"203.0.113.10","status":"` + status +
		`","deliveryHours":` + jsonInt(deliveryHours) + `,"contractTerm":"monthly","linkSpeedMbps":1000}`
}

func jsonInt(n int) string {
	b, _ := json.Marshal(n)
	return string(b)
}

func runCreate(t *testing.T, r *bareMetalResource, plan bareMetalResourceModel) *resource.CreateResponse {
	t.Helper()

	s := resourceSchema(t)
	resp := &resource.CreateResponse{State: tfsdk.State{Schema: s, Raw: nullValue(s)}}
	r.Create(context.Background(), resource.CreateRequest{Plan: toPlan(t, s, plan)}, resp)

	return resp
}

func runRead(t *testing.T, r *bareMetalResource, state bareMetalResourceModel) *resource.ReadResponse {
	t.Helper()

	st := toState(t, resourceSchema(t), state)
	resp := &resource.ReadResponse{State: st}
	r.Read(context.Background(), resource.ReadRequest{State: st}, resp)

	return resp
}

func runUpdate(t *testing.T, r *bareMetalResource, state, plan bareMetalResourceModel) *resource.UpdateResponse {
	t.Helper()

	s := resourceSchema(t)
	p := toPlan(t, s, plan)
	resp := &resource.UpdateResponse{State: tfsdk.State{Schema: s, Raw: p.Raw}}
	r.Update(context.Background(), resource.UpdateRequest{Plan: p, State: toState(t, s, state)}, resp)

	return resp
}

func runDelete(t *testing.T, r *bareMetalResource, state bareMetalResourceModel) *resource.DeleteResponse {
	t.Helper()

	st := toState(t, resourceSchema(t), state)
	resp := &resource.DeleteResponse{State: st}
	r.Delete(context.Background(), resource.DeleteRequest{State: st}, resp)

	return resp
}

func requireNoErrors(t *testing.T, diags interface {
	HasError() bool
}, detail any) {
	t.Helper()
	if diags.HasError() {
		t.Fatalf("unexpected errors: %v", detail)
	}
}

func int64s(t *testing.T, set types.Set) []int64 {
	t.Helper()

	ids, diags := setInt64s(context.Background(), set)
	if diags.HasError() {
		t.Fatalf("set: %v", diags)
	}

	return ids
}

func TestCreate_OrdersWaitsForOSAndAttachesNetworks(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON())
	api.on(http.MethodPost, deployPath, ok(deployJSON(govpsie.BareMetalOrderSold, 0)))
	// The server detail lags behind the networks endpoint: it never lists
	// network 7, and the applied value must still equal the planned one.
	api.on(http.MethodGet, testServerPath,
		ok(serverJSON("installing", "web1.example.com", `[]`)),
		ok(serverJSON("running", "web1.example.com", `[]`)))
	api.on(http.MethodPost, testNetworksPath, accepted(`{"networkId":7,"status":"attaching"}`))
	api.on(http.MethodGet, testNetworksPath,
		ok(networksJSON(`[{"networkId":7,"status":"attaching"}]`)),
		ok(networksJSON(`[{"networkId":7,"status":"attached"}]`)))

	resp := runCreate(t, newTestResource(client), plannedModel())
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)
	if !resp.State.Raw.IsFullyKnown() {
		t.Fatal("the saved state has unknown values")
	}

	got := fromState(t, resp.State)
	if got.Identifier.ValueString() != testServer || got.ID.ValueString() != testServer {
		t.Errorf("identifier = %q, id = %q", got.Identifier.ValueString(), got.ID.ValueString())
	}
	if got.OsPhase.ValueString() != govpsie.BareMetalOSPhaseRunning || got.OrderStatus.ValueString() != govpsie.BareMetalOrderSold {
		t.Errorf("os_phase = %q, order_status = %q", got.OsPhase.ValueString(), got.OrderStatus.ValueString())
	}
	if ids := int64s(t, got.PrivateNetworkIDs); len(ids) != 1 || ids[0] != 7 {
		t.Errorf("private_network_ids = %v", ids)
	}
	if got.ContractTerm.ValueString() != "monthly" || got.LinkSpeedMbps.ValueInt64() != 1000 || got.ProjectIdentifier.ValueString() != "proj-1" {
		t.Errorf("contract_term = %s, link_speed_mbps = %s, project = %s", got.ContractTerm, got.LinkSpeedMbps, got.ProjectIdentifier)
	}
	if !got.RemoteAccess.Equal(defaultRemoteAccess()) {
		t.Errorf("remote_access = %s", got.RemoteAccess)
	}

	deploys := api.callsTo(http.MethodPost, deployPath)
	if len(deploys) != 1 {
		t.Fatalf("deploy calls = %d, want 1", len(deploys))
	}
	if key := deploys[0].header.Get("Idempotency-Key"); !strings.HasPrefix(key, idempotencyKeyPrefix) || len(key) != len(idempotencyKeyPrefix)+2*idempotencyKeyRandBytes {
		t.Errorf("Idempotency-Key = %q", key)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(deploys[0].body), &body); err != nil {
		t.Fatalf("deploy body: %v", err)
	}
	if body["planId"] != "plan-1" || body["osIdentifier"] != "bmos_ubuntu2404" || body["hostname"] != "web1.example.com" {
		t.Errorf("deploy body = %v", body)
	}
	if _, sent := body["remoteAccess"]; sent {
		t.Errorf("remoteAccess sent although not configured: %v", body)
	}
	if attaches := api.callsTo(http.MethodPost, testNetworksPath); len(attaches) != 1 || !strings.Contains(attaches[0].body, `"networkId":7`) {
		t.Errorf("attach calls = %v", attaches)
	}
}

func TestCreate_ResendsUnconfirmedOrderWithSameKey(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON())
	// A gateway failure, a Cloudflare timeout, then the first attempt still
	// running for more rounds than a counted retry would allow.
	responses := []fakeResponse{apiError(http.StatusServiceUnavailable, "unavailable"), apiError(524, "")}
	for i := 0; i < 6; i++ {
		responses = append(responses, apiError(http.StatusConflict, codeDeployInProgress))
	}
	api.on(http.MethodPost, deployPath, append(responses, ok(deployJSON(govpsie.BareMetalOrderSold, 0)))...)
	api.on(http.MethodGet, testServerPath, ok(serverJSON("running", "web1.example.com", `[{"networkId":3,"status":"attached"}]`)))

	plan := plannedModel()
	plan.PrivateNetworkIDs = types.SetUnknown(types.Int64Type)
	resp := runCreate(t, newTestResource(client), plan)
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)

	deploys := api.callsTo(http.MethodPost, deployPath)
	if len(deploys) != 9 {
		t.Fatalf("deploy calls = %d, want 9", len(deploys))
	}
	if resp.Diagnostics.WarningsCount() != 0 {
		t.Errorf("a confirmed order reported as adopted: %v", resp.Diagnostics)
	}
	key := deploys[0].header.Get("Idempotency-Key")
	for i, d := range deploys {
		if d.header.Get("Idempotency-Key") != key {
			t.Errorf("attempt %d used key %q, want %q", i+1, d.header.Get("Idempotency-Key"), key)
		}
	}
	if ids := int64s(t, fromState(t, resp.State).PrivateNetworkIDs); len(ids) != 1 || ids[0] != 3 {
		t.Errorf("private_network_ids not taken from the API: %v", ids)
	}
}

func TestCreate_OrderFoundAfterTheCreateTimeoutIsSavedTainted(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON(), listJSON("bm-1:web1.example.com:dc-1"))
	api.on(http.MethodPost, deployPath, apiError(524, ""))
	api.on(http.MethodGet, testServerPath, ok(serverJSON("installing", "web1.example.com", `[]`)))

	plan := plannedModel()
	plan.Timeouts = timeouts.Value{Object: types.ObjectValueMust(map[string]attr.Type{
		"create": types.StringType, "update": types.StringType, "delete": types.StringType,
	}, map[string]attr.Value{
		"create": types.StringValue("100ms"), "update": types.StringNull(), "delete": types.StringNull(),
	})}
	resp := runCreate(t, newTestResource(client), plan)
	if !resp.Diagnostics.HasError() || resp.Diagnostics.WarningsCount() != 1 {
		t.Fatalf("want one error (tainted) and one warning (adopted), got %v", resp.Diagnostics)
	}
	if resp.State.Raw.IsNull() || !resp.State.Raw.IsFullyKnown() {
		t.Fatal("the server found by the lookup must be saved, fully known, so it is tracked rather than lost")
	}
	got := fromState(t, resp.State)
	if got.Identifier.ValueString() != testServer || got.OsPhase.ValueString() != "installing" || got.IPAddress.ValueString() != "203.0.113.10" {
		t.Errorf("identifier = %q, os_phase = %q, ip = %q", got.Identifier.ValueString(), got.OsPhase.ValueString(), got.IPAddress.ValueString())
	}
	if n := len(api.callsTo(http.MethodPost, testNetworksPath)); n != 0 {
		t.Errorf("attached networks after the create timeout: %d calls", n)
	}
}

func TestCreate_ScriptsAreSentInTheConfiguredOrder(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON())
	api.on(http.MethodPost, deployPath, ok(deployJSON(govpsie.BareMetalOrderSold, 0)))
	api.on(http.MethodGet, testServerPath, ok(serverJSON("running", "web1.example.com", `[]`)))
	api.on(http.MethodPost, testNetworksPath, accepted(`{"networkId":7,"status":"attaching"}`))
	api.on(http.MethodGet, testNetworksPath, ok(networksJSON(`[{"networkId":7,"status":"attached"}]`)))

	plan := plannedModel()
	plan.ScriptIdentifiers = stringList("zz-install-agent", "aa-configure-agent")
	resp := runCreate(t, newTestResource(client), plan)
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)

	deploys := api.callsTo(http.MethodPost, deployPath)
	if len(deploys) != 1 || !strings.Contains(deploys[0].body, `"scriptIds":["zz-install-agent","aa-configure-agent"]`) {
		t.Fatalf("deploy body = %v", deploys)
	}
	if got := fromState(t, resp.State).ScriptIdentifiers; !got.Equal(plan.ScriptIdentifiers) {
		t.Errorf("script_identifiers = %s", got)
	}
}

func TestCreate_RefusedOrderIsNotResent(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON())
	api.on(http.MethodPost, deployPath, apiError(http.StatusUnprocessableEntity, "bareMetalOutOfStock"))

	resp := runCreate(t, newTestResource(client), plannedModel())
	if !resp.Diagnostics.HasError() {
		t.Fatal("want an error")
	}
	if n := len(api.callsTo(http.MethodPost, deployPath)); n != 1 {
		t.Errorf("deploy calls = %d, want 1", n)
	}
	if !resp.State.Raw.IsNull() {
		t.Error("state saved for an order that was refused")
	}
}

func TestCreate_FailedInstallKeepsServerInState(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON())
	api.on(http.MethodPost, deployPath, ok(deployJSON(govpsie.BareMetalOrderSold, 0)))
	api.on(http.MethodGet, testServerPath, ok(serverJSON("failed", "web1.example.com", `[]`)))

	resp := runCreate(t, newTestResource(client), plannedModel())
	if !resp.Diagnostics.HasError() {
		t.Fatal("want an error")
	}
	if resp.State.Raw.IsNull() || !resp.State.Raw.IsFullyKnown() {
		t.Fatal("the ordered server must be saved, fully known, so it is tainted rather than lost")
	}
	if got := fromState(t, resp.State); got.Identifier.ValueString() != testServer || got.OsPhase.ValueString() != govpsie.BareMetalOSPhaseFailed {
		t.Errorf("identifier = %q, os_phase = %q", got.Identifier.ValueString(), got.OsPhase.ValueString())
	}
	if n := len(api.callsTo(http.MethodPost, testNetworksPath)); n != 0 {
		t.Errorf("attached networks to a failed server: %d calls", n)
	}
}

func TestCreate_BuildToOrderSavesTheOrder(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON())
	api.on(http.MethodPost, deployPath, ok(deployJSON(govpsie.BareMetalOrderPendingBuild, 72)))

	resp := runCreate(t, newTestResource(client), plannedModel())
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)
	if resp.Diagnostics.WarningsCount() != 1 {
		t.Errorf("warnings = %d, want 1", resp.Diagnostics.WarningsCount())
	}
	if !resp.State.Raw.IsFullyKnown() {
		t.Fatal("the saved state has unknown values")
	}
	got := fromState(t, resp.State)
	if got.OrderStatus.ValueString() != govpsie.BareMetalOrderPendingBuild || got.Status.ValueString() != statusOrdered || got.DeliveryHours.ValueInt64() != 72 {
		t.Errorf("order_status = %q, status = %q, delivery_hours = %d", got.OrderStatus.ValueString(), got.Status.ValueString(), got.DeliveryHours.ValueInt64())
	}
	if n := len(api.callsTo(http.MethodGet, testServerPath)); n != 0 {
		t.Errorf("read an undelivered server %d times", n)
	}
}

func TestRead_NotFoundRemovesResource(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath, apiError(http.StatusNotFound, "not_found"))

	resp := runRead(t, newTestResource(client), appliedModel())
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)
	if !resp.State.Raw.IsNull() {
		t.Error("a server that answers 404 must leave the state")
	}
}

func TestRead_UndeliveredOrderStaysInState(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath, apiError(http.StatusNotFound, "not_found"))

	state := appliedModel()
	state.OrderStatus = types.StringValue(govpsie.BareMetalOrderPendingBuild)
	resp := runRead(t, newTestResource(client), state)
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)
	if resp.State.Raw.IsNull() {
		t.Fatal("an undelivered order was removed from state")
	}
	if resp.Diagnostics.WarningsCount() != 1 {
		t.Errorf("warnings = %d, want 1", resp.Diagnostics.WarningsCount())
	}
}

func TestRead_OtherErrorKeepsState(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath, apiError(http.StatusInternalServerError, "internal"))

	resp := runRead(t, newTestResource(client), appliedModel())
	if !resp.Diagnostics.HasError() {
		t.Fatal("want an error")
	}
	if resp.State.Raw.IsNull() {
		t.Error("a failed read removed the server from state")
	}
}

func TestRead_RefreshesWithoutTouchingInstallInputs(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath, ok(serverJSON("running", "renamed.example.com", `[{"networkId":9,"status":"attached"},{"networkId":4,"status":"detaching"}]`)))

	resp := runRead(t, newTestResource(client), appliedModel())
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)
	got := fromState(t, resp.State)
	if got.Hostname.ValueString() != "renamed.example.com" {
		t.Errorf("hostname = %q", got.Hostname.ValueString())
	}
	if ids := int64s(t, got.PrivateNetworkIDs); len(ids) != 1 || ids[0] != 9 {
		t.Errorf("private_network_ids = %v, want [9]", ids)
	}
	if got.PlanIdentifier.ValueString() != "plan-1" || got.OsIdentifier.ValueString() != "bmos_ubuntu2404" || got.Password.ValueString() != "Initial-pass1" {
		t.Errorf("install inputs changed by a read: plan %s, os %s", got.PlanIdentifier, got.OsIdentifier)
	}
}

func TestImport_AdoptsServerAndFirstApplyNeedsNoReset(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath, ok(serverJSON("running", "web1.example.com", `[{"networkId":7,"status":"attached"}]`)))
	r := newTestResource(client)
	s := resourceSchema(t)

	imp := &resource.ImportStateResponse{State: tfsdk.State{Schema: s, Raw: nullValue(s)}}
	r.ImportState(context.Background(), resource.ImportStateRequest{ID: testServer}, imp)
	requireNoErrors(t, imp.Diagnostics, imp.Diagnostics)
	imported := fromState(t, imp.State)
	if imported.Identifier.ValueString() != testServer || imported.ID.ValueString() != testServer {
		t.Fatalf("imported identifier = %q, id = %q", imported.Identifier.ValueString(), imported.ID.ValueString())
	}

	read := &resource.ReadResponse{State: imp.State}
	r.Read(context.Background(), resource.ReadRequest{State: imp.State}, read)
	requireNoErrors(t, read.Diagnostics, read.Diagnostics)
	adopted := fromState(t, read.State)
	if adopted.DcIdentifier.ValueString() != "dc-1" || adopted.Hostname.ValueString() != "web1.example.com" || adopted.ContractTerm.ValueString() != "monthly" {
		t.Errorf("read after import: dc %s, hostname %s, contract %s", adopted.DcIdentifier, adopted.Hostname, adopted.ContractTerm)
	}
	if !adopted.PlanIdentifier.IsNull() || !adopted.Password.IsNull() {
		t.Errorf("plan_identifier and password must stay null after import")
	}
	if !adoptedByImport(context.Background(), read.State) {
		t.Error("an imported server is not recognised as adopted")
	}

	// First apply: the configuration supplies the inputs the API does not
	// return; the password is adopted without a reset.
	plan := appliedModel()
	plan.Password = types.StringValue("Configured-pass1")
	resp := runUpdate(t, r, adopted, plan)
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)
	if n := len(api.callsTo(http.MethodPost, testServerPath+"/reset-password")); n != 0 {
		t.Errorf("reset the password of an imported server: %d calls", n)
	}
	if got := fromState(t, resp.State); got.Password.ValueString() != "Configured-pass1" || got.PlanIdentifier.ValueString() != "plan-1" {
		t.Errorf("first apply did not adopt the configuration: plan %s", got.PlanIdentifier)
	}
}

func TestUpdate_PasswordResetWaitsForTheServer(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodPost, testServerPath+"/reset-password",
		apiError(http.StatusConflict, codeBusy),
		ok(`{"success":true,"started":true,"osPhase":"password_reset"}`))
	api.on(http.MethodGet, testServerPath,
		ok(serverJSON("password_reset", "web1.example.com", `[{"networkId":7,"status":"attached"}]`)),
		ok(serverJSON("running", "web1.example.com", `[{"networkId":7,"status":"attached"}]`)))

	state := appliedModel()
	plan := appliedModel()
	plan.Password = types.StringValue("Rotated-pass2")
	resp := runUpdate(t, newTestResource(client), state, plan)
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)

	resets := api.callsTo(http.MethodPost, testServerPath+"/reset-password")
	if len(resets) != 2 || !strings.Contains(resets[1].body, `"password":"Rotated-pass2"`) {
		t.Fatalf("reset calls = %v", resets)
	}
	if n := len(api.callsTo(http.MethodGet, testServerPath)); n < 2 {
		t.Errorf("did not wait for the reset: %d reads", n)
	}
	if got := fromState(t, resp.State); got.Password.ValueString() != "Rotated-pass2" {
		t.Error("new password not saved")
	}
}

func TestUpdate_FailedResetKeepsOldPassword(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodPost, testServerPath+"/reset-password", ok(`{"success":true,"started":true,"osPhase":"password_reset"}`))
	api.on(http.MethodGet, testServerPath, ok(serverJSON("failed", "web1.example.com", `[{"networkId":7,"status":"attached"}]`)))

	plan := appliedModel()
	plan.Password = types.StringValue("Rotated-pass2")
	resp := runUpdate(t, newTestResource(client), appliedModel(), plan)
	if !resp.Diagnostics.HasError() {
		t.Fatal("want an error")
	}
	if got := fromState(t, resp.State); got.Password.ValueString() != "Initial-pass1" {
		t.Errorf("password in state = %q, want the old one so the next apply retries", got.Password.ValueString())
	}
}

func TestUpdate_NetworksDetachBeforeAttach(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodDelete, testNetworksPath+"/7", accepted(`{}`))
	api.on(http.MethodPost, testNetworksPath, accepted(`{"networkId":9,"status":"attaching"}`))
	api.on(http.MethodGet, testNetworksPath,
		ok(networksJSON(`[{"networkId":7,"status":"detaching"}]`)),
		ok(networksJSON(`[]`)),
		ok(networksJSON(`[{"networkId":9,"status":"attached"}]`)))
	api.on(http.MethodGet, testServerPath, ok(serverJSON("running", "web1.example.com", `[{"networkId":9,"status":"attached"}]`)))

	plan := appliedModel()
	plan.PrivateNetworkIDs = int64Set([]int64{9})
	resp := runUpdate(t, newTestResource(client), appliedModel(), plan)
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)

	detach, attach := -1, -1
	for i, call := range api.order() {
		switch call {
		case http.MethodDelete + " " + testNetworksPath + "/7":
			detach = i
		case http.MethodPost + " " + testNetworksPath:
			attach = i
		}
	}
	if detach < 0 || attach < 0 || detach > attach {
		t.Errorf("detach at %d, attach at %d: want detach first; calls %v", detach, attach, api.order())
	}
	if ids := int64s(t, fromState(t, resp.State).PrivateNetworkIDs); len(ids) != 1 || ids[0] != 9 {
		t.Errorf("private_network_ids = %v", ids)
	}
}

func TestDelete_ReleasesWithAccountPasswordAndWaits(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath,
		ok(serverJSON("running", "web1.example.com", `[]`)),
		ok(serverJSON("running", "web1.example.com", `[]`)),
		apiError(http.StatusNotFound, "not_found"))
	api.on(http.MethodPost, testServerPath+"/release", apiError(http.StatusConflict, codeBusy), ok(`{}`))

	resp := runDelete(t, newTestResource(client), appliedModel())
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)

	releases := api.callsTo(http.MethodPost, testServerPath+"/release")
	if len(releases) != 2 {
		t.Fatalf("release calls = %d, want 2 (busy, then accepted)", len(releases))
	}
	var body map[string]string
	if err := json.Unmarshal([]byte(releases[1].body), &body); err != nil {
		t.Fatalf("release body: %v", err)
	}
	if body["password"] != testAccountPassword || body["reason"] != releaseReasonDefault {
		t.Errorf("release reason = %q, password is the account password: %t", body["reason"], body["password"] == testAccountPassword)
	}
	if n := len(api.callsTo(http.MethodGet, testServerPath)); n != 3 {
		t.Errorf("reads = %d, want 3 (check, wait, gone)", n)
	}
}

func TestDelete_NeedsAccountPassword(t *testing.T) {
	api, client := newFakeAPI(t)
	client.SetAccountPassword("")

	resp := runDelete(t, newTestResource(client), appliedModel())
	if !resp.Diagnostics.HasError() {
		t.Fatal("want an error")
	}
	if n := len(api.order()); n != 0 {
		t.Errorf("made %d requests without the account password", n)
	}
}

func TestDelete_AlreadyReleased(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath, apiError(http.StatusNotFound, "not_found"))

	resp := runDelete(t, newTestResource(client), appliedModel())
	requireNoErrors(t, resp.Diagnostics, resp.Diagnostics)
	if n := len(api.callsTo(http.MethodPost, testServerPath+"/release")); n != 0 {
		t.Errorf("released a server that is gone: %d calls", n)
	}
}

func TestDelete_UndeliveredOrderIsRefused(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, testServerPath, apiError(http.StatusNotFound, "not_found"))

	state := appliedModel()
	state.OrderStatus = types.StringValue(govpsie.BareMetalOrderPendingBuild)
	resp := runDelete(t, newTestResource(client), state)
	if !resp.Diagnostics.HasError() {
		t.Fatal("an undelivered order must not leave state silently")
	}
	if !strings.Contains(resp.Diagnostics.Errors()[0].Detail(), "terraform state rm") {
		t.Errorf("detail = %q", resp.Diagnostics.Errors()[0].Detail())
	}
}
