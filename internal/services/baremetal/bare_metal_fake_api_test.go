package baremetal

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
	"github.com/vpsie/govpsie"
)

const (
	testServer          = "bm-1"
	listPath            = "/apps/v2/bare-metal"
	testServerPath      = "/apps/v2/bare-metal/bm-1"
	testNetworksPath    = testServerPath + "/private-networks"
	testAccountPassword = "account-secret"
)

// fakeResponse is one canned answer of the fake API.
type fakeResponse struct {
	status int
	body   string
}

func ok(data string) fakeResponse {
	return fakeResponse{status: http.StatusOK, body: `{"error":false,"data":` + data + `}`}
}

func accepted(data string) fakeResponse {
	return fakeResponse{status: http.StatusAccepted, body: `{"error":false,"data":` + data + `}`}
}

func apiError(status int, code string) fakeResponse {
	return fakeResponse{status: status, body: fmt.Sprintf(`{"error":true,"message":"%s","code":"%s"}`, code, code)}
}

// fakeCall is one request the fake API received.
type fakeCall struct {
	method string
	path   string
	query  string
	body   string
	header http.Header
}

// fakeAPI answers each "METHOD /path" with its queued responses in order;
// the last one repeats. A request to a route without responses fails the
// test.
type fakeAPI struct {
	t      *testing.T
	mu     sync.Mutex
	routes map[string][]fakeResponse
	calls  []fakeCall
}

func newFakeAPI(t *testing.T) (*fakeAPI, *govpsie.Client) {
	t.Helper()

	f := &fakeAPI{t: t, routes: map[string][]fakeResponse{}}
	srv := httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(srv.Close)

	client := govpsie.NewClient(srv.Client())
	if err := client.SetBaseURL(srv.URL); err != nil {
		t.Fatalf("SetBaseURL: %v", err)
	}
	client.SetAccountPassword(testAccountPassword)

	return f, client
}

func (f *fakeAPI) on(method, path string, responses ...fakeResponse) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routes[method+" "+path] = append(f.routes[method+" "+path], responses...)
}

func (f *fakeAPI) serve(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)

	f.mu.Lock()
	key := r.Method + " " + r.URL.Path
	f.calls = append(f.calls, fakeCall{method: r.Method, path: r.URL.Path, query: r.URL.RawQuery, body: string(body), header: r.Header.Clone()})
	queue := f.routes[key]
	resp := fakeResponse{status: http.StatusTeapot, body: `{"error":true,"message":"unexpected request","code":"unexpected"}`}
	switch {
	case len(queue) == 1:
		resp = queue[0]
	case len(queue) > 1:
		resp = queue[0]
		f.routes[key] = queue[1:]
	default:
		f.t.Errorf("unexpected request %s", key)
	}
	f.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(resp.status)
	_, _ = io.WriteString(w, resp.body)
}

// callsTo returns the requests made to "METHOD /path", in order.
func (f *fakeAPI) callsTo(method, path string) []fakeCall {
	f.mu.Lock()
	defer f.mu.Unlock()

	var out []fakeCall
	for _, c := range f.calls {
		if c.method == method && c.path == path {
			out = append(out, c)
		}
	}

	return out
}

// order returns "METHOD /path" of every request, in order.
func (f *fakeAPI) order() []string {
	f.mu.Lock()
	defer f.mu.Unlock()

	out := make([]string, 0, len(f.calls))
	for _, c := range f.calls {
		out = append(out, c.method+" "+c.path)
	}

	return out
}

// fastWaits shortens the poll and retry pauses for the test.
func fastWaits(t *testing.T) {
	t.Helper()

	poll, retry, bound := pollInterval, retryInterval, deployAnswerBound
	pollInterval, retryInterval = time.Millisecond, time.Millisecond
	t.Cleanup(func() { pollInterval, retryInterval, deployAnswerBound = poll, retry, bound })
}

// serverJSON is the GET /bare-metal/:id answer for testServer.
func serverJSON(osPhase, hostname, networks string) string {
	return fmt.Sprintf(`{"identifier":"bm-1","name":%q,"powerState":"on","osFamily":"linux","osName":"Ubuntu 24.04 LTS",`+
		`"osPhase":%q,"displayState":"ready","deployedAt":"2026-10-21T10:00:00Z","remoteAccess":{"enabled":true,"winrm":false},`+
		`"datacenterIdentifier":"dc-1","ip_address":"203.0.113.10","netmask":"255.255.255.0","gateway":"203.0.113.1",`+
		`"contract":{"term":"monthly"},"linkSpeedMbps":1000,"projectIdentifier":"proj-1","privateNetworks":{"items":%s}}`,
		hostname, osPhase, networks)
}

func networksJSON(items string) string {
	return `{"available":true,"maxNetworks":8,"items":` + items + `}`
}

func newTestResource(client *govpsie.Client) *bareMetalResource {
	return &bareMetalResource{client: client}
}

func resourceSchema(t *testing.T) schema.Schema {
	t.Helper()

	var resp resource.SchemaResponse
	(&bareMetalResource{}).Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema: %v", resp.Diagnostics)
	}

	return resp.Schema
}

func nullValue(s schema.Schema) tftypes.Value {
	return tftypes.NewValue(s.Type().TerraformType(context.Background()), nil)
}

func toState(t *testing.T, s schema.Schema, m bareMetalResourceModel) tfsdk.State {
	t.Helper()

	st := tfsdk.State{Schema: s, Raw: nullValue(s)}
	if diags := st.Set(context.Background(), &m); diags.HasError() {
		t.Fatalf("state: %v", diags)
	}

	return st
}

func toPlan(t *testing.T, s schema.Schema, m bareMetalResourceModel) tfsdk.Plan {
	t.Helper()

	st := toState(t, s, m)

	return tfsdk.Plan{Schema: s, Raw: st.Raw}
}

func fromState(t *testing.T, st tfsdk.State) bareMetalResourceModel {
	t.Helper()

	var m bareMetalResourceModel
	if diags := st.Get(context.Background(), &m); diags.HasError() {
		t.Fatalf("reading state: %v", diags)
	}

	return m
}

func noTimeouts() timeouts.Value {
	return timeouts.Value{Object: types.ObjectNull(map[string]attr.Type{
		"create": types.StringType,
		"update": types.StringType,
		"delete": types.StringType,
	})}
}

func stringList(values ...string) types.List {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}

	return types.ListValueMust(types.StringType, elems)
}

func stringSet(values ...string) types.Set {
	elems := make([]attr.Value, 0, len(values))
	for _, v := range values {
		elems = append(elems, types.StringValue(v))
	}

	return types.SetValueMust(types.StringType, elems)
}

// plannedModel is the plan of a new server: the configured values set, the
// values left to the API unknown, private network 7 configured.
func plannedModel() bareMetalResourceModel {
	unknown := types.StringUnknown()

	return bareMetalResourceModel{
		ID:                unknown,
		Identifier:        unknown,
		DcIdentifier:      types.StringValue("dc-1"),
		PlanIdentifier:    types.StringValue("plan-1"),
		OsIdentifier:      types.StringValue("bmos_ubuntu2404"),
		Hostname:          types.StringValue("web1.example.com"),
		Password:          types.StringValue("Initial-pass1"),
		SSHKeyIdentifiers: stringSet("key-1"),
		RemoteAccess:      types.ObjectUnknown(remoteAccessAttrTypes),
		DNSZoneIdentifier: types.StringNull(),
		ProjectIdentifier: unknown,
		ContractTerm:      unknown,
		LinkSpeedMbps:     types.Int64Unknown(),
		UserData:          types.StringNull(),
		ScriptIdentifiers: types.ListNull(types.StringType),
		PrivateNetworkIDs: int64Set([]int64{7}),
		ReleaseReason:     types.StringNull(),
		ReleaseNote:       types.StringNull(),
		OrderStatus:       unknown,
		DeliveryHours:     types.Int64Unknown(),
		IPAddress:         unknown,
		Netmask:           unknown,
		Gateway:           unknown,
		Status:            unknown,
		OsPhase:           unknown,
		PowerState:        unknown,
		OsName:            unknown,
		OsFamily:          unknown,
		DeployedAt:        unknown,
		Timeouts:          noTimeouts(),
	}
}

// appliedModel is the state of testServer after a create.
func appliedModel() bareMetalResourceModel {
	m := plannedModel()
	m.ID = types.StringValue(testServer)
	m.Identifier = types.StringValue(testServer)
	m.RemoteAccess = defaultRemoteAccess()
	m.ProjectIdentifier = types.StringValue("proj-1")
	m.ContractTerm = types.StringValue("monthly")
	m.LinkSpeedMbps = types.Int64Value(1000)
	m.OrderStatus = types.StringValue(govpsie.BareMetalOrderSold)
	m.DeliveryHours = types.Int64Value(0)
	m.IPAddress = types.StringValue("203.0.113.10")
	m.Netmask = types.StringValue("255.255.255.0")
	m.Gateway = types.StringValue("203.0.113.1")
	m.Status = types.StringValue("ready")
	m.OsPhase = types.StringValue(govpsie.BareMetalOSPhaseRunning)
	m.PowerState = types.StringValue("on")
	m.OsName = types.StringValue("Ubuntu 24.04 LTS")
	m.OsFamily = types.StringValue("linux")
	m.DeployedAt = types.StringValue("2026-10-21T10:00:00Z")

	return m
}
