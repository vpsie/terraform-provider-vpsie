package files

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
	_ resource.Resource                     = &smbUserResource{}
	_ resource.ResourceWithConfigure        = &smbUserResource{}
	_ resource.ResourceWithImportState      = &smbUserResource{}
	_ resource.ResourceWithConfigValidators = &smbUserResource{}
)

type smbUserModel struct {
	Identifier        types.String `tfsdk:"identifier"`
	Username          types.String `tfsdk:"username"`
	Password          types.String `tfsdk:"password"`
	DcIdentifier      types.String `tfsdk:"dc_identifier"`
	VPCID             types.Int64  `tfsdk:"vpc_id"`
	ShareIdentifier   types.String `tfsdk:"share_identifier"`
	ShareCount        types.Int64  `tfsdk:"share_count"`
	PasswordChangedAt types.String `tfsdk:"password_changed_at"`
	CreatedAt         types.String `tfsdk:"created_at"`
}

type smbUserResource struct {
	api *shareAPI
}

// NewSmbUserResource creates the vpsie_file_smb_user resource.
func NewSmbUserResource() resource.Resource {
	return &smbUserResource{}
}

func (r *smbUserResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_smb_user"
}

// replaceUnlessImported replaces the user when the value changes, but not when
// the state holds none: the platform does not report the VPC id or the share a
// user was created for, so an imported user takes them from the configuration.
func replaceUnlessImportedString() planmodifier.String {
	return stringplanmodifier.RequiresReplaceIf(
		func(_ context.Context, req planmodifier.StringRequest, resp *stringplanmodifier.RequiresReplaceIfFuncResponse) {
			resp.RequiresReplace = !req.StateValue.IsNull()
		},
		"Replaces the user when the value changes; an imported user has none to compare.",
		"Replaces the user when the value changes; an imported user has none to compare.",
	)
}

func (r *smbUserResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "An SMB user of the Files service: a name and password that may be granted on SMB shares " +
			"served in one VPC. Name the VPC with `dc_identifier` and `vpc_id`, or name a share with `share_identifier` " +
			"to take the VPC that share is served in. The platform keeps only a hash of the password, so the password is " +
			"never read back: it stays in the Terraform state (marked sensitive) as configured, and changing it sets a new " +
			"password and ends the user's open sessions. Grant the user on a share with `vpsie_file_share_smb`. " +
			"Needs the `fileStorage` licence module.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The SMB user identifier.",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The user name: 3 to 32 lower-case letters, digits, hyphens or underscores, starting with a letter. " +
					"Reserved names are refused. Changing it replaces the user.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password: 12 to 128 characters, a letter and a digit or symbol, not containing the user name. " +
					"Changing it sets a new password.",
				Required:  true,
				Sensitive: true,
				Validators: []validator.String{
					stringvalidator.LengthBetween(12, 128),
				},
			},
			"dc_identifier": schema.StringAttribute{
				MarkdownDescription: "The datacenter identifier. Set it with `vpc_id`; with `share_identifier` the platform fills it in. " +
					"Changing it replaces the user.",
				Optional:      true,
				Computed:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown(), stringplanmodifier.RequiresReplace()},
			},
			"vpc_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric id of the VPC the user belongs to. Set it with `dc_identifier`. The platform does not report it " +
					"back, so the configured value is kept. Changing it replaces the user.",
				Optional:   true,
				Validators: []validator.Int64{int64validator.AtLeast(1)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplaceIf(
					func(_ context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !req.StateValue.IsNull()
					},
					"Replaces the user when the VPC changes; an imported user has no VPC id to compare.",
					"Replaces the user when the VPC changes; an imported user has no VPC id to compare.",
				)},
			},
			"share_identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier of a share whose VPC the user belongs to, instead of `dc_identifier` and `vpc_id`. " +
					"Changing it replaces the user.",
				Optional:      true,
				PlanModifiers: []planmodifier.String{replaceUnlessImportedString()},
			},
			"share_count": schema.Int64Attribute{
				MarkdownDescription: "How many shares currently grant the user.",
				Computed:            true,
			},
			"password_changed_at": schema.StringAttribute{
				MarkdownDescription: "When the password was last set (RFC 3339).",
				Computed:            true,
			},
			"created_at": schema.StringAttribute{
				MarkdownDescription: "When the user was created (RFC 3339).",
				Computed:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
		},
	}
}

// ConfigValidators requires exactly one way to name the VPC.
func (r *smbUserResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		exactlyOneVPCName(),
	}
}

func (r *smbUserResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	api, err := rawapi.FromProviderData(req.ProviderData)
	if err != nil {
		resp.Diagnostics.AddError("Unexpected Resource Configuration Type", err.Error()+" Please report this issue to the provider developers.")
		return
	}
	r.api = &shareAPI{api: api}
}

func applySmbUser(m *smbUserModel, u *smbUser) {
	m.Identifier = types.StringValue(u.Identifier)
	m.Username = types.StringValue(u.Username)
	if u.DcIdentifier != "" {
		m.DcIdentifier = types.StringValue(u.DcIdentifier)
	} else if m.DcIdentifier.IsUnknown() {
		m.DcIdentifier = types.StringNull()
	}
	m.ShareCount = types.Int64Null()
	if u.ShareCount != nil {
		m.ShareCount = types.Int64Value(*u.ShareCount)
	}
	m.PasswordChangedAt = types.StringValue(u.PasswordChangedAt)
	m.CreatedAt = types.StringValue(u.CreatedAt)
}

func (r *smbUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan smbUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	body := smbUserRequest{Username: plan.Username.ValueString(), Password: plan.Password.ValueString()}
	if plan.ShareIdentifier.IsNull() {
		body.DcIdentifier, body.VPCID = plan.DcIdentifier.ValueString(), plan.VPCID.ValueInt64()
	} else {
		body.ShareIdentifier = plan.ShareIdentifier.ValueString()
	}
	created, err := r.api.createSmbUser(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating SMB user", err.Error())
		return
	}
	applySmbUser(&plan, created)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *smbUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state smbUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	u, err := r.api.getSmbUser(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading SMB user", err.Error())
		return
	}
	if u == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	applySmbUser(&state, u)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *smbUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state smbUserModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Identifier.ValueString()
	plan.Identifier = state.Identifier
	if !plan.Password.Equal(state.Password) {
		if err := r.api.setSmbPassword(ctx, id, plan.Password.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error setting the SMB user password", err.Error())
			return
		}
	}
	u, err := r.api.getSmbUser(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading SMB user after update", err.Error())
		return
	}
	if u == nil {
		resp.Diagnostics.AddError("SMB user disappeared", fmt.Sprintf("SMB user %s is gone after the password change.", id))
		return
	}
	applySmbUser(&plan, u)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *smbUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state smbUserModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Identifier.ValueString()
	if err := r.api.removeSmbUser(ctx, id); err != nil {
		// The client cannot tell a 404 from other failures: a user the list no longer holds is gone.
		if u, getErr := r.api.getSmbUser(ctx, id); getErr == nil && u == nil {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error deleting SMB user %q", state.Username.ValueString()), err.Error())
	}
}

func (r *smbUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}
