package snapshot

import (
	"context"
	"fmt"
	"strconv"

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
	_ resource.Resource                = &snapshotPolicyResource{}
	_ resource.ResourceWithConfigure   = &snapshotPolicyResource{}
	_ resource.ResourceWithImportState = &snapshotPolicyResource{}
)

type snapshotPolicyResource struct {
	client *govpsie.Client
}

type snapshotPolicyResourceModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	BackupPlan types.String `tfsdk:"backup_plan"`
	PlanEvery  types.String `tfsdk:"plan_every"`
	Keep       types.String `tfsdk:"keep"`
	CreatedOn  types.String `tfsdk:"created_on"`
	CreatedBy  types.String `tfsdk:"created_by"`
	Disabled   types.Int64  `tfsdk:"disabled"`
	Vms        types.List   `tfsdk:"vms"`
}

func NewSnapshotPolicyResource() resource.Resource {
	return &snapshotPolicyResource{}
}

func (s *snapshotPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_snapshot_policy"
}

func (s *snapshotPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"backup_plan": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan_every": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"keep": schema.StringAttribute{
				MarkdownDescription: "How many snapshots to retain (maximum 5). Updated in place; " +
					"changing it does not detach the policy from its servers.",
				Required: true,
			},
			"vms": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
			},
			"created_on": schema.StringAttribute{
				Computed: true,
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
			"disabled": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (s *snapshotPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	s.client = client
}

func (s *snapshotPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan snapshotPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vms []string
	if !plan.Vms.IsNull() {
		diags = plan.Vms.ElementsAs(ctx, &vms, false)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	createReq := &govpsie.CreateSnapShotPolicyReq{
		Name:       plan.Name.ValueString(),
		BackupPlan: plan.BackupPlan.ValueString(),
		PlanEvery:  plan.PlanEvery.ValueString(),
		Keep:       plan.Keep.ValueString(),
		Vms:        vms,
	}

	err := s.client.Snapshot.CreateSnapShotPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating snapshot policy", err.Error())
		return
	}

	policy, err := s.GetPolicyByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading snapshot policy after creation", err.Error())
		return
	}

	plan.Identifier = types.StringValue(policy.Identifier)
	plan.CreatedOn = types.StringValue(policy.CreatedOn)
	plan.CreatedBy = types.StringValue(policy.CreatedBy)
	plan.Disabled = types.Int64Value(policy.Disabled)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (s *snapshotPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state snapshotPolicyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := s.client.Snapshot.GetSnapShotPolicy(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading snapshot policy",
			"couldn't read snapshot policy: "+err.Error(),
		)
		return
	}

	// The API returns no policy (data: false) when it has been removed out of
	// band; drop it from state so Terraform plans its recreation.
	if policy == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(policy.Name)
	state.BackupPlan = types.StringValue(policy.BackupPlan)
	state.PlanEvery = types.StringValue(fmt.Sprintf("%d", policy.PlanEvery))
	state.Keep = types.StringValue(fmt.Sprintf("%d", policy.Keep))
	state.CreatedOn = types.StringValue(policy.CreatedOn)
	state.CreatedBy = types.StringValue(policy.CreatedBy)
	state.Disabled = types.Int64Value(policy.Disabled)

	// Refresh the VM set from the API when it has any members or the config
	// already tracks it. Leaving a null config value untouched when the policy
	// has no VMs keeps plans clean for the common (VM-less) case.
	if len(policy.Vms) > 0 || !state.Vms.IsNull() {
		vmIdentifiers := make([]string, 0, len(policy.Vms))
		for _, vm := range policy.Vms {
			vmIdentifiers = append(vmIdentifiers, vm.Identifier)
		}
		vmsList, vmsDiags := types.ListValueFrom(ctx, types.StringType, vmIdentifiers)
		resp.Diagnostics.Append(vmsDiags...)
		if resp.Diagnostics.HasError() {
			return
		}
		state.Vms = vmsList
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (s *snapshotPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan snapshotPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state snapshotPolicyResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Handle VM attachment changes
	var planVms, stateVms []string
	if !plan.Vms.IsNull() {
		diags = plan.Vms.ElementsAs(ctx, &planVms, false)
		resp.Diagnostics.Append(diags...)
	}
	if !state.Vms.IsNull() {
		diags = state.Vms.ElementsAs(ctx, &stateVms, false)
		resp.Diagnostics.Append(diags...)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	stateSet := make(map[string]bool)
	for _, vm := range stateVms {
		stateSet[vm] = true
	}
	planSet := make(map[string]bool)
	for _, vm := range planVms {
		planSet[vm] = true
	}

	var toAttach []string
	for _, vm := range planVms {
		if !stateSet[vm] {
			toAttach = append(toAttach, vm)
		}
	}
	if len(toAttach) > 0 {
		err := s.client.Snapshot.AttachSnapShotPolicy(ctx, state.Identifier.ValueString(), toAttach)
		if err != nil {
			resp.Diagnostics.AddError("Error attaching VMs to snapshot policy", err.Error())
			return
		}
	}

	// Retention is updatable in place. Replacing the policy instead would
	// detach it from every server it protects, silently dropping coverage.
	if !plan.Keep.Equal(state.Keep) {
		keep, err := strconv.ParseInt(plan.Keep.ValueString(), 10, 64)
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid keep value",
				fmt.Sprintf("keep must be a whole number, got %q: %s", plan.Keep.ValueString(), err),
			)
			return
		}

		if err := s.client.Snapshot.ManageRetainSnapShotPolicy(ctx, state.Identifier.ValueString(), keep); err != nil {
			resp.Diagnostics.AddError("Error updating policy retention", err.Error())
			return
		}

		state.Keep = plan.Keep
	}

	var toDetach []string
	for _, vm := range stateVms {
		if !planSet[vm] {
			toDetach = append(toDetach, vm)
		}
	}
	if len(toDetach) > 0 {
		err := s.client.Snapshot.DetachSnapShotPolicy(ctx, state.Identifier.ValueString(), toDetach)
		if err != nil {
			resp.Diagnostics.AddError("Error detaching VMs from snapshot policy", err.Error())
			return
		}
	}

	state.Vms = plan.Vms

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (s *snapshotPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state snapshotPolicyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Snapshot.DeleteSnapShotPolicy(ctx, state.Identifier.ValueString(), state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting snapshot policy",
			"couldn't delete snapshot policy, unexpected error: "+err.Error(),
		)
	}
}

func (s *snapshotPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (s *snapshotPolicyResource) GetPolicyByName(ctx context.Context, name string) (*govpsie.SnapShotPolicyListDetail, error) {
	policies, err := s.client.Snapshot.ListSnapShotPolicies(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		if policy.Name == name {
			return &policy, nil
		}
	}

	return nil, fmt.Errorf("snapshot policy with name %s not found", name)
}
