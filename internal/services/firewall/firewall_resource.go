package firewall

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &firewallResource{}
	_ resource.ResourceWithConfigure   = &firewallResource{}
	_ resource.ResourceWithImportState = &firewallResource{}
)

type firewallResource struct {
	client *govpsie.Client
}

type firewallResourceModel struct {
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

func NewFirewallResource() resource.Resource {
	return &firewallResource{}
}

func (g *firewallResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_firewall"
}

var commonResource map[string]schema.Attribute = map[string]schema.Attribute{
	"id": schema.Int64Attribute{
		Computed: true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"group_id": schema.Int64Attribute{
		Computed: true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"user_id": schema.Int64Attribute{
		Computed: true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"action": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"type": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"comment": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"dest": schema.ListAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		ElementType: types.ListType{
			ElemType: types.StringType,
		},
	},
	"dport": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"proto": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"source": schema.ListAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.List{
			listplanmodifier.UseStateForUnknown(),
		},
		ElementType: types.ListType{
			ElemType: types.StringType,
		},
	},
	"sport": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"enable": schema.Int64Attribute{
		Computed: true,
		PlanModifiers: []planmodifier.Int64{
			int64planmodifier.UseStateForUnknown(),
		},
	},
	"iface": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"log": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"macro": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"identifier": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"created_on": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
	"updated_on": schema.StringAttribute{
		Computed: true,
		PlanModifiers: []planmodifier.String{
			stringplanmodifier.UseStateForUnknown(),
		},
	},
}

func (g *firewallResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"group_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"action": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fullname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"category": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"rules": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"in_bound": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: commonResource,
							},
						},
						"out_bound": schema.ListNestedAttribute{
							Computed: true,
							NestedObject: schema.NestedAttributeObject{
								Attributes: commonResource,
							},
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
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"identifier": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"fullname": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
						"category": schema.StringAttribute{
							Computed: true,
							PlanModifiers: []planmodifier.String{
								stringplanmodifier.UseStateForUnknown(),
							},
						},
					},
				},
			},
		},
	}
}

func (g *firewallResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (g *firewallResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan firewallResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	var rulesToCreate []govpsie.FirewallUpdateReq
	for _, rule := range plan.Rules {
		for _, inBound := range rule.InBound {
			dest := []string{}
			source := []string{}
			for _, d := range inBound.Dest.Elements() {
				dest = append(dest, d.String())
			}
			for _, s := range inBound.Source.Elements() {
				source = append(source, s.String())
			}
			rulesToCreate = append(rulesToCreate, govpsie.FirewallUpdateReq{
				Action:  inBound.Action.String(),
				Type:    inBound.Type.String(),
				Comment: inBound.Comment.String(),
				Dest:    dest,
				Dport:   inBound.Dport.String(),
				Proto:   inBound.Proto.String(),
				Source:  source,
				Sport:   inBound.Sport.String(),
				Enable:  inBound.Enable.ValueInt64(),
				Macro:   inBound.Macro.String(),
			})
		}

		for _, outBound := range rule.OutBound {
			dest := []string{}
			source := []string{}
			for _, d := range outBound.Dest.Elements() {
				dest = append(dest, d.String())
			}
			for _, s := range outBound.Source.Elements() {
				source = append(source, s.String())
			}
			rulesToCreate = append(rulesToCreate, govpsie.FirewallUpdateReq{
				Action:  outBound.Action.String(),
				Type:    outBound.Type.String(),
				Comment: outBound.Comment.String(),
				Dest:    dest,
				Dport:   outBound.Dport.String(),
				Proto:   outBound.Proto.String(),
				Source:  source,
				Sport:   outBound.Sport.String(),
				Enable:  outBound.Enable.ValueInt64(),
				Macro:   outBound.Macro.String(),
			})
		}
	}

	err := g.client.FirewallGroup.Create(ctx, plan.GroupName.String(), rulesToCreate)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating firewall",
			"couldn't create firewall, unexpected error: "+err.Error(),
		)
		return
	}

	firewall, err := g.GetFirewallGroupByName(ctx, plan.GroupName.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading firewall",
			"couldn't read firewall, unexpected error: "+err.Error(),
		)
		return
	}

	var rules []FirewallRules
	for _, rule := range firewall.Rules {
		var inBound []InBoundFirewallRules
		var outBound []OutBoundFirewallRules
		for _, in := range rule.InBound {

			dest := []types.String{}
			source := []types.String{}
			for _, d := range in.Dest {
				dest = append(dest, types.StringValue(d))
			}
			for _, s := range in.Source {
				source = append(source, types.StringValue(s))
			}

			dest_list, _ := types.ListValueFrom(ctx, types.StringType, dest)
			source_list, _ := types.ListValueFrom(ctx, types.StringType, source)

			inBound = append(inBound, InBoundFirewallRules{
				Action:  types.StringValue(in.Action),
				Type:    types.StringValue(in.Type),
				Comment: types.StringValue(in.Comment),
				Dest:    dest_list,
				Dport:   types.StringValue(in.Dport),
				Proto:   types.StringValue(in.Proto),
				Source:  source_list,
				Sport:   types.StringValue(in.Sport),
				Enable:  types.Int64Value(in.Enable),
				Macro:   types.StringValue(in.Macro),
			})
		}
		for _, out := range rule.OutBound {

			dest := []types.String{}
			source := []types.String{}
			for _, d := range out.Dest {
				dest = append(dest, types.StringValue(d))
			}
			for _, s := range out.Source {
				source = append(source, types.StringValue(s))
			}

			dest_list, _ := types.ListValueFrom(ctx, types.StringType, dest)
			source_list, _ := types.ListValueFrom(ctx, types.StringType, source)

			outBound = append(outBound, OutBoundFirewallRules{
				Action:  types.StringValue(out.Action),
				Type:    types.StringValue(out.Type),
				Comment: types.StringValue(out.Comment),
				Dest:    dest_list,
				Dport:   types.StringValue(out.Dport),
				Proto:   types.StringValue(out.Proto),
				Source:  source_list,
				Sport:   types.StringValue(out.Sport),
				Enable:  types.Int64Value(out.Enable),
				Macro:   types.StringValue(out.Macro),
			})
		}
		rules = append(rules, FirewallRules{
			InBound:  inBound,
			OutBound: outBound,
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

	plan.ID = types.Int64Value(firewall.ID)
	plan.UserName = types.StringValue(firewall.UserName)
	plan.GroupName = types.StringValue(firewall.GroupName)
	plan.Identifier = types.StringValue(firewall.Identifier)
	plan.CreatedOn = types.StringValue(firewall.CreatedOn)
	plan.UpdatedOn = types.StringValue(firewall.UpdatedOn)
	plan.InboundCount = types.Int64Value(firewall.InboundCount)
	plan.OutboundCount = types.Int64Value(firewall.OutboundCount)
	plan.Vms = types.Int64Value(firewall.Vms)
	plan.CreatedBy = types.Int64Value(firewall.CreatedBy)
	plan.Rules = rules
	plan.VmsData = vmsData

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (g *firewallResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state firewallResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	firewall, err := g.client.FirewallGroup.Get(ctx, state.ID.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading firewall",
			"couldn't read firewall, unexpected error: "+err.Error(),
		)
		return
	}

	var rules []FirewallRules
	for _, rule := range firewall.Rules {
		var inBound []InBoundFirewallRules
		var outBound []OutBoundFirewallRules
		for _, in := range rule.InBound {

			dest := []types.String{}
			source := []types.String{}
			for _, d := range in.Dest {
				dest = append(dest, types.StringValue(d))
			}
			for _, s := range in.Source {
				source = append(source, types.StringValue(s))
			}

			dest_list, _ := types.ListValueFrom(ctx, types.StringType, dest)
			source_list, _ := types.ListValueFrom(ctx, types.StringType, source)

			inBound = append(inBound, InBoundFirewallRules{
				Action:  types.StringValue(in.Action),
				Type:    types.StringValue(in.Type),
				Comment: types.StringValue(in.Comment),
				Dest:    dest_list,
				Dport:   types.StringValue(in.Dport),
				Proto:   types.StringValue(in.Proto),
				Source:  source_list,
				Sport:   types.StringValue(in.Sport),
				Enable:  types.Int64Value(in.Enable),
				Macro:   types.StringValue(in.Macro),
			})
		}
		for _, out := range rule.OutBound {

			dest := []types.String{}
			source := []types.String{}
			for _, d := range out.Dest {
				dest = append(dest, types.StringValue(d))
			}
			for _, s := range out.Source {
				source = append(source, types.StringValue(s))
			}

			dest_list, _ := types.ListValueFrom(ctx, types.StringType, dest)
			source_list, _ := types.ListValueFrom(ctx, types.StringType, source)

			outBound = append(outBound, OutBoundFirewallRules{
				Action:  types.StringValue(out.Action),
				Type:    types.StringValue(out.Type),
				Comment: types.StringValue(out.Comment),
				Dest:    dest_list,
				Dport:   types.StringValue(out.Dport),
				Proto:   types.StringValue(out.Proto),
				Source:  source_list,
				Sport:   types.StringValue(out.Sport),
				Enable:  types.Int64Value(out.Enable),
				Macro:   types.StringValue(out.Macro),
			})
		}
		rules = append(rules, FirewallRules{
			InBound:  inBound,
			OutBound: outBound,
		})

	}

	var vmsData []VmsData
	for _, vm := range firewall.Vms {
		vmsData = append(vmsData, VmsData{
			Hostname:   types.StringValue(vm.Hostname),
			Identifier: types.StringValue(vm.Identifier),
			Fullname:   types.StringValue(vm.Fullname),
			Category:   types.StringValue(vm.Category),
		})

	}

	state.ID = types.Int64Value(firewall.Group.ID)
	state.UserName = types.StringValue(firewall.Group.UserName)
	state.GroupName = types.StringValue(firewall.Group.GroupName)
	state.Identifier = types.StringValue(firewall.Group.Identifier)
	state.CreatedOn = types.StringValue(firewall.Group.CreatedOn)
	state.UpdatedOn = types.StringValue(firewall.Group.UpdatedOn)
	state.InboundCount = types.Int64Value(firewall.Group.InboundCount)
	state.OutboundCount = types.Int64Value(firewall.Group.OutboundCount)
	state.Vms = types.Int64Value(firewall.Group.Vms)
	state.CreatedBy = types.Int64Value(firewall.Group.CreatedBy)
	state.Rules = rules
	state.VmsData = vmsData

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (g *firewallResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {

}

// Delete deletes the resource and removes the Terraform state on success.
func (f *firewallResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state firewallResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := f.client.FirewallGroup.Delete(ctx, state.ID.String())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting firewall",
			"couldn't delete firewall, unexpected error: "+err.Error(),
		)
	}
}

func (f *firewallResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (f *firewallResource) GetFirewallGroupByName(ctx context.Context, name string) (*govpsie.FirewallGroupListData, error) {
	firewalls, err := f.client.FirewallGroup.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, firewall := range firewalls {
		if firewall.GroupName == name {
			return &firewall, nil
		}
	}

	return nil, fmt.Errorf("firewall group not found: %s", name)
}
