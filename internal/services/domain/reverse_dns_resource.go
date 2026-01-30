package domain

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
	_ resource.Resource              = &reverseDnsResource{}
	_ resource.ResourceWithConfigure = &reverseDnsResource{}
)

type reverseDnsResource struct {
	client *govpsie.Client
}

type reverseDnsResourceModel struct {
	ID               types.String `tfsdk:"id"`
	VmIdentifier     types.String `tfsdk:"vm_identifier"`
	IP               types.String `tfsdk:"ip"`
	DomainIdentifier types.String `tfsdk:"domain_identifier"`
	HostName         types.String `tfsdk:"hostname"`
}

func NewReverseDnsResource() resource.Resource {
	return &reverseDnsResource{}
}

func (r *reverseDnsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_reverse_dns"
}

func (r *reverseDnsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"ip": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
			},
			"hostname": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (r *reverseDnsResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

func (r *reverseDnsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan reverseDnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reverseReq := &govpsie.ReverseRequest{
		VmIdentifier:     plan.VmIdentifier.ValueString(),
		Ip:               plan.IP.ValueString(),
		DomainIdentifier: plan.DomainIdentifier.ValueString(),
		HostName:         plan.HostName.ValueString(),
	}

	err := r.client.Domain.AddReverse(ctx, reverseReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating reverse DNS", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s", plan.VmIdentifier.ValueString(), plan.IP.ValueString()))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *reverseDnsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state reverseDnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	records, err := r.client.Domain.ListReversePTRRecords(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Error reading reverse DNS records", err.Error())
		return
	}

	found := false
	for _, record := range records {
		if record.Ip == state.IP.ValueString() && record.VmIdentifier == state.VmIdentifier.ValueString() {
			state.HostName = types.StringValue(record.HostName)
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

func (r *reverseDnsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan reverseDnsResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	reverseReq := &govpsie.ReverseRequest{
		VmIdentifier:     plan.VmIdentifier.ValueString(),
		Ip:               plan.IP.ValueString(),
		DomainIdentifier: plan.DomainIdentifier.ValueString(),
		HostName:         plan.HostName.ValueString(),
	}

	err := r.client.Domain.UpdateReverse(ctx, reverseReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating reverse DNS", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (r *reverseDnsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state reverseDnsResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Domain.DeleteReverse(ctx, state.IP.ValueString(), state.VmIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting reverse DNS",
			"couldn't delete reverse DNS, unexpected error: "+err.Error(),
		)
	}
}
