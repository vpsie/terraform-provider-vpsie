package vpc

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource              = &vpcServerAssignmentResource{}
	_ resource.ResourceWithConfigure = &vpcServerAssignmentResource{}
)

type vpcServerAssignmentResource struct {
	client *govpsie.Client
}

type vpcServerAssignmentResourceModel struct {
	ID           types.String `tfsdk:"id"`
	VmIdentifier types.String `tfsdk:"vm_identifier"`
	VpcID        types.Int64  `tfsdk:"vpc_id"`
	DcIdentifier types.String `tfsdk:"dc_identifier"`
	PrivateIPID  types.Int64  `tfsdk:"private_ip_id"`
}

func NewVpcServerAssignmentResource() resource.Resource {
	return &vpcServerAssignmentResource{}
}

func (v *vpcServerAssignmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vpc_server_assignment"
}

func (v *vpcServerAssignmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"vpc_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.RequiresReplace(),
				},
			},
			"dc_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"private_ip_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (v *vpcServerAssignmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	v.client = client
}

func (v *vpcServerAssignmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vpcServerAssignmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	assignReq := &govpsie.AssignServerReq{
		VmIdentifier: plan.VmIdentifier.ValueString(),
		VpcID:        int(plan.VpcID.ValueInt64()),
		DcIdentifier: plan.DcIdentifier.ValueString(),
	}

	err := v.client.VPC.AssignServer(ctx, assignReq)
	if err != nil {
		resp.Diagnostics.AddError("Error assigning server to VPC", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%d", plan.VmIdentifier.ValueString(), plan.VpcID.ValueInt64()))

	// Try to find the private IP ID from the IP list
	ips, err := v.client.IP.ListPrivateIPs(ctx, nil)
	if err == nil {
		for _, ip := range ips {
			if ip.BoxIdentifier == plan.VmIdentifier.ValueString() {
				plan.PrivateIPID = types.Int64Value(int64(ip.ID))
				break
			}
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (v *vpcServerAssignmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vpcServerAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Verify the private IP still exists
	ips, err := v.client.IP.ListPrivateIPs(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading private IPs", err.Error())
		return
	}

	found := false
	for _, ip := range ips {
		if ip.BoxIdentifier == state.VmIdentifier.ValueString() {
			state.PrivateIPID = types.Int64Value(int64(ip.ID))
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

func (v *vpcServerAssignmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// All fields are ForceNew, so Update is never called
}

func (v *vpcServerAssignmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vpcServerAssignmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := v.client.VPC.ReleasePrivateIP(ctx, state.VmIdentifier.ValueString(), int(state.PrivateIPID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError(
			"Error releasing VPC server assignment",
			"couldn't release private IP, unexpected error: "+err.Error(),
		)
	}
}
