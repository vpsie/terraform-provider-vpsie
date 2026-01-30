package backup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &backupPolicyResource{}
	_ resource.ResourceWithConfigure   = &backupPolicyResource{}
	_ resource.ResourceWithImportState = &backupPolicyResource{}
)

type backupPolicyResource struct {
	client *govpsie.Client
}

type backupPolicyResourceModel struct {
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

func NewBackupPolicyResource() resource.Resource {
	return &backupPolicyResource{}
}

func (b *backupPolicyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup_policy"
}

func (b *backupPolicyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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
			},
		},
	}
}

func (b *backupPolicyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	b.client = client
}

func (b *backupPolicyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backupPolicyResourceModel
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

	createReq := &govpsie.CreateBackupPolicyReq{
		Name:       plan.Name.ValueString(),
		BackupPlan: plan.BackupPlan.ValueString(),
		PlanEvery:  plan.PlanEvery.ValueString(),
		Keep:       plan.Keep.ValueString(),
		Vms:        vms,
	}

	err := b.client.Backup.CreateBackupPolicy(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating backup policy", err.Error())
		return
	}

	policy, err := b.GetPolicyByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading backup policy after creation", err.Error())
		return
	}

	plan.Identifier = types.StringValue(policy.Identifier)
	plan.CreatedOn = types.StringValue(policy.CreatedOn)
	plan.CreatedBy = types.StringValue(policy.CreatedBy)
	plan.Disabled = types.Int64Value(int64(policy.Disabled))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (b *backupPolicyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backupPolicyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	policy, err := b.client.Backup.GetBackupPolicy(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading backup policy",
			"couldn't read backup policy: "+err.Error(),
		)
		return
	}

	state.Name = types.StringValue(policy.Name)
	state.BackupPlan = types.StringValue(policy.BackupPlan)
	state.PlanEvery = types.StringValue(fmt.Sprintf("%d", policy.PlanEvery))
	state.Keep = types.StringValue(fmt.Sprintf("%d", policy.Keep))
	state.CreatedOn = types.StringValue(policy.CreatedOn)
	state.CreatedBy = types.StringValue(policy.CreatedBy)
	state.Disabled = types.Int64Value(int64(policy.Disabled))

	vmsList, vmsDiags := types.ListValueFrom(ctx, types.StringType, policy.Vms)
	resp.Diagnostics.Append(vmsDiags...)
	if !resp.Diagnostics.HasError() {
		state.Vms = vmsList
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (b *backupPolicyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backupPolicyResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state backupPolicyResourceModel
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

	// Attach new VMs
	var toAttach []string
	for _, vm := range planVms {
		if !stateSet[vm] {
			toAttach = append(toAttach, vm)
		}
	}
	if len(toAttach) > 0 {
		err := b.client.Backup.AttachBackupPolicy(ctx, state.Identifier.ValueString(), toAttach)
		if err != nil {
			resp.Diagnostics.AddError("Error attaching VMs to backup policy", err.Error())
			return
		}
	}

	// Detach removed VMs
	var toDetach []string
	for _, vm := range stateVms {
		if !planSet[vm] {
			toDetach = append(toDetach, vm)
		}
	}
	if len(toDetach) > 0 {
		err := b.client.Backup.DetachBackupPolicy(ctx, state.Identifier.ValueString(), toDetach)
		if err != nil {
			resp.Diagnostics.AddError("Error detaching VMs from backup policy", err.Error())
			return
		}
	}

	state.Vms = plan.Vms

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (b *backupPolicyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backupPolicyResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := b.client.Backup.DeleteBackupPolicy(ctx, state.Identifier.ValueString(), state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting backup policy",
			"couldn't delete backup policy, unexpected error: "+err.Error(),
		)
	}
}

func (b *backupPolicyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (b *backupPolicyResource) GetPolicyByName(ctx context.Context, name string) (*govpsie.BackupPolicyListDetail, error) {
	policies, err := b.client.Backup.ListBackupPolicies(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, policy := range policies {
		if policy.Name == name {
			return &policy, nil
		}
	}

	return nil, fmt.Errorf("backup policy with name %s not found", name)
}
