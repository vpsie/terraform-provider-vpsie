package vpc

import (
	"context"
	"fmt"
	"time"

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
	VpcIP        types.String `tfsdk:"vpc_ip"`
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
			"vpc_ip": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Private IPv4 address assigned to the server inside the VPC. Use this to point other resources, such as a load balancer backend, at the server.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
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

	// The address is allocated asynchronously, so poll the VPC's server listing
	// until it appears. Delete needs the private-IP id to release the address,
	// so giving up without it would leak the allocation on destroy.
	//
	// Note this uses the VPC server listing rather than the private-IP listing:
	// addresses handed out inside a VPC are not reported by the latter.
	server, err := waitForVpcServer(ctx, v.client, int(plan.VpcID.ValueInt64()), plan.VmIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error resolving VPC address for server", err.Error())
		return
	}

	plan.PrivateIPID = types.Int64Value(int64(server.PrivateIPID))
	plan.VpcIP = types.StringValue(server.VpcIP)

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

	servers, err := v.client.VPC.ListServers(ctx, int(state.VpcID.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Error reading VPC servers", err.Error())
		return
	}

	found := false
	for _, server := range servers {
		if server.Identifier == state.VmIdentifier.ValueString() {
			state.PrivateIPID = types.Int64Value(int64(server.PrivateIPID))
			state.VpcIP = types.StringValue(server.VpcIP)
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

// waitForVpcServer polls a VPC's server listing until the given server appears
// with an allocated address, or the context deadline passes.
func waitForVpcServer(ctx context.Context, client *govpsie.Client, vpcID int, vmIdentifier string) (*govpsie.VpcServer, error) {
	const maxAttempts = 30

	for attempt := 0; ; attempt++ {
		servers, err := client.VPC.ListServers(ctx, vpcID)
		if err != nil {
			return nil, err
		}

		for _, server := range servers {
			if server.Identifier == vmIdentifier && server.VpcIP != "" {
				return &server, nil
			}
		}

		if attempt == maxAttempts-1 {
			return nil, fmt.Errorf(
				"server %q did not receive an address in VPC %d; the assignment may still be in progress, re-run to pick it up",
				vmIdentifier, vpcID)
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(5 * time.Second):
		}
	}
}
