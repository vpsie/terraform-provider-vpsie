package gateway

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &gatewayResource{}
	_ resource.ResourceWithConfigure   = &gatewayResource{}
	_ resource.ResourceWithImportState = &gatewayResource{}
)

type gatewayResource struct {
	client *govpsie.Client
}

type gatewayResourceModel struct {
	ID                   types.Int64  `tfsdk:"id"`
	DatacenterID         types.Int64  `tfsdk:"datacenter_id"`
	IPPropertiesID       types.Int64  `tfsdk:"ip_properties_id"`
	IP                   types.String `tfsdk:"ip"`
	IsReserved           types.Int64  `tfsdk:"is_reserved"`
	IPVersion            types.String `tfsdk:"ip_version"`
	BoxID                types.Int64  `tfsdk:"box_id,omitempty"`
	IsPrimary            types.Int64  `tfsdk:"is_primary"`
	Notes                types.String `tfsdk:"notes,omitempty"`
	UserID               types.Int64  `tfsdk:"user_id"`
	UpdatedAt            types.String `tfsdk:"updated_at"`
	IsGatewayReserved    types.Int64  `tfsdk:"is_gateway_reserved"`
	IsUserAccountGateway types.Int64  `tfsdk:"is_user_account_gateway"`
	DatacenterName       types.String `tfsdk:"datacenter_name"`
	State                types.String `tfsdk:"state"`
	DcIdentifier         types.String `tfsdk:"dc_identifier"`
	CreatedBy            types.String `tfsdk:"created_by"`
	AttachedVms          []AttachedVM `tfsdk:"attached_vms"`
}

func NewGatewayResource() resource.Resource {
	return &gatewayResource{}
}

func (g *gatewayResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_gateway"
}

func (g *gatewayResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"datacenter_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ip_properties_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_reserved": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ip_version": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"box_id": schema.Int64Attribute{
				Optional: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_primary": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"notes": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_gateway_reserved": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_user_account_gateway": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"datacenter_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dc_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"attached_vms": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"gateway_mapping_id": schema.Int64Attribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Required: true,
						},
					},
				},
			},
		},
	}
}

func (g *gatewayResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (g *gatewayResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan gatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	gateway, err := g.CreateAndReturnGateway(ctx, plan.IPVersion.ValueString(), plan.DcIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating domain", err.Error())
		return
	}

	// Overwrite items with refreshed state

	plan.ID = types.Int64Value(gateway.ID)
	plan.DatacenterID = types.Int64Value(gateway.DatacenterID)
	plan.IPPropertiesID = types.Int64Value(gateway.IPPropertiesID)
	plan.IP = types.StringValue(gateway.IP)
	plan.IsReserved = types.Int64Value(gateway.IsReserved)
	plan.IPVersion = types.StringValue(gateway.IPVersion)
	plan.BoxID = types.Int64PointerValue(gateway.BoxID)
	plan.IsPrimary = types.Int64Value(gateway.IsPrimary)
	plan.Notes = types.StringPointerValue(gateway.Notes)
	plan.UserID = types.Int64Value(gateway.UserID)
	plan.UpdatedAt = types.StringValue(gateway.UpdatedAt.String())
	plan.IsGatewayReserved = types.Int64Value(gateway.IsGatewayReserved)
	plan.IsUserAccountGateway = types.Int64Value(gateway.IsUserAccountGateway)
	plan.DatacenterName = types.StringValue(gateway.DatacenterName)
	plan.State = types.StringValue(gateway.State)
	plan.DcIdentifier = types.StringValue(gateway.DcIdentifier)
	plan.CreatedBy = types.StringValue(gateway.CreatedBy)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (g *gatewayResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state gatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	gateway, err := g.client.Gateway.Get(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading vpsie gateways",
			"couldn't read vpsie gateways id "+state.ID.String()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	attached := []AttachedVM{}
	if len(gateway.AttachedVms) > 0 {
		for _, vm := range gateway.AttachedVms {
			attached = append(attached, AttachedVM{
				Identifier:       types.StringValue(vm.Identifier),
				GatewayMappingID: types.Int64Value(vm.GatewayMappingID),
			})
		}
	}

	state.ID = types.Int64Value(gateway.ID)
	state.DatacenterName = types.StringValue(gateway.DatacenterName)
	state.DatacenterID = types.Int64Value(gateway.DatacenterID)
	state.IPPropertiesID = types.Int64Value(gateway.IPPropertiesID)
	state.IP = types.StringValue(gateway.IP)
	state.IsReserved = types.Int64Value(gateway.IsReserved)
	state.IPVersion = types.StringValue(gateway.IPVersion)
	state.BoxID = types.Int64PointerValue(gateway.BoxID)
	state.IsPrimary = types.Int64Value(gateway.IsPrimary)
	state.Notes = types.StringPointerValue(gateway.Notes)
	state.UserID = types.Int64Value(gateway.UserID)
	state.UpdatedAt = types.StringValue(gateway.UpdatedAt.String())
	state.IsGatewayReserved = types.Int64Value(gateway.IsGatewayReserved)
	state.IsUserAccountGateway = types.Int64Value(gateway.IsUserAccountGateway)
	state.State = types.StringValue(gateway.State)
	state.DcIdentifier = types.StringValue(gateway.DcIdentifier)
	state.CreatedBy = types.StringValue(gateway.CreatedBy)
	state.AttachedVms = attached

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (g *gatewayResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan gatewayResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state gatewayResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	setOfState := map[string]int64{}
	for _, vm := range state.AttachedVms {
		setOfState[vm.Identifier.ValueString()] = vm.GatewayMappingID.ValueInt64()
	}

	for _, vm := range plan.AttachedVms {
		if setOfState[vm.Identifier.ValueString()] == 0 {
			err := g.client.Gateway.AttachVM(ctx, state.ID.ValueInt64(), []string{vm.Identifier.ValueString()}, 1)
			if err != nil {
				resp.Diagnostics.AddError(
					"Error attaching vm to gateway",
					"couldn't attach vm to gateway, unexpected error: "+err.Error(),
				)
			}
		}

		delete(setOfState, vm.Identifier.ValueString())
	}

	for identifier := range setOfState {
		err := g.client.Gateway.DetachVM(ctx, state.ID.ValueInt64(), []int64{setOfState[identifier]})
		if err != nil {
			resp.Diagnostics.AddError(
				"Error detaching vm from gateway",
				"couldn't detach vm from gateway, unexpected error: "+err.Error(),
			)
		}
	}

	// Overwrite items with refreshed state
	gateway, err := g.client.Gateway.Get(ctx, state.ID.ValueInt64())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading vpsie gateways",
			"couldn't read vpsie gateways id "+state.ID.String()+": "+err.Error(),
		)
		return
	}

	attached := []AttachedVM{}
	if len(gateway.AttachedVms) > 0 {
		for _, vm := range gateway.AttachedVms {
			attached = append(attached, AttachedVM{
				Identifier:       types.StringValue(vm.Identifier),
				GatewayMappingID: types.Int64Value(vm.GatewayMappingID),
			})
		}
	}

	state.ID = types.Int64Value(gateway.ID)
	state.DatacenterName = types.StringValue(gateway.DatacenterName)
	state.DatacenterID = types.Int64Value(gateway.DatacenterID)
	state.IPPropertiesID = types.Int64Value(gateway.IPPropertiesID)
	state.IP = types.StringValue(gateway.IP)
	state.IsReserved = types.Int64Value(gateway.IsReserved)
	state.IPVersion = types.StringValue(gateway.IPVersion)
	state.BoxID = types.Int64PointerValue(gateway.BoxID)
	state.IsPrimary = types.Int64Value(gateway.IsPrimary)
	state.Notes = types.StringPointerValue(gateway.Notes)
	state.UserID = types.Int64Value(gateway.UserID)
	state.UpdatedAt = types.StringValue(gateway.UpdatedAt.String())
	state.IsGatewayReserved = types.Int64Value(gateway.IsGatewayReserved)
	state.IsUserAccountGateway = types.Int64Value(gateway.IsUserAccountGateway)
	state.State = types.StringValue(gateway.State)
	state.DcIdentifier = types.StringValue(gateway.DcIdentifier)
	state.CreatedBy = types.StringValue(gateway.CreatedBy)
	state.AttachedVms = attached

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (g *gatewayResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state gatewayResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := g.client.Gateway.Delete(ctx, int(state.ID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting gateway",
			"couldn't delete gateway, unexpected error: "+err.Error(),
		)
	}
}

func (g *gatewayResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (g *gatewayResource) GetDomainByName(ctx context.Context, domainName string) (*govpsie.Domain, error) {
	domains, err := g.client.Domain.ListDomains(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, domain := range domains {
		if domainName == domain.DomainName {
			return &domain, nil
		}
	}

	return nil, fmt.Errorf("domain with name %s not found", domainName)
}

func (g *gatewayResource) CreateAndReturnGateway(ctx context.Context, ipType, dcIdentifier string) (*govpsie.Gateway, error) {

	allGateways, err := g.client.Gateway.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	gatewaySet := make(map[string]bool)
	for _, gateway := range allGateways {
		gatewaySet[gateway.IP] = true
	}

	createReq := govpsie.CreateGatewayReq{
		IPType:       ipType,
		DcIdentifier: dcIdentifier,
	}

	err = g.client.Gateway.Create(ctx, &createReq)
	if err != nil {
		return nil, err
	}

	lstAllGateways, err := g.client.Gateway.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, gateway := range lstAllGateways {
		if !gatewaySet[gateway.IP] {
			return &gateway, nil
		}
	}

	return nil, fmt.Errorf("gateway not found")
}
