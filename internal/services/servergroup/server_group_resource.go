package servergroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &serverGroupResource{}
	_ resource.ResourceWithConfigure   = &serverGroupResource{}
	_ resource.ResourceWithImportState = &serverGroupResource{}
)

type serverGroupResource struct {
	client *govpsie.Client
}

type serverGroupResourceModel struct {
	Identifier       types.String `tfsdk:"identifier"`
	GroupName        types.String `tfsdk:"group_name"`
	GroupDescription types.String `tfsdk:"group_description"`
	IsDistributed    types.Bool   `tfsdk:"is_distributed"`
	CreatedOn        types.String `tfsdk:"created_on"`
}

// NewServerGroupResource is a helper function to create the resource.
func NewServerGroupResource() resource.Resource {
	return &serverGroupResource{}
}

func (s *serverGroupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_group"
}

func (s *serverGroupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VPSie server (VM) group.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the server group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"group_name": schema.StringAttribute{
				MarkdownDescription: "The name of the server group.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"group_description": schema.StringAttribute{
				MarkdownDescription: "The description of the server group.",
				Optional:            true,
				Computed:            true,
				Default:             stringdefault.StaticString(""),
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"is_distributed": schema.BoolAttribute{
				MarkdownDescription: "Whether the group's servers are distributed across hosts.",
				Optional:            true,
				Computed:            true,
				Default:             booldefault.StaticBool(false),
				PlanModifiers: []planmodifier.Bool{
					boolplanmodifier.RequiresReplace(),
				},
			},
			"created_on": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the server group.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (s *serverGroupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (s *serverGroupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// The create endpoint returns no identifier, so remember the identifiers that
	// exist before the call and resolve the new group as the one that appears
	// afterwards. Matching by name alone could adopt an older group with the
	// same name.
	before, err := s.client.ServerGroup.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating server group",
			"couldn't list existing server groups, unexpected error: "+err.Error(),
		)

		return
	}
	known := make(map[string]struct{}, len(before))
	for i := range before {
		known[before[i].Identifier] = struct{}{}
	}

	err = s.client.ServerGroup.Create(ctx, plan.GroupName.ValueString(), plan.GroupDescription.ValueString(), plan.IsDistributed.ValueBool())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating server group",
			"couldn't create server group, unexpected error: "+err.Error(),
		)

		return
	}

	group, err := s.getNewGroup(ctx, plan.GroupName.ValueString(), known)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating server group",
			"server group created but couldn't resolve its identifier: "+err.Error(),
		)

		return
	}

	// Only populate genuinely computed fields; keep the plan's config values so
	// they never diverge from configuration (which would force replacement of
	// these RequiresReplace attributes).
	plan.Identifier = types.StringValue(group.Identifier)
	plan.CreatedOn = types.StringValue(group.CreatedOn)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (s *serverGroupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	group, err := s.getGroupByIdentifier(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading server group",
			"couldn't read server group "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}

	if group == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// The group's configurable attributes are immutable (RequiresReplace), so a
	// managed group keeps its configured values to avoid spurious replacements if
	// the API normalizes a value. After an import only the identifier is set:
	// those attributes are then taken from the API, otherwise the next plan
	// would replace (destroy) the imported group.
	if state.GroupName.IsNull() || state.GroupName.IsUnknown() {
		state.GroupName = types.StringValue(group.GroupName)
	}
	if state.GroupDescription.IsNull() || state.GroupDescription.IsUnknown() {
		state.GroupDescription = types.StringValue(group.GroupDescription)
	}
	if state.IsDistributed.IsNull() || state.IsDistributed.IsUnknown() {
		state.IsDistributed = types.BoolValue(group.IsDistributed)
	}
	state.CreatedOn = types.StringValue(group.CreatedOn)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is a no-op: all configurable attributes force replacement.
func (s *serverGroupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverGroupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (s *serverGroupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverGroupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.ServerGroup.Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting server group",
			"couldn't delete server group "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}
}

func (s *serverGroupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// getNewGroup returns the group named name whose identifier is not in known.
func (s *serverGroupResource) getNewGroup(ctx context.Context, name string, known map[string]struct{}) (*govpsie.ServerGroup, error) {
	groups, err := s.client.ServerGroup.List(ctx)
	if err != nil {
		return nil, err
	}

	var found *govpsie.ServerGroup
	for i := range groups {
		if groups[i].GroupName != name {
			continue
		}
		if _, ok := known[groups[i].Identifier]; ok {
			continue
		}
		if found != nil {
			return nil, fmt.Errorf("several new server groups named %q appeared, cannot tell which one was created", name)
		}
		found = &groups[i]
	}

	if found == nil {
		return nil, fmt.Errorf("new server group %q not found", name)
	}

	return found, nil
}

func (s *serverGroupResource) getGroupByIdentifier(ctx context.Context, identifier string) (*govpsie.ServerGroup, error) {
	groups, err := s.client.ServerGroup.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range groups {
		if groups[i].Identifier == identifier {
			return &groups[i], nil
		}
	}

	return nil, nil
}
