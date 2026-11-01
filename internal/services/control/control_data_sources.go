// Package control holds the Terraform data sources of VPSie Control: the
// platforms outside VPSie an account manages and the collectors that reach them.
// Connecting a platform needs a credential, a probe and a TLS pin confirmed by a
// person, and a collector shows its enrolment code once, so both stay in the
// portal; the data sources let a configuration depend on them.
package control

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

const (
	platformsPath  = "/apps/v2/control/platforms"
	collectorsPath = "/apps/v2/control/collectors"
)

var (
	_ datasource.DataSource              = &platformsDataSource{}
	_ datasource.DataSourceWithConfigure = &platformsDataSource{}
	_ datasource.DataSource              = &collectorsDataSource{}
	_ datasource.DataSourceWithConfigure = &collectorsDataSource{}
)

// answer is the common shape of a Control list: a level of the offering that
// says no answers 200 with available false and a reason.
type answer[T any] struct {
	Data struct {
		Available *bool  `json:"available"`
		Reason    string `json:"reason"`
		Platforms []T    `json:"platforms"`
		Collector []T    `json:"collectors"`
	} `json:"data"`
}

type platform struct {
	Identifier     string  `json:"identifier"`
	Name           string  `json:"name"`
	ConnectorType  string  `json:"connectorType"`
	EndpointHost   string  `json:"endpointHost"`
	Region         *string `json:"region"`
	AccessMode     string  `json:"accessMode"`
	State          string  `json:"state"`
	LastSyncAt     *string `json:"lastSyncAt"`
	LastSyncStatus *string `json:"lastSyncStatus"`
	VMCount        int64   `json:"vmCount"`
}

type collector struct {
	Identifier    string  `json:"identifier"`
	Name          string  `json:"name"`
	Mode          string  `json:"mode"`
	State         string  `json:"state"`
	Purpose       string  `json:"purpose"`
	AgentVersion  *string `json:"agentVersion"`
	LastSeenAt    *string `json:"lastSeenAt"`
	PlatformCount int64   `json:"platformCount"`
}

func optString(s *string) types.String {
	if s == nil {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

func fetch[T any](ctx context.Context, api *rawapi.API, path string, pick func(answer[T]) []T) (available bool, reason string, items []T, err error) {
	var out answer[T]
	if err = api.Call(ctx, rawapi.Get, path, nil, &out); err != nil {
		return false, "", nil, err
	}
	available = out.Data.Available == nil || *out.Data.Available
	return available, out.Data.Reason, pick(out), nil
}

func configure(req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) *rawapi.API {
	if req.ProviderData == nil {
		return nil
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Data Source Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return nil
	}
	return api
}

// --- vpsie_control_platforms ---

type platformsDataSource struct{ api *rawapi.API }

type platformsModel struct {
	ID        types.String    `tfsdk:"id"`
	Available types.Bool      `tfsdk:"available"`
	Reason    types.String    `tfsdk:"reason"`
	Platforms []platformModel `tfsdk:"platforms"`
}

type platformModel struct {
	Identifier     types.String `tfsdk:"identifier"`
	Name           types.String `tfsdk:"name"`
	ConnectorType  types.String `tfsdk:"connector_type"`
	EndpointHost   types.String `tfsdk:"endpoint_host"`
	Region         types.String `tfsdk:"region"`
	AccessMode     types.String `tfsdk:"access_mode"`
	State          types.String `tfsdk:"state"`
	LastSyncAt     types.String `tfsdk:"last_sync_at"`
	LastSyncStatus types.String `tfsdk:"last_sync_status"`
	VMCount        types.Int64  `tfsdk:"vm_count"`
}

// NewPlatformsDataSource creates the vpsie_control_platforms data source.
func NewPlatformsDataSource() datasource.DataSource { return &platformsDataSource{} }

func (d *platformsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_control_platforms"
}

func listHead() map[string]schema.Attribute {
	return map[string]schema.Attribute{
		"id": schema.StringAttribute{Computed: true},
		"available": schema.BoolAttribute{
			MarkdownDescription: "Whether Control is on for the account. When it is not, the list is empty and `reason` says why.",
			Computed:            true,
		},
		"reason": schema.StringAttribute{
			MarkdownDescription: "Why Control is not available: `notLicensed`, `disabledByOperator` or `disabledForTenant`.",
			Computed:            true,
		},
	}
}

func str(desc string) schema.StringAttribute {
	return schema.StringAttribute{MarkdownDescription: desc, Computed: true}
}

func (d *platformsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := listHead()
	attrs["platforms"] = schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"identifier":       str("The platform identifier."),
			"name":             str("The platform name."),
			"connector_type":   str("The platform type, for example `vmware-vcenter` or `aws-ec2`."),
			"endpoint_host":    str("The host the collector connects to."),
			"region":           str("The region, for a public cloud."),
			"access_mode":      str("`read` or `operate`."),
			"state":            str("`pending`, `connected`, `stale` or `error`."),
			"last_sync_at":     str("When the inventory was last read (RFC 3339)."),
			"last_sync_status": str("`succeeded`, `partial` or `failed`."),
			"vm_count":         schema.Int64Attribute{MarkdownDescription: "Servers in the last inventory.", Computed: true},
		}},
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the platforms outside VPSie (VMware, Nutanix, public clouds) the account manages with Control.",
		Attributes:          attrs,
	}
}

func (d *platformsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.api = configure(req, resp)
}

func (d *platformsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	available, reason, items, err := fetch(ctx, d.api, platformsPath, func(a answer[platform]) []platform { return a.Data.Platforms })
	if err != nil {
		resp.Diagnostics.AddError("Error listing Control platforms", err.Error())
		return
	}
	out := platformsModel{
		ID: types.StringValue("control_platforms"), Available: types.BoolValue(available),
		Reason: types.StringValue(reason), Platforms: make([]platformModel, 0, len(items)),
	}
	for _, p := range items {
		out.Platforms = append(out.Platforms, platformModel{
			Identifier: types.StringValue(p.Identifier), Name: types.StringValue(p.Name),
			ConnectorType: types.StringValue(p.ConnectorType), EndpointHost: types.StringValue(p.EndpointHost),
			Region: optString(p.Region), AccessMode: types.StringValue(p.AccessMode), State: types.StringValue(p.State),
			LastSyncAt: optString(p.LastSyncAt), LastSyncStatus: optString(p.LastSyncStatus), VMCount: types.Int64Value(p.VMCount),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}

// --- vpsie_control_collectors ---

type collectorsDataSource struct{ api *rawapi.API }

type collectorsModel struct {
	ID         types.String     `tfsdk:"id"`
	Available  types.Bool       `tfsdk:"available"`
	Reason     types.String     `tfsdk:"reason"`
	Collectors []collectorModel `tfsdk:"collectors"`
}

type collectorModel struct {
	Identifier    types.String `tfsdk:"identifier"`
	Name          types.String `tfsdk:"name"`
	Mode          types.String `tfsdk:"mode"`
	State         types.String `tfsdk:"state"`
	Purpose       types.String `tfsdk:"purpose"`
	AgentVersion  types.String `tfsdk:"agent_version"`
	LastSeenAt    types.String `tfsdk:"last_seen_at"`
	PlatformCount types.Int64  `tfsdk:"platform_count"`
}

// NewCollectorsDataSource creates the vpsie_control_collectors data source.
func NewCollectorsDataSource() datasource.DataSource { return &collectorsDataSource{} }

func (d *collectorsDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_control_collectors"
}

func (d *collectorsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	attrs := listHead()
	attrs["collectors"] = schema.ListNestedAttribute{
		Computed: true,
		NestedObject: schema.NestedAttributeObject{Attributes: map[string]schema.Attribute{
			"identifier":     str("The collector identifier."),
			"name":           str("The collector name."),
			"mode":           str("How it reaches the platforms: `helper`, `vpn` or `zt`."),
			"state":          str("`pending`, `enrolled` or `revoked`."),
			"purpose":        str("`control` or `migrate`, when the platform reports it."),
			"agent_version":  str("The collector agent version, once it has reported."),
			"last_seen_at":   str("When the collector last reported (RFC 3339)."),
			"platform_count": schema.Int64Attribute{MarkdownDescription: "Platforms assigned to the collector.", Computed: true},
		}},
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists the collectors (the agents that read and operate platforms outside VPSie) of the account.",
		Attributes:          attrs,
	}
}

func (d *collectorsDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	d.api = configure(req, resp)
}

func (d *collectorsDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	available, reason, items, err := fetch(ctx, d.api, collectorsPath, func(a answer[collector]) []collector { return a.Data.Collector })
	if err != nil {
		resp.Diagnostics.AddError("Error listing Control collectors", err.Error())
		return
	}
	out := collectorsModel{
		ID: types.StringValue("control_collectors"), Available: types.BoolValue(available),
		Reason: types.StringValue(reason), Collectors: make([]collectorModel, 0, len(items)),
	}
	for _, c := range items {
		out.Collectors = append(out.Collectors, collectorModel{
			Identifier: types.StringValue(c.Identifier), Name: types.StringValue(c.Name), Mode: types.StringValue(c.Mode),
			State: types.StringValue(c.State), Purpose: types.StringValue(c.Purpose), AgentVersion: optString(c.AgentVersion),
			LastSeenAt: optString(c.LastSeenAt), PlatformCount: types.Int64Value(c.PlatformCount),
		})
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, out)...)
}
