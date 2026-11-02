package protect

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

var (
	_ resource.Resource              = &serverResource{}
	_ resource.ResourceWithConfigure = &serverResource{}
)

// identifierPattern is what the platform accepts as a policy or source identifier.
var identifierPattern = regexp.MustCompile(`^[A-Za-z0-9-]{1,64}$`)

type serverModelT struct {
	Identifier types.String `tfsdk:"identifier"`
	Policy     types.String `tfsdk:"policy"`
	Source     types.String `tfsdk:"source"`
	VMRef      types.String `tfsdk:"vm_ref"`
	Paused     types.Bool   `tfsdk:"paused"`
	Name       types.String `tfsdk:"name"`
	Tier       types.String `tfsdk:"tier"`
	State      types.String `tfsdk:"state"`
	PausedBy   types.String `tfsdk:"paused_by"`
	CreatedAt  types.String `tfsdk:"created_at"`
}

type serverResource struct {
	api *serverAPI
}

// NewServerResource creates the vpsie_protected_server resource.
func NewServerResource() resource.Resource {
	return &serverResource{}
}

func (r *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_protected_server"
}

func (r *serverResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	identifierRule := stringvalidator.RegexMatches(identifierPattern, "must be 1 to 64 letters, digits or hyphens")
	resp.Schema = schema.Schema{
		MarkdownDescription: "A server of a connected Migrate source that is protected under a Backup or Disaster-recovery policy. " +
			"Adding it starts the first copy; pausing and resuming stop and restart the copies; destroying the resource ends the " +
			"protection, and the platform keeps the recovery points until their immutability locks end. The same server cannot be " +
			"protected again before then. The policy's tier needs the matching licence module and the account's permission to create.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The protected server identifier.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"policy": schema.StringAttribute{
				MarkdownDescription: "The identifier of an active protection policy (`vpsie_protect_policy`). Changing it replaces the protection and starts a new first copy.",
				Required:            true,
				Validators:          []validator.String{identifierRule},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"source": schema.StringAttribute{
				MarkdownDescription: "The identifier of an active Migrate source (`vpsie_migration_source`) whose connector supports protection. Changing it replaces the protection.",
				Required:            true,
				Validators:          []validator.String{identifierRule},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vm_ref": schema.StringAttribute{
				MarkdownDescription: "The server's reference on the source, as the source lists it. Changing it replaces the protection.",
				Required:            true,
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 512)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"paused": schema.BoolAttribute{
				MarkdownDescription: "Whether the protection is paused by you. Defaults to `false`. A pause the platform made itself reads as `true` here and cannot be ended from Terraform (see `paused_by`).",
				Optional:            true,
				Computed:            true,
				PlanModifiers:       []planmodifier.Bool{boolplanmodifier.UseStateForUnknown()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The server name on the source. Removing the protection needs this name, so it is read from the state at destroy.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: "The tier of the policy: `backup` or `dr`.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"state": schema.StringAttribute{
				MarkdownDescription: "The protection state: `pending_approval`, `seeding`, `protected` or `paused`.",
				Computed:            true,
			},
			"paused_by": schema.StringAttribute{
				MarkdownDescription: "Who paused the protection when it is paused: `tenant` or `platform`.",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the protection was created (RFC 3339).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

func (r *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Resource Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return
	}
	r.api = &serverAPI{api: api}
}

// applyServer copies what the platform reports into the model. The policy,
// source and vm_ref the user configured stay as written.
func applyServer(m *serverModelT, p *protectedServer) {
	m.Identifier = types.StringValue(p.Identifier)
	m.Name = types.StringValue(p.Name)
	m.Tier = types.StringValue(p.Tier)
	m.State = types.StringValue(p.State)
	m.PausedBy = stringOrNull(p.PausedBy)
	m.Paused = types.BoolValue(p.State == "paused")
	m.CreatedAt = types.StringValue(p.CreatedAt)
	if p.Policy.Identifier != nil {
		m.Policy = types.StringValue(*p.Policy.Identifier)
	}
	if p.Source.Identifier != nil {
		m.Source = types.StringValue(*p.Source.Identifier)
	}
}

func (r *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverModelT
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.api.add(ctx, plan.Policy.ValueString(), plan.Source.ValueString(), plan.VMRef.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error protecting server", err.Error())
		return
	}
	if !plan.Paused.IsNull() && !plan.Paused.IsUnknown() && plan.Paused.ValueBool() {
		// Keep the identifier in the state even when the pause fails.
		applyServer(&plan, created)
		resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
		paused, pauseErr := r.api.setPaused(ctx, created.Identifier, true)
		if pauseErr != nil {
			resp.Diagnostics.AddError("Error pausing protected server", pauseErr.Error())
			return
		}
		created = paused
	}
	applyServer(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverModelT
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	p, err := r.api.get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading protected server", err.Error())
		return
	}
	if p == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	applyServer(&state, p)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state serverModelT
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	want := !plan.Paused.IsNull() && !plan.Paused.IsUnknown() && plan.Paused.ValueBool()
	current := &protectedServer{
		Identifier: state.Identifier.ValueString(), Name: state.Name.ValueString(), Tier: state.Tier.ValueString(),
		State: state.State.ValueString(), CreatedAt: state.CreatedAt.ValueString(),
	}
	if state.PausedBy.ValueString() != "" {
		by := state.PausedBy.ValueString()
		current.PausedBy = &by
	}
	if want != state.Paused.ValueBool() {
		changed, err := r.api.setPaused(ctx, state.Identifier.ValueString(), want)
		if err != nil {
			verb := "resuming"
			if want {
				verb = "pausing"
			}
			resp.Diagnostics.AddError(fmt.Sprintf("Error %s protected server", verb), err.Error())
			return
		}
		current = changed
	}
	applyServer(&plan, current)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverModelT
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.api.remove(ctx, state.Identifier.ValueString(), state.Name.ValueString()); err != nil {
		if p, getErr := r.api.get(ctx, state.Identifier.ValueString()); getErr == nil && p == nil {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error ending protection of %q", state.Name.ValueString()), err.Error())
	}
}
