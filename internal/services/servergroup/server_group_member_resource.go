package servergroup

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &serverGroupMemberResource{}
	_ resource.ResourceWithConfigure   = &serverGroupMemberResource{}
	_ resource.ResourceWithImportState = &serverGroupMemberResource{}
)

type serverGroupMemberResource struct {
	client *govpsie.Client
}

type serverGroupMemberResourceModel struct {
	GroupIdentifier types.String `tfsdk:"group_identifier"`
	VMIdentifier    types.String `tfsdk:"vm_identifier"`
}

// NewServerGroupMemberResource is a helper function to create the resource.
func NewServerGroupMemberResource() resource.Resource {
	return &serverGroupMemberResource{}
}

func (s *serverGroupMemberResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_group_member"
}

func (s *serverGroupMemberResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Attaches a VM to a VPSie server group. Destroying this resource detaches the VM from the group.",
		Attributes: map[string]schema.Attribute{
			"group_identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier of the server group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vm_identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier of the VM to attach to the group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (s *serverGroupMemberResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	s.client = client
}

// Create attaches the VM to the server group.
func (s *serverGroupMemberResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverGroupMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.ServerGroup.Join(ctx, plan.GroupIdentifier.ValueString(), plan.VMIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error attaching VM to server group",
			"couldn't attach VM to server group, unexpected error: "+err.Error(),
		)

		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read confirms the target server group still exists. If the group is gone the
// membership cannot exist either, so the resource is removed from state.
func (s *serverGroupMemberResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverGroupMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	groups, err := s.client.ServerGroup.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading server group membership",
			"couldn't list server groups: "+err.Error(),
		)

		return
	}

	found := false
	for i := range groups {
		if groups[i].Identifier == state.GroupIdentifier.ValueString() {
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

// Update is a no-op: both attributes force replacement.
func (s *serverGroupMemberResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverGroupMemberResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete detaches the VM from the server group.
func (s *serverGroupMemberResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverGroupMemberResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.ServerGroup.Detach(ctx, state.GroupIdentifier.ValueString(), state.VMIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error detaching VM from server group",
			"couldn't detach VM from server group, unexpected error: "+err.Error(),
		)

		return
	}
}

// ImportState imports a membership using the composite ID
// "group_identifier,vm_identifier".
func (s *serverGroupMemberResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Expected import ID in the format "group_identifier,vm_identifier".`,
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("group_identifier"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("vm_identifier"), parts[1])...)
}
