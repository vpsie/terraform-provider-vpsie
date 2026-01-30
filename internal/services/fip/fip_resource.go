package fip

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource              = &fipResource{}
	_ resource.ResourceWithConfigure = &fipResource{}
)

type fipResource struct {
	client *govpsie.Client
}

type fipResourceModel struct {
	ID           types.String `tfsdk:"id"`
	VmIdentifier types.String `tfsdk:"vm_identifier"`
	DcIdentifier types.String `tfsdk:"dc_identifier"`
	IpType       types.String `tfsdk:"ip_type"`
	IP           types.String `tfsdk:"ip"`
}

func NewFipResource() resource.Resource {
	return &fipResource{}
}

func (f *fipResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_floating_ip"
}

func (f *fipResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vm_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"dc_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (f *fipResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	f.client = client
}

func (f *fipResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan fipResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Get existing IPs to find the new one after creation
	existingIPs, err := f.client.IP.ListAllIPs(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing IPs", err.Error())
		return
	}

	existingSet := make(map[string]bool)
	for _, ip := range existingIPs {
		existingSet[ip.IP] = true
	}

	err = f.client.Fip.CreateFloatingIP(ctx, plan.VmIdentifier.ValueString(), plan.DcIdentifier.ValueString(), plan.IpType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating floating IP", err.Error())
		return
	}

	// Find the newly created IP
	allIPs, err := f.client.IP.ListAllIPs(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error listing IPs after creation", err.Error())
		return
	}

	var newIP *govpsie.IP
	for i, ip := range allIPs {
		if !existingSet[ip.IP] {
			newIP = &allIPs[i]
			break
		}
	}

	if newIP == nil {
		resp.Diagnostics.AddError("Error finding floating IP", "Could not find newly created floating IP")
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%d", newIP.ID))
	plan.IP = types.StringValue(newIP.IP)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (f *fipResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state fipResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	allIPs, err := f.client.IP.ListAllIPs(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading IPs", err.Error())
		return
	}

	found := false
	for _, ip := range allIPs {
		if ip.IP == state.IP.ValueString() {
			state.ID = types.StringValue(fmt.Sprintf("%d", ip.ID))
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (f *fipResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are ForceNew, so Update is never called
}

func (f *fipResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state fipResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := f.client.Fip.UnassignFloatingIP(ctx, state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting floating IP",
			"couldn't delete floating IP, unexpected error: "+err.Error(),
		)
	}
}
