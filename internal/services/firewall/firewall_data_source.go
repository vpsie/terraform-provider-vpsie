package firewall

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type firewallDataSource struct {
	client *govpsie.Client
}

type firewallDataSourceModel struct {
	Firewalls []firewallsModel `tfsdk:"firewalls"`
	ID        types.String     `tfsdk:"id"`
}

type firewallsModel struct {
	ID            types.Int64  `tfsdk:"id"`
	UserName      types.String `tfsdk:"user_name"`
	GroupName     types.String `tfsdk:"group_name"`
	Identifier    types.String `tfsdk:"identifier"`
	CreatedOn     types.String `tfsdk:"created_on"`
	UpdatedOn     types.String `tfsdk:"updated_on"`
	InboundCount  types.Int64  `tfsdk:"inbound_count"`
	OutboundCount types.Int64  `tfsdk:"outbound_count"`
	Vms           types.Int64  `tfsdk:"vms"`
	CreatedBy     types.Int64  `tfsdk:"created_by"`

	Rules   []FirewallRules `tfsdk:"rules"`
	VmsData []VmsData       `tfsdk:"vms_data"`
}

type FirewallRules struct {
	InBound  []InBoundFirewallRules  `tfsdk:"inBound"`
	OutBound []OutBoundFirewallRules `tfsdk:"outBound"`
}

type InBoundFirewallRules struct {
	ID         types.Int64    `tfsdk:"id"`
	GroupID    types.Int64    `tfsdk:"group_id"`
	UserID     types.Int64    `tfsdk:"user_id"`
	Action     types.String   `tfsdk:"action"`
	Type       types.String   `tfsdk:"type"`
	Comment    types.String   `tfsdk:"comment"`
	Dest       []types.String `tfsdk:"dest"`
	Dport      types.String   `tfsdk:"dport"`
	Proto      types.String   `tfsdk:"proto"`
	Source     []types.String `tfsdk:"source,omitempty"`
	Sport      types.String   `tfsdk:"sport"`
	Enable     types.Int64    `tfsdk:"enable"`
	Iface      types.String   `tfsdk:"iface,omitempty"`
	Log        types.String   `tfsdk:"log,omitempty"`
	Macro      types.String   `tfsdk:"macro,omitempty"`
	Identifier types.String   `tfsdk:"identifier"`
	CreatedOn  types.String   `tfsdk:"created_on"`
	UpdatedOn  types.String   `tfsdk:"updated_on"`
}

type OutBoundFirewallRules struct {
	ID         types.Int64    `tfsdk:"id"`
	GroupID    types.Int64    `tfsdk:"group_id"`
	UserID     types.Int64    `tfsdk:"user_id"`
	Action     types.String   `tfsdk:"action"`
	Type       types.String   `tfsdk:"type"`
	Comment    types.String   `tfsdk:"comment"`
	Dest       []types.String `tfsdk:"dest,omitempty"`
	Dport      types.String   `tfsdk:"dport"`
	Proto      types.String   `tfsdk:"proto"`
	Source     []types.String `tfsdk:"source"`
	Sport      types.String   `tfsdk:"sport"`
	Enable     types.Int64    `tfsdk:"enable"`
	Iface      types.String   `tfsdk:"iface,omitempty"`
	Log        types.String   `tfsdk:"log,omitempty"`
	Macro      types.String   `tfsdk:"macro,omitempty"`
	Identifier types.String   `tfsdk:"identifier"`
	CreatedOn  types.String   `tfsdk:"created_on"`
	UpdatedOn  types.String   `tfsdk:"updated_on"`
}

type VmsData struct {
	Hostname   types.String `tfsdk:"hostname"`
	Identifier types.String `tfsdk:"identifier"`
	Fullname   types.String `tfsdk:"fullname"`
	Category   types.String `tfsdk:"category"`
}

type AttachedVM struct {
	Identifier       types.String `tfsdk:"identifier"`
	GatewayMappingID types.Int64  `tfsdk:"gateway_mapping_id"`
}

// NewFirewallDataSource is a helper function to create the data source.
func NewFirewallDataSource() *firewallDataSource {
	return &firewallDataSource{}
}

// Metadata returns the data source type name.
func (g *firewallDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "firewalls"
}

// Schema defines the schema for the data source.
func (g *firewallDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"firewalls": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (f *firewallDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state firewallDataSourceModel

	firewalls, err := f.client.FirewallGroup.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting firewalls",
			"Could not get firewalls, unexpected error: "+err.Error(),
		)

		return
	}

	for _, firewall := range firewalls {

		firewallState := firewallsModel{
			ID: types.Int64Value(int64(firewall.ID)),
		}

		state.Firewalls = append(state.Firewalls, firewallState)
	}
}
func (g *firewallDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report  this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	g.client = client
}
