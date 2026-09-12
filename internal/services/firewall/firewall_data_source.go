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

	Rules   FirewallRules `tfsdk:"rules"`
	VmsData []VmsData     `tfsdk:"vms_data"`
}

type FirewallRules struct {
	InBound  []InBoundFirewallRules  `tfsdk:"inbound"`
	OutBound []OutBoundFirewallRules `tfsdk:"outbound"`
}

type InBoundFirewallRules struct {
	ID         types.Int64  `tfsdk:"id"`
	GroupID    types.Int64  `tfsdk:"group_id"`
	UserID     types.Int64  `tfsdk:"user_id"`
	Action     types.String `tfsdk:"action"`
	Type       types.String `tfsdk:"type"`
	Comment    types.String `tfsdk:"comment"`
	Dest       types.List   `tfsdk:"dest"`
	Dport      types.String `tfsdk:"dport"`
	Proto      types.String `tfsdk:"proto"`
	Source     types.List   `tfsdk:"source"`
	Sport      types.String `tfsdk:"sport"`
	Enable     types.Int64  `tfsdk:"enable"`
	Iface      types.String `tfsdk:"iface"`
	Log        types.String `tfsdk:"log"`
	Macro      types.String `tfsdk:"macro"`
	Identifier types.String `tfsdk:"identifier"`
	CreatedOn  types.String `tfsdk:"created_on"`
	UpdatedOn  types.String `tfsdk:"updated_on"`
}

type OutBoundFirewallRules struct {
	ID         types.Int64  `tfsdk:"id"`
	GroupID    types.Int64  `tfsdk:"group_id"`
	UserID     types.Int64  `tfsdk:"user_id"`
	Action     types.String `tfsdk:"action"`
	Type       types.String `tfsdk:"type"`
	Comment    types.String `tfsdk:"comment"`
	Dest       types.List   `tfsdk:"dest"`
	Dport      types.String `tfsdk:"dport"`
	Proto      types.String `tfsdk:"proto"`
	Source     types.List   `tfsdk:"source"`
	Sport      types.String `tfsdk:"sport"`
	Enable     types.Int64  `tfsdk:"enable"`
	Iface      types.String `tfsdk:"iface"`
	Log        types.String `tfsdk:"log"`
	Macro      types.String `tfsdk:"macro"`
	Identifier types.String `tfsdk:"identifier"`
	CreatedOn  types.String `tfsdk:"created_on"`
	UpdatedOn  types.String `tfsdk:"updated_on"`
}

type VmsData struct {
	Hostname   types.String `tfsdk:"hostname"`
	Identifier types.String `tfsdk:"identifier"`
	Fullname   types.String `tfsdk:"fullname"`
	Category   types.String `tfsdk:"category"`
}

// NewFirewallDataSource is a helper function to create the data source.
func NewFirewallDataSource() datasource.DataSource {
	return &firewallDataSource{}
}

// Metadata returns the data source type name.
func (g *firewallDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewalls"
}

var common map[string]schema.Attribute = map[string]schema.Attribute{
	"id": schema.Int64Attribute{
		Computed: true,
	},
	"group_id": schema.Int64Attribute{
		Computed: true,
	},
	"user_id": schema.Int64Attribute{
		Computed: true,
	},
	"action": schema.StringAttribute{
		Computed: true,
	},
	"type": schema.StringAttribute{
		Computed: true,
	},
	"comment": schema.StringAttribute{
		Computed: true,
	},
	"dest": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
	},
	"dport": schema.StringAttribute{
		Computed: true,
	},
	"proto": schema.StringAttribute{
		Computed: true,
	},
	"source": schema.ListAttribute{
		ElementType: types.StringType,
		Computed:    true,
	},
	"sport": schema.StringAttribute{
		Computed: true,
	},
	"enable": schema.Int64Attribute{
		Computed: true,
	},
	"iface": schema.StringAttribute{
		Computed: true,
	},
	"log": schema.StringAttribute{
		Computed: true,
	},
	"macro": schema.StringAttribute{
		Computed: true,
	},
	"identifier": schema.StringAttribute{
		Computed: true,
	},
	"created_on": schema.StringAttribute{
		Computed: true,
	},
	"updated_on": schema.StringAttribute{
		Computed: true,
	},
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
						"user_name": schema.StringAttribute{
							Computed: true,
						},
						"group_name": schema.StringAttribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"updated_on": schema.StringAttribute{
							Computed: true,
						},
						"inbound_count": schema.Int64Attribute{
							Computed: true,
						},
						"outbound_count": schema.Int64Attribute{
							Computed: true,
						},
						"vms": schema.Int64Attribute{
							Computed: true,
						},
						"created_by": schema.Int64Attribute{
							Computed: true,
						},
						"rules": schema.SingleNestedAttribute{
							Computed: true,
							Attributes: map[string]schema.Attribute{
								"inbound": schema.ListNestedAttribute{
									Computed: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: common,
									},
								},
								"outbound": schema.ListNestedAttribute{
									Computed: true,
									NestedObject: schema.NestedAttributeObject{
										Attributes: common,
									},
								},
							},
						},
						"vms_data": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: map[string]schema.Attribute{
									"hostname": schema.StringAttribute{
										Computed: true,
									},
									"identifier": schema.StringAttribute{
										Computed: true,
									},
									"fullname": schema.StringAttribute{
										Computed: true,
									},
									"category": schema.StringAttribute{
										Computed: true,
									},
								},
							},
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
		var inBound []InBoundFirewallRules
		var outBound []OutBoundFirewallRules

		for _, in := range firewall.Rules.InBound {
			destList, _ := types.ListValueFrom(ctx, types.StringType, in.Dest)
			sourceList, _ := types.ListValueFrom(ctx, types.StringType, in.Source)
			inBound = append(inBound, InBoundFirewallRules{
				ID:         types.Int64Value(in.ID),
				GroupID:    types.Int64Value(in.GroupID),
				UserID:     types.Int64Value(in.UserID),
				Action:     types.StringValue(in.Action),
				Type:       types.StringValue(in.Type),
				Comment:    types.StringValue(in.Comment),
				Dest:       destList,
				Dport:      types.StringValue(in.Dport),
				Proto:      types.StringValue(in.Proto),
				Source:     sourceList,
				Sport:      types.StringValue(in.Sport),
				Enable:     types.Int64Value(in.Enable),
				Iface:      types.StringValue(in.Iface),
				Log:        types.StringValue(in.Log),
				Macro:      types.StringValue(in.Macro),
				Identifier: types.StringValue(in.Identifier),
				CreatedOn:  types.StringValue(in.CreatedOn.String()),
				UpdatedOn:  types.StringValue(in.UpdatedOn.String()),
			})
		}

		for _, out := range firewall.Rules.OutBound {
			destList, _ := types.ListValueFrom(ctx, types.StringType, out.Dest)
			sourceList, _ := types.ListValueFrom(ctx, types.StringType, out.Source)
			outBound = append(outBound, OutBoundFirewallRules{
				ID:         types.Int64Value(out.ID),
				GroupID:    types.Int64Value(out.GroupID),
				UserID:     types.Int64Value(out.UserID),
				Action:     types.StringValue(out.Action),
				Type:       types.StringValue(out.Type),
				Comment:    types.StringValue(out.Comment),
				Dest:       destList,
				Dport:      types.StringValue(out.Dport),
				Proto:      types.StringValue(out.Proto),
				Source:     sourceList,
				Sport:      types.StringValue(out.Sport),
				Enable:     types.Int64Value(out.Enable),
				Iface:      types.StringValue(out.Iface),
				Log:        types.StringValue(out.Log),
				Macro:      types.StringValue(out.Macro),
				Identifier: types.StringValue(out.Identifier),
				CreatedOn:  types.StringValue(out.CreatedOn.String()),
				UpdatedOn:  types.StringValue(out.UpdatedOn.String()),
			})
		}

		var vmsData []VmsData
		for _, vm := range firewall.VmsData {
			vmsData = append(vmsData, VmsData{
				Hostname:   types.StringValue(vm.Hostname),
				Identifier: types.StringValue(vm.Identifier),
				Fullname:   types.StringValue(vm.Fullname),
				Category:   types.StringValue(vm.Category),
			})
		}

		firewallState := firewallsModel{
			ID:            types.Int64Value(firewall.ID),
			UserName:      types.StringValue(firewall.UserName),
			GroupName:     types.StringValue(firewall.GroupName),
			Identifier:    types.StringValue(firewall.Identifier),
			CreatedOn:     types.StringValue(firewall.CreatedOn),
			UpdatedOn:     types.StringValue(firewall.UpdatedOn),
			InboundCount:  types.Int64Value(firewall.InboundCount),
			OutboundCount: types.Int64Value(firewall.OutboundCount),
			Vms:           types.Int64Value(firewall.Vms),
			CreatedBy:     types.Int64Value(firewall.CreatedBy),
			Rules: FirewallRules{
				InBound:  inBound,
				OutBound: outBound,
			},
			VmsData: vmsData,
		}

		state.Firewalls = append(state.Firewalls, firewallState)
	}

	state.ID = types.StringValue("firewalls")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
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
