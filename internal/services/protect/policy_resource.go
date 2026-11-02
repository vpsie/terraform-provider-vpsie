package protect

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

var (
	_ resource.Resource                = &policyResource{}
	_ resource.ResourceWithConfigure   = &policyResource{}
	_ resource.ResourceWithImportState = &policyResource{}
)

// rpoChoices are the recovery point objectives the platform accepts, in minutes.
var rpoChoices = []int64{15, 30, 60, 120, 240, 480, 720, 1440}

type retentionModel struct {
	KeepRecentHours types.Int64 `tfsdk:"keep_recent_hours"`
	KeepDaily       types.Int64 `tfsdk:"keep_daily"`
	KeepWeekly      types.Int64 `tfsdk:"keep_weekly"`
}

type policyModel struct {
	Identifier       types.String    `tfsdk:"identifier"`
	Name             types.String    `tfsdk:"name"`
	Tier             types.String    `tfsdk:"tier"`
	Datacenter       types.String    `tfsdk:"datacenter"`
	Plan             types.String    `tfsdk:"plan"`
	RPOMinutes       types.Int64     `tfsdk:"rpo_minutes"`
	Retention        *retentionModel `tfsdk:"retention"`
	ImmutabilityDays types.Int64     `tfsdk:"immutability_days"`
	Consistency      types.String    `tfsdk:"consistency"`
	VerifyMode       types.String    `tfsdk:"verify_mode"`
	RTOTargetMinutes types.Int64     `tfsdk:"rto_target_minutes"`
	State            types.String    `tfsdk:"state"`
	Version          types.Int64     `tfsdk:"version"`
	CreatedAt        types.String    `tfsdk:"created_at"`
}

type policyResource struct {
	api *policyAPI
}

// NewPolicyResource creates the vpsie_protect_policy resource.
func NewPolicyResource() resource.Resource {
	return &policyResource{}
}

func (r *policyResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protect_policy"
}

func int64OneOf(values ...int64) validator.Int64 {
	return int64validator.OneOf(values...)
}

func (r *policyResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	keepAttr := func(desc string, upper int64) schema.Int64Attribute {
		return schema.Int64Attribute{
			MarkdownDescription: desc,
			Required:            true,
			Validators:          []validator.Int64{int64validator.Between(0, upper)},
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Backup or Disaster-recovery protection policy: how often protected servers are copied (RPO), " +
			"how long their recovery points are kept and how they are verified. Servers are added to a policy in the portal " +
			"or through the API; a policy that still has protected servers cannot be removed. The `backup` and `dr` tiers " +
			"need the matching licence module and the account's permission to create.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The policy identifier.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The policy name. Removing the policy needs this name, so it is read from the state at destroy.",
				Required:            true,
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: "`backup` or `dr` (disaster recovery). Changing it replaces the policy.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.OneOf("backup", "dr")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"datacenter": schema.StringAttribute{
				MarkdownDescription: "The datacenter identifier the protection copies into. Changing it replaces the policy.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"plan": schema.StringAttribute{
				MarkdownDescription: "The protection plan identifier (GUID) the policy is billed under, when the account has more than one. Changing it replaces the policy.",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace(), stringplanmodifier.UseStateForUnknown()},
			},
			"rpo_minutes": schema.Int64Attribute{
				MarkdownDescription: "The recovery point objective in minutes: 15, 30, 60, 120, 240, 480, 720 or 1440.",
				Required:            true,
				Validators:          []validator.Int64{int64OneOf(rpoChoices...)},
			},
			"retention": schema.SingleNestedAttribute{
				MarkdownDescription: "How many recovery points are kept.",
				Required:            true,
				Attributes: map[string]schema.Attribute{
					"keep_recent_hours": keepAttr("Hours of every recovery point to keep (0 to 168).", 168),
					"keep_daily":        keepAttr("Daily recovery points to keep (0 to 60).", 60),
					"keep_weekly":       keepAttr("Weekly recovery points to keep (0 to 26).", 26),
				},
			},
			"immutability_days": schema.Int64Attribute{
				MarkdownDescription: "Days a recovery point cannot be deleted (0 to 30). The platform default is 7.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.Int64{int64validator.Between(0, 30)},
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"consistency": schema.StringAttribute{
				MarkdownDescription: "`app` (application-consistent, the default) or `crash`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.OneOf("app", "crash")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"verify_mode": schema.StringAttribute{
				MarkdownDescription: "Boot checks of recovery points: `off` (the default), `latest-daily` or `latest-weekly`.",
				Optional:            true,
				Computed:            true,
				Validators:          []validator.String{stringvalidator.OneOf("off", "latest-daily", "latest-weekly")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"rto_target_minutes": schema.Int64Attribute{
				MarkdownDescription: "The recovery time objective to report against, in minutes (5 to 1440). Removing it clears the target.",
				Optional:            true,
				Validators:          []validator.Int64{int64validator.Between(5, 1440)},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The policy state: `active`, `paused` or `pending_approval`.",
				Computed:            true,
			},
			"version": schema.Int64Attribute{
				MarkdownDescription: "The policy version the platform uses to refuse edits made from a stale read.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the policy was created (RFC 3339).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *policyResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Resource Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return
	}
	r.api = &policyAPI{api: api}
}

func optionalInt(v types.Int64) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	n := v.ValueInt64()
	return &n
}

func retentionOf(m *retentionModel) retention {
	return retention{
		KeepRecentHours: m.KeepRecentHours.ValueInt64(),
		KeepDaily:       m.KeepDaily.ValueInt64(),
		KeepWeekly:      m.KeepWeekly.ValueInt64(),
	}
}

func stringOrNull(s *string) types.String {
	if s == nil || *s == "" {
		return types.StringNull()
	}
	return types.StringValue(*s)
}

// applyPolicy copies what the platform reports into the model.
func applyPolicy(m *policyModel, p *policy) {
	m.Identifier = types.StringValue(p.Identifier)
	m.Name = types.StringValue(p.Name)
	m.Tier = types.StringValue(p.Tier)
	if p.Datacenter.Identifier != nil {
		m.Datacenter = types.StringValue(*p.Datacenter.Identifier)
	}
	m.Plan = stringOrNull(p.Plan.Identifier)
	m.RPOMinutes = types.Int64Value(p.RPOMinutes)
	m.Retention = &retentionModel{
		KeepRecentHours: types.Int64Value(p.Retention.KeepRecentHours),
		KeepDaily:       types.Int64Value(p.Retention.KeepDaily),
		KeepWeekly:      types.Int64Value(p.Retention.KeepWeekly),
	}
	m.ImmutabilityDays = types.Int64Value(p.ImmutabilityDays)
	m.Consistency = types.StringValue(p.Consistency)
	m.VerifyMode = types.StringValue(p.VerifyMode)
	if p.RTOTargetMinutes != nil {
		m.RTOTargetMinutes = types.Int64Value(*p.RTOTargetMinutes)
	} else {
		m.RTOTargetMinutes = types.Int64Null()
	}
	m.State = types.StringValue(p.State)
	m.Version = types.Int64Value(p.Version)
	m.CreatedAt = types.StringValue(p.CreatedAt)
}

func (r *policyResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.api.create(ctx, createRequest{
		Name:             plan.Name.ValueString(),
		Tier:             plan.Tier.ValueString(),
		Datacenter:       plan.Datacenter.ValueString(),
		Plan:             plan.Plan.ValueString(),
		RPOMinutes:       plan.RPOMinutes.ValueInt64(),
		Retention:        retentionOf(plan.Retention),
		ImmutabilityDays: optionalInt(plan.ImmutabilityDays),
		Consistency:      plan.Consistency.ValueString(),
		VerifyMode:       plan.VerifyMode.ValueString(),
		RTOTargetMinutes: optionalInt(plan.RTOTargetMinutes),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating protection policy", err.Error())
		return
	}
	applyPolicy(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *policyResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state policyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.api.get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading protection policy", err.Error())
		return
	}
	if p == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	applyPolicy(&state, p)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *policyResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state policyModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	edited, err := r.api.edit(ctx, state.Identifier.ValueString(), editRequest{
		Version:          state.Version.ValueInt64(),
		Name:             plan.Name.ValueString(),
		RPOMinutes:       plan.RPOMinutes.ValueInt64(),
		Retention:        retentionOf(plan.Retention),
		ImmutabilityDays: optionalInt(plan.ImmutabilityDays),
		Consistency:      plan.Consistency.ValueString(),
		VerifyMode:       plan.VerifyMode.ValueString(),
		RTOTargetMinutes: optionalInt(plan.RTOTargetMinutes),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error updating protection policy", err.Error())
		return
	}
	applyPolicy(&plan, edited)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *policyResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state policyModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.api.remove(ctx, state.Identifier.ValueString(), state.Name.ValueString()); err != nil {
		if p, getErr := r.api.get(ctx, state.Identifier.ValueString()); getErr == nil && p == nil {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error removing protection policy %q", state.Name.ValueString()), err.Error())
	}
}

func (r *policyResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}
