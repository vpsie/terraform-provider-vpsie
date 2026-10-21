package baremetal

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/vpsie/govpsie"
)

func TestInstallInputReplacement(t *testing.T) {
	ctx := context.Background()
	s := resourceSchema(t)

	applied := appliedModel()
	imported := appliedModel()
	imported.PlanIdentifier = types.StringNull()
	imported.OsIdentifier = types.StringNull()
	imported.SSHKeyIdentifiers = types.SetNull(types.StringType)

	nullState := tfsdk.State{Schema: s, Raw: nullValue(s)}
	nullPlan := tfsdk.Plan{Schema: s, Raw: nullValue(s)}

	tests := []struct {
		name    string
		state   tfsdk.State
		plan    tfsdk.Plan
		stateV  types.String
		planV   types.String
		configV types.String
		want    bool
	}{
		{"create", nullState, toPlan(t, s, plannedModel()), types.StringNull(), types.StringValue("plan-1"), types.StringValue("plan-1"), false},
		{"destroy", toState(t, s, applied), nullPlan, types.StringValue("plan-1"), types.StringNull(), types.StringNull(), false},
		{"unchanged", toState(t, s, applied), toPlan(t, s, applied), types.StringValue("plan-1"), types.StringValue("plan-1"), types.StringValue("plan-1"), false},
		{"changed", toState(t, s, applied), toPlan(t, s, applied), types.StringValue("plan-1"), types.StringValue("plan-2"), types.StringValue("plan-2"), true},
		{"left to the API", toState(t, s, applied), toPlan(t, s, applied), types.StringValue("monthly"), types.StringUnknown(), types.StringNull(), false},
		{"adopted after import", toState(t, s, imported), toPlan(t, s, applied), types.StringNull(), types.StringValue("plan-1"), types.StringValue("plan-1"), false},
		{"set where it was null, not imported", toState(t, s, applied), toPlan(t, s, applied), types.StringNull(), types.StringValue("zone-1"), types.StringValue("zone-1"), true},
		{"removed", toState(t, s, applied), toPlan(t, s, applied), types.StringValue("zone-1"), types.StringNull(), types.StringNull(), true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := planmodifier.StringRequest{
				Path: path.Root("plan_identifier"), State: tc.state, Plan: tc.plan,
				StateValue: tc.stateV, PlanValue: tc.planV, ConfigValue: tc.configV,
			}
			var resp planmodifier.StringResponse
			replaceInstallInput().PlanModifyString(ctx, req, &resp)
			if resp.RequiresReplace != tc.want {
				t.Errorf("RequiresReplace = %t, want %t", resp.RequiresReplace, tc.want)
			}
		})
	}

	t.Run("ssh keys added after import", func(t *testing.T) {
		req := planmodifier.SetRequest{
			State: toState(t, s, imported), Plan: toPlan(t, s, applied),
			StateValue: types.SetNull(types.StringType), PlanValue: stringSet("key-1"), ConfigValue: stringSet("key-1"),
		}
		var resp planmodifier.SetResponse
		replaceInstallInput().PlanModifySet(ctx, req, &resp)
		if resp.RequiresReplace {
			t.Error("adopting SSH keys after an import must not replace the server")
		}
	})

	t.Run("ssh keys changed", func(t *testing.T) {
		req := planmodifier.SetRequest{
			State: toState(t, s, applied), Plan: toPlan(t, s, applied),
			StateValue: stringSet("key-1"), PlanValue: stringSet("key-2"), ConfigValue: stringSet("key-2"),
		}
		var resp planmodifier.SetResponse
		replaceInstallInput().PlanModifySet(ctx, req, &resp)
		if !resp.RequiresReplace {
			t.Error("changing SSH keys must replace the server")
		}
	})

	t.Run("script order changed", func(t *testing.T) {
		req := planmodifier.ListRequest{
			State: toState(t, s, applied), Plan: toPlan(t, s, applied),
			StateValue:  stringList("install-agent", "configure-agent"),
			PlanValue:   stringList("configure-agent", "install-agent"),
			ConfigValue: stringList("configure-agent", "install-agent"),
		}
		var resp planmodifier.ListResponse
		replaceInstallInput().PlanModifyList(ctx, req, &resp)
		if !resp.RequiresReplace {
			t.Error("the scripts run in order at first boot, so a new order must replace the server")
		}
	})
}

func TestPasswordProblem(t *testing.T) {
	tests := map[string]bool{
		"Good-pass1":                   true,
		"short":                        false,
		strings.Repeat("a", 128):       true,
		strings.Repeat("a", 129):       false,
		`has"quote1`:                   false,
		`has\backslash`:                false,
		"has{{braces}}":                false,
		"tab\tinside1":                 false,
		"non-ascii-é1":                 false,
		"single{brace}and spaces ok 1": true,
	}
	for password, valid := range tests {
		if got := passwordProblem(password) == ""; got != valid {
			t.Errorf("passwordProblem(%q) valid = %t, want %t", password, got, valid)
		}
	}
}

func TestHostnamePattern(t *testing.T) {
	tests := map[string]bool{
		"web1":                   true,
		"web1.example.com":       true,
		"a":                      true,
		"-web":                   false,
		"web-":                   false,
		"web_1":                  false,
		"web 1":                  false,
		strings.Repeat("a", 253): true,
		strings.Repeat("a", 254): false,
	}
	for hostname, valid := range tests {
		if got := hostnamePattern.MatchString(hostname); got != valid {
			t.Errorf("hostname %q valid = %t, want %t", hostname, got, valid)
		}
	}
}

func TestUserDataValidator(t *testing.T) {
	tests := map[string]bool{
		"#cloud-config\npackages: [nginx]\n":           true,
		"#!/bin/bash\necho hi\n":                       true,
		"packages: [nginx]\n":                          false,
		"#cloud-config\n" + strings.Repeat("a", 65536): false,
		"#!/bin/sh\x00":                                false,
	}
	for data, valid := range tests {
		var resp validator.StringResponse
		userDataValidator{}.ValidateString(context.Background(), validator.StringRequest{
			Path: path.Root("user_data"), ConfigValue: types.StringValue(data),
		}, &resp)
		if got := !resp.Diagnostics.HasError(); got != valid {
			t.Errorf("user data %.20q valid = %t, want %t", data, got, valid)
		}
	}
}

func TestOrderOutcomeUnknown(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"transport", errors.New("connection reset"), true},
		{"attempt cut off", &url.Error{Op: "Post", URL: "https://api", Err: context.DeadlineExceeded}, true},
		{"gateway", &govpsie.APIError{StatusCode: http.StatusBadGateway}, true},
		{"unavailable", &govpsie.APIError{StatusCode: http.StatusServiceUnavailable}, true},
		{"gateway timeout", &govpsie.APIError{StatusCode: http.StatusGatewayTimeout}, true},
		{"cloudflare 520", &govpsie.APIError{StatusCode: 520}, true},
		{"cloudflare 522", &govpsie.APIError{StatusCode: 522}, true},
		{"cloudflare 524", &govpsie.APIError{StatusCode: 524}, true},
		{"in progress", &govpsie.APIError{StatusCode: http.StatusConflict, Code: codeDeployInProgress}, true},
		{"other conflict", &govpsie.APIError{StatusCode: http.StatusConflict, Code: "bareMetalIdempotencyKeyReused"}, false},
		{"internal", &govpsie.APIError{StatusCode: http.StatusInternalServerError}, false},
		{"invalid", &govpsie.APIError{StatusCode: http.StatusUnprocessableEntity}, false},
	}
	for _, tc := range tests {
		if got := orderOutcomeUnknown(tc.err); got != tc.want {
			t.Errorf("%s: orderOutcomeUnknown = %t, want %t", tc.name, got, tc.want)
		}
	}
}

// deployOrder is the order the deploy tests send.
func deployOrder() *govpsie.BareMetalDeployRequest {
	return &govpsie.BareMetalDeployRequest{DCIdentifier: "dc-1", PlanIdentifier: "plan-1", Hostname: "web1.example.com"}
}

// listJSON is the GET /bare-metal answer with the given servers, each
// "identifier:hostname:dc".
func listJSON(servers ...string) fakeResponse {
	rows := make([]string, 0, len(servers))
	for _, s := range servers {
		f := strings.SplitN(s, ":", 3)
		rows = append(rows, `{"identifier":"`+f[0]+`","machine_name":"`+f[1]+`","datacenter_identifier":"`+f[2]+`","osPhase":"installing"}`)
	}

	return fakeResponse{status: http.StatusOK, body: `{"success":true,"data":[` + strings.Join(rows, ",") + `]}`}
}

func TestDeploy_UnknownOutcomeIsResentUntilTheDeadline(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON("bm-old:web1.example.com:dc-1"))
	api.on(http.MethodPost, deployPath, apiError(524, ""))

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	_, adopted, err := deploy(ctx, client, deployOrder())
	if err == nil || adopted {
		t.Fatalf("want an error, got adopted = %t, err = %v", adopted, err)
	}
	// More attempts than the five of a counted retry: the deadline bounds it.
	if n := len(api.callsTo(http.MethodPost, deployPath)); n <= 5 {
		t.Errorf("deploy calls = %d, want more than 5 within the deadline", n)
	}
	for _, want := range []string{`"web1.example.com"`, "dc-1", "no new server", "terraform import"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error %q does not name %s", err, want)
		}
	}
	// The server that was there before the order is never adopted.
	if n := len(api.callsTo(http.MethodGet, listPath)); n != 2 {
		t.Errorf("list calls = %d, want 2 (before the order and the last lookup)", n)
	}
}

func TestDeploy_UnknownOutcomeAtTheDeadlineAdoptsTheNewServer(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath,
		listJSON("bm-old:web1.example.com:dc-1"),
		listJSON("bm-old:web1.example.com:dc-1", "bm-other:web1.example.com:dc-2", "bm-1:WEB1.example.com:dc-1"))
	api.on(http.MethodPost, deployPath, apiError(524, ""))

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	order, adopted, err := deploy(ctx, client, deployOrder())
	if err != nil || !adopted {
		t.Fatalf("adopted = %t, err = %v", adopted, err)
	}
	if order.Identifier != testServer || order.Status != govpsie.BareMetalOrderSold {
		t.Errorf("order = %+v", order)
	}
}

func TestDeploy_AmbiguousLookupIsNotAdopted(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON(), listJSON("bm-1:web1.example.com:dc-1", "bm-2:web1.example.com:dc-1"))
	api.on(http.MethodPost, deployPath, apiError(http.StatusGatewayTimeout, ""))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, adopted, err := deploy(ctx, client, deployOrder())
	if err == nil || adopted || !strings.Contains(err.Error(), "bm-1, bm-2") {
		t.Fatalf("adopted = %t, err = %v", adopted, err)
	}
}

func TestDeploy_NoAdoptionWithoutTheListBeforeTheOrder(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, apiError(http.StatusInternalServerError, ""), listJSON("bm-1:web1.example.com:dc-1"))
	api.on(http.MethodPost, deployPath, apiError(524, ""))

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	_, adopted, err := deploy(ctx, client, deployOrder())
	if err == nil || adopted || !strings.Contains(err.Error(), "could not be listed before the order") {
		t.Fatalf("adopted = %t, err = %v", adopted, err)
	}
}

func TestDeploy_KeyHeldPastTheAnswerBoundIsLookedUp(t *testing.T) {
	fastWaits(t)
	deployAnswerBound = 20 * time.Millisecond
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON(), listJSON(), listJSON("bm-1:web1.example.com:dc-1"))
	api.on(http.MethodPost, deployPath, apiError(http.StatusConflict, codeDeployInProgress))

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	order, adopted, err := deploy(ctx, client, deployOrder())
	if err != nil || !adopted || order.Identifier != testServer {
		t.Fatalf("order = %+v, adopted = %t, err = %v", order, adopted, err)
	}
	if ctx.Err() != nil {
		t.Error("the order was found only by the last lookup at the deadline, not while the key was held")
	}
}

func TestDeploy_RefusalAfterUnknownOutcomeLooksTheOrderUp(t *testing.T) {
	tests := map[string]struct {
		lookup  fakeResponse
		refusal fakeResponse
		adopted bool
		errHas  []string
	}{
		"one new server is adopted": {
			lookup:  listJSON("bm-old:web1.example.com:dc-1", "bm-other:web1.example.com:dc-2", "bm-1:web1.example.com:dc-1"),
			refusal: apiError(http.StatusInternalServerError, ""),
			adopted: true,
		},
		"no new server": {
			lookup:  listJSON("bm-old:web1.example.com:dc-1"),
			refusal: apiError(http.StatusForbidden, "forbidden"),
			errHas:  []string{`"web1.example.com"`, "earlier attempt whose outcome is unknown", "check the bare-metal servers of the account before applying again", "no new server with that name was found"},
		},
		"two new servers": {
			lookup:  listJSON("bm-old:web1.example.com:dc-1", "bm-1:web1.example.com:dc-1", "bm-2:web1.example.com:dc-1"),
			refusal: apiError(http.StatusTooManyRequests, ""),
			errHas:  []string{`"web1.example.com"`, "bm-1, bm-2", "check the bare-metal servers of the account before applying again"},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			fastWaits(t)
			api, client := newFakeAPI(t)
			api.on(http.MethodGet, listPath, listJSON("bm-old:web1.example.com:dc-1"), tc.lookup)
			api.on(http.MethodPost, deployPath, apiError(524, ""), tc.refusal)

			order, adopted, err := deploy(context.Background(), client, deployOrder())
			if n := len(api.callsTo(http.MethodPost, deployPath)); n != 2 {
				t.Errorf("deploy calls = %d, want 2", n)
			}
			if n := len(api.callsTo(http.MethodGet, listPath)); n != 2 {
				t.Errorf("list calls = %d, want 2 (before the order and one lookup)", n)
			}
			if tc.adopted {
				if err != nil || !adopted || order.Identifier != testServer {
					t.Fatalf("order = %+v, adopted = %t, err = %v", order, adopted, err)
				}
				return
			}
			if err == nil || adopted {
				t.Fatalf("adopted = %t, err = %v", adopted, err)
			}
			for _, want := range tc.errHas {
				if !strings.Contains(err.Error(), want) {
					t.Errorf("err = %v, want it to contain %q", err, want)
				}
			}
			var apiErr *govpsie.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != tc.refusal.status {
				t.Errorf("err = %v, want it to wrap the %d refusal", err, tc.refusal.status)
			}
		})
	}
}

func TestDeploy_RefusalWithoutUnknownOutcomeIsFinal(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, listPath, listJSON(), listJSON("bm-1:web1.example.com:dc-1"))
	api.on(http.MethodPost, deployPath, apiError(http.StatusUnprocessableEntity, "NoMachineAvailable"))

	_, adopted, err := deploy(context.Background(), client, deployOrder())
	var apiErr *govpsie.APIError
	if adopted || !errors.As(err, &apiErr) || apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("adopted = %t, err = %v", adopted, err)
	}
	if strings.Contains(err.Error(), "check the bare-metal servers") {
		t.Errorf("err = %v, want the refusal as it is", err)
	}
	if n := len(api.callsTo(http.MethodPost, deployPath)); n != 1 {
		t.Errorf("deploy calls = %d, want 1", n)
	}
	if n := len(api.callsTo(http.MethodGet, listPath)); n != 1 {
		t.Errorf("list calls = %d, want 1 (no lookup after a plain refusal)", n)
	}
}

func TestScriptIdentifiers_OrderedUniqueList(t *testing.T) {
	attr, ok := resourceSchema(t).Attributes["script_identifiers"].(schema.ListAttribute)
	if !ok {
		t.Fatalf("script_identifiers is %T, want a list: the API runs the scripts in order", resourceSchema(t).Attributes["script_identifiers"])
	}

	tests := map[string]struct {
		ids   []string
		valid bool
	}{
		"ordered":   {[]string{"install-agent", "configure-agent"}, true},
		"duplicate": {[]string{"install-agent", "install-agent"}, false},
		"too many":  {[]string{"a", "b", "c", "d", "e", "f"}, false},
		"pattern":   {[]string{"bad id"}, false},
		"too long":  {[]string{strings.Repeat("a", 65)}, false},
	}
	for name, tc := range tests {
		var resp validator.ListResponse
		for _, v := range attr.ListValidators() {
			v.ValidateList(context.Background(), validator.ListRequest{
				Path: path.Root("script_identifiers"), ConfigValue: stringList(tc.ids...),
			}, &resp)
		}
		if got := !resp.Diagnostics.HasError(); got != tc.valid {
			t.Errorf("%s: valid = %t, want %t (%v)", name, got, tc.valid, resp.Diagnostics)
		}
	}
}

func TestWaitForOS_Bounds(t *testing.T) {
	t.Run("not found", func(t *testing.T) {
		fastWaits(t)
		api, client := newFakeAPI(t)
		api.on(http.MethodGet, testServerPath, apiError(http.StatusNotFound, "not_found"))

		_, err := waitForOS(context.Background(), client, testServer, "the OS install")
		if !errors.Is(err, errServerGone) {
			t.Fatalf("err = %v, want errServerGone", err)
		}
		if n := len(api.callsTo(http.MethodGet, testServerPath)); n != maxConsecutiveNotFound {
			t.Errorf("reads = %d, want %d", n, maxConsecutiveNotFound)
		}
	})

	t.Run("read failures", func(t *testing.T) {
		fastWaits(t)
		api, client := newFakeAPI(t)
		api.on(http.MethodGet, testServerPath, apiError(http.StatusInternalServerError, "internal"))

		if _, err := waitForOS(context.Background(), client, testServer, "the OS install"); err == nil {
			t.Fatal("want an error")
		}
		if n := len(api.callsTo(http.MethodGet, testServerPath)); n != maxConsecutiveReadFailures {
			t.Errorf("reads = %d, want %d", n, maxConsecutiveReadFailures)
		}
	})

	t.Run("suspended", func(t *testing.T) {
		fastWaits(t)
		api, client := newFakeAPI(t)
		api.on(http.MethodGet, testServerPath, ok(strings.Replace(serverJSON("installing", "web1", `[]`), `"ready"`, `"suspended"`, 1)))

		bm, err := waitForOS(context.Background(), client, testServer, "the OS install")
		if err == nil || bm == nil {
			t.Fatalf("err = %v, server = %v: want an error and the last read", err, bm)
		}
	})
}

func TestDetachNetwork_NotAttachedCountsAsDone(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodDelete, testNetworksPath+"/7", apiError(http.StatusNotFound, codeNetworkNotAttached))

	if err := detachNetwork(context.Background(), client, testServer, 7); err != nil {
		t.Fatalf("err = %v", err)
	}
	if n := len(api.callsTo(http.MethodGet, testNetworksPath)); n != 0 {
		t.Errorf("waited on a network that is not attached: %d reads", n)
	}
}

func TestAttachNetwork_FailedAttachReportsCode(t *testing.T) {
	fastWaits(t)
	api, client := newFakeAPI(t)
	api.on(http.MethodPost, testNetworksPath, accepted(`{"networkId":7,"status":"attaching"}`))
	api.on(http.MethodGet, testNetworksPath, ok(networksJSON(`[{"networkId":7,"status":"failed","errorCode":"vlan_exhausted"}]`)))

	err := attachNetwork(context.Background(), client, testServer, 7)
	if err == nil || !strings.Contains(err.Error(), "vlan_exhausted") {
		t.Fatalf("err = %v, want the API error code", err)
	}
}

func dataSourceSchema(t *testing.T, d datasource.DataSource) dsschema.Schema {
	t.Helper()

	var resp datasource.SchemaResponse
	d.Schema(context.Background(), datasource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", resp.Diagnostics)
	}

	return resp.Schema
}

func readDataSource(t *testing.T, d datasource.DataSource, config any) tfsdk.State {
	t.Helper()
	ctx := context.Background()

	s := dataSourceSchema(t, d)
	null := tftypes.NewValue(s.Type().TerraformType(ctx), nil)
	cfg := tfsdk.State{Schema: s, Raw: null}
	if diags := cfg.Set(ctx, config); diags.HasError() {
		t.Fatalf("config: %v", diags)
	}

	resp := &datasource.ReadResponse{State: tfsdk.State{Schema: s, Raw: null}}
	d.Read(ctx, datasource.ReadRequest{Config: tfsdk.Config{Schema: s, Raw: cfg.Raw}}, resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("read: %v", resp.Diagnostics)
	}

	return resp.State
}

func TestPlansDataSource_FiltersAndFlattens(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, "/apps/v2/bare-metal/dc/dc-1/hardware-plans", ok(`[{"identifier":"plan-1","name":"GPU-1","cpuCount":32,"ramGb":128,`+
		`"diskGb":1920,"gpuModel":"L40S","gpuCount":2,"price":"899.00","currency":"USD","setupFee":0,"delivery":"instant",`+
		`"contractTerms":[{"term":"monthly","discountPct":0},{"term":"hourly"}],"linkSpeedOptions":[1000,10000],"availableServers":3,"isActive":1}]`))

	d := &bareMetalPlansDataSource{client: client}
	state := readDataSource(t, d, &bareMetalPlansDataSourceModel{
		ID: types.StringNull(), DcIdentifier: types.StringValue("dc-1"),
		MinRAMGB: types.Int64Value(64), MinDiskGB: types.Int64Null(), GPU: types.BoolValue(true),
	})

	calls := api.callsTo(http.MethodGet, "/apps/v2/bare-metal/dc/dc-1/hardware-plans")
	if len(calls) != 1 {
		t.Fatalf("calls = %d", len(calls))
	}
	if q, _ := url.ParseQuery(calls[0].query); q.Get("minRamGb") != "64" || q.Get("gpu") != "true" || q.Has("minDiskGb") {
		t.Errorf("query = %q", calls[0].query)
	}

	var got bareMetalPlansDataSourceModel
	if diags := state.Get(context.Background(), &got); diags.HasError() {
		t.Fatalf("state: %v", diags)
	}
	if len(got.Plans) != 1 {
		t.Fatalf("plans = %d", len(got.Plans))
	}
	p := got.Plans[0]
	if p.Identifier.ValueString() != "plan-1" || p.GPUCount.ValueInt64() != 2 || p.Price.ValueString() != "899.00" || p.SetupFee.ValueString() != "0" {
		t.Errorf("plan = %+v", p)
	}
	if len(p.ContractTerms) != 2 || p.ContractTerms[1].ValueString() != "hourly" || len(p.LinkSpeedOptions) != 2 {
		t.Errorf("terms = %v, speeds = %v", p.ContractTerms, p.LinkSpeedOptions)
	}
	if got.ID.ValueString() != "dc-1" {
		t.Errorf("id = %q", got.ID.ValueString())
	}
}

func TestOSDataSource_FlattensVersions(t *testing.T) {
	api, client := newFakeAPI(t)
	api.on(http.MethodGet, "/apps/v2/bare-metal/dc/dc-1/os", ok(`[`+
		`{"name":"Ubuntu","is_default":1,"available_versions":[{"identifier":"bmos_u2404","fullname":"Ubuntu 24.04 LTS","osFamily":"linux","is_default":1},`+
		`{"identifier":"bmos_u2404_gpu","fullname":"Ubuntu 24.04 LTS (GPU)","osFamily":"linux","gpu":true,"is_default":0}]},`+
		`{"name":"Windows Server","is_default":0,"available_versions":[{"identifier":"bmos_w2022","fullname":"Windows Server 2022","osFamily":"windows","remoteAccessConfigurable":true,"is_default":0}]}]`))

	d := &bareMetalOSDataSource{client: client}
	state := readDataSource(t, d, &bareMetalOSDataSourceModel{
		ID: types.StringNull(), DcIdentifier: types.StringValue("dc-1"), PlanIdentifier: types.StringValue("plan-1"),
	})

	if calls := api.callsTo(http.MethodGet, "/apps/v2/bare-metal/dc/dc-1/os"); len(calls) != 1 || calls[0].query != "planId=plan-1" {
		t.Fatalf("calls = %v", calls)
	}

	var got bareMetalOSDataSourceModel
	if diags := state.Get(context.Background(), &got); diags.HasError() {
		t.Fatalf("state: %v", diags)
	}
	if len(got.OperatingSystems) != 3 {
		t.Fatalf("operating systems = %d, want 3", len(got.OperatingSystems))
	}
	def, gpu, win := got.OperatingSystems[0], got.OperatingSystems[1], got.OperatingSystems[2]
	if !def.IsDefault.ValueBool() || gpu.IsDefault.ValueBool() || win.IsDefault.ValueBool() {
		t.Errorf("is_default = %s, %s, %s; want true, false, false", def.IsDefault, gpu.IsDefault, win.IsDefault)
	}
	if gpu.Identifier.ValueString() != "bmos_u2404_gpu" || !gpu.GPU.ValueBool() || gpu.Distribution.ValueString() != "Ubuntu" {
		t.Errorf("gpu entry = %+v", gpu)
	}
	if win.OSFamily.ValueString() != "windows" || !win.RemoteAccessConfigurable.ValueBool() || win.Distribution.ValueString() != "Windows Server" {
		t.Errorf("windows entry = %+v", win)
	}
	if got.ID.ValueString() != "dc-1/plan-1" {
		t.Errorf("id = %q", got.ID.ValueString())
	}
}
