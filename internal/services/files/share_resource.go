package files

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/terraform-provider-vpsie/internal/services/rawapi"
)

var (
	_ resource.Resource                = &shareResource{}
	_ resource.ResourceWithConfigure   = &shareResource{}
	_ resource.ResourceWithImportState = &shareResource{}
)

var shareNameRE = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)

type accessModel struct {
	VMIdentifier types.String `tfsdk:"vm_identifier"`
	CIDR         types.String `tfsdk:"cidr"`
	Mode         types.String `tfsdk:"mode"`
}

type shareModel struct {
	Identifier   types.String  `tfsdk:"identifier"`
	Name         types.String  `tfsdk:"name"`
	DcIdentifier types.String  `tfsdk:"dc_identifier"`
	VPCID        types.Int64   `tfsdk:"vpc_id"`
	SizeGb       types.Int64   `tfsdk:"size_gb"`
	Access       []accessModel `tfsdk:"access"`
	State        types.String  `tfsdk:"state"`
	GatewayIP    types.String  `tfsdk:"gateway_vpc_ip"`
	MountNFS4    types.String  `tfsdk:"mount_nfs4"`
	CreatedAt    types.String  `tfsdk:"created_at"`
}

type shareResource struct {
	api *shareAPI
}

// NewShareResource creates the vpsie_file_share resource.
func NewShareResource() resource.Resource {
	return &shareResource{}
}

func (r *shareResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_file_share"
}

func (r *shareResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computed := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{
			MarkdownDescription: desc,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "An NFS file share served inside a VPC. Creating, growing, shrinking and deleting run on the " +
			"platform, so each waits until the share is ready (or gone) before it returns. Creating and growing need the " +
			"`fileStorage` licence module and quota; reading, changing access, shrinking and deleting never do. " +
			"Deleting refuses a share that still has an import running and removes it only when the share is not in use.",
		Attributes: map[string]schema.Attribute{
			"identifier": computed("The share identifier."),
			"name": schema.StringAttribute{
				MarkdownDescription: "The share name: 1 to 63 lower-case letters, digits or hyphens, not starting with a hyphen. " +
					"It is the export name and is needed to delete the share. Changing it replaces the share.",
				Required: true,
				Validators: []validator.String{
					stringvalidator.RegexMatches(shareNameRE, "must be 1 to 63 lower-case letters, digits or hyphens, not starting with a hyphen"),
				},
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"dc_identifier": schema.StringAttribute{
				MarkdownDescription: "The datacenter identifier. Changing it replaces the share.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"vpc_id": schema.Int64Attribute{
				MarkdownDescription: "The numeric id of the VPC the share is served on. The platform does not report it back, so the " +
					"value in the configuration is kept, and an imported share takes it from the configuration at the next apply. " +
					"Changing it replaces the share.",
				Required:   true,
				Validators: []validator.Int64{int64validator.AtLeast(1)},
				PlanModifiers: []planmodifier.Int64{int64planmodifier.RequiresReplaceIf(
					func(_ context.Context, req planmodifier.Int64Request, resp *int64planmodifier.RequiresReplaceIfFuncResponse) {
						resp.RequiresReplace = !req.StateValue.IsNull()
					},
					"Replaces the share when the VPC changes; an imported share has no VPC id to compare.",
					"Replaces the share when the VPC changes; an imported share has no VPC id to compare.",
				)},
			},
			"size_gb": schema.Int64Attribute{
				MarkdownDescription: "The share size in GB (10 to 65536). Growing needs quota; shrinking is refused when the share holds more data than the new size.",
				Required:            true,
				Validators:          []validator.Int64{int64validator.Between(10, 65536)},
			},
			"access": schema.ListNestedAttribute{
				MarkdownDescription: "Who may mount the share: 1 to 64 entries, each a server (`vm_identifier`) or a network (`cidr`).",
				Required:            true,
				Validators:          []validator.List{listvalidator.SizeBetween(1, 64)},
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"vm_identifier": schema.StringAttribute{
							MarkdownDescription: "The identifier of a server in the VPC. Set this or `cidr`.",
							Optional:            true,
							Validators:          []validator.String{stringvalidator.ExactlyOneOf(path.MatchRelative().AtParent().AtName("cidr"))},
						},
						"cidr": schema.StringAttribute{
							MarkdownDescription: "An IPv4 network such as `10.0.0.0/24`. Set this or `vm_identifier`.",
							Optional:            true,
						},
						"mode": schema.StringAttribute{
							MarkdownDescription: "`rw` (the default) or `ro`.",
							Optional:            true,
							Computed:            true,
							Default:             stringdefault.StaticString("rw"),
							Validators:          []validator.String{stringvalidator.OneOf("rw", "ro")},
						},
					},
				},
			},
			"state": schema.StringAttribute{MarkdownDescription: "The share state; `ready` once a change has finished.", Computed: true},
			"gateway_vpc_ip": schema.StringAttribute{
				MarkdownDescription: "The address inside the VPC that serves the share, once the share has a gateway.",
				Computed:            true,
			},
			"mount_nfs4": schema.StringAttribute{
				MarkdownDescription: "The `mount` command for NFSv4.2 clients, once the share is ready.",
				Computed:            true,
			},
			"created_at": computed("When the share was created (RFC 3339)."),
		},
	}
}

func (r *shareResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func accessOf(models []accessModel) []accessEntry {
	out := make([]accessEntry, 0, len(models))
	for _, m := range models {
		out = append(out, accessEntry{
			VMIdentifier: m.VMIdentifier.ValueString(),
			CIDR:         m.CIDR.ValueString(),
			Mode:         m.Mode.ValueString(),
		})
	}
	return out
}

func accessModels(entries []accessEntry) []accessModel {
	out := make([]accessModel, 0, len(entries))
	for _, e := range entries {
		m := accessModel{VMIdentifier: types.StringNull(), CIDR: types.StringNull(), Mode: types.StringValue(e.Mode)}
		if e.VMIdentifier != "" {
			m.VMIdentifier = types.StringValue(e.VMIdentifier)
		}
		if e.CIDR != "" {
			m.CIDR = types.StringValue(e.CIDR)
		}
		out = append(out, m)
	}
	return out
}

// applyShare copies what the platform reports. The VPC id is kept as given.
func applyShare(m *shareModel, s *share) {
	m.Identifier = types.StringValue(s.Identifier)
	m.Name = types.StringValue(s.Name)
	m.DcIdentifier = types.StringValue(s.DcIdentifier)
	m.SizeGb = types.Int64Value(s.SizeGb)
	m.Access = orderLike(m.Access, accessModels(s.Access))
	m.State = types.StringValue(s.State)
	m.CreatedAt = types.StringValue(s.CreatedAt)
	m.GatewayIP, m.MountNFS4 = types.StringNull(), types.StringNull()
	if s.Gateway != nil && s.Gateway.VPCIP != "" {
		m.GatewayIP = types.StringValue(s.Gateway.VPCIP)
	}
	if s.Mount != nil && s.Mount.NFS4 != "" {
		m.MountNFS4 = types.StringValue(s.Mount.NFS4)
	}
}

func (r *shareResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan shareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	created, err := r.api.create(ctx, createRequest{
		Name:         plan.Name.ValueString(),
		DcIdentifier: plan.DcIdentifier.ValueString(),
		VPCID:        plan.VPCID.ValueInt64(),
		SizeGb:       plan.SizeGb.ValueInt64(),
		Access:       accessOf(plan.Access),
	})
	if err != nil {
		resp.Diagnostics.AddError("Error creating file share", err.Error())
		return
	}
	// Save the identifier first so a failed wait leaves a share Terraform can find and delete.
	plan.Identifier = types.StringValue(created.Identifier)
	plan.State, plan.GatewayIP, plan.MountNFS4, plan.CreatedAt = types.StringValue(created.State), types.StringNull(), types.StringNull(), types.StringValue(created.CreatedAt)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
	ready, err := r.api.waitReady(ctx, created.Identifier)
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for file share", err.Error())
		return
	}
	applyShare(&plan, ready)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *shareResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state shareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	s, err := r.api.get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading file share", err.Error())
		return
	}
	if s == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	applyShare(&state, s)
	// While a resize is running the platform still reports the old size; keep what was asked.
	if s.ResizingToGb != nil {
		state.SizeGb = types.Int64Value(*s.ResizingToGb)
	}
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *shareResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state shareModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Identifier.ValueString()
	plan.Identifier = state.Identifier
	if plan.SizeGb.ValueInt64() != state.SizeGb.ValueInt64() {
		if err := r.api.resize(ctx, id, plan.SizeGb.ValueInt64()); err != nil {
			resp.Diagnostics.AddError("Error resizing file share", err.Error())
			return
		}
		if _, err := r.api.waitReady(ctx, id); err != nil {
			resp.Diagnostics.AddError("Error waiting for file share resize", err.Error())
			return
		}
	}
	if !accessEqual(plan.Access, state.Access) {
		if err := r.api.setAccess(ctx, id, accessOf(plan.Access)); err != nil {
			resp.Diagnostics.AddError("Error changing file share access", err.Error())
			return
		}
	}
	s, err := r.api.waitReady(ctx, id)
	if err != nil {
		resp.Diagnostics.AddError("Error reading file share after update", err.Error())
		return
	}
	applyShare(&plan, s)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func accessEqual(a, b []accessModel) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !a[i].VMIdentifier.Equal(b[i].VMIdentifier) || !a[i].CIDR.Equal(b[i].CIDR) || !a[i].Mode.Equal(b[i].Mode) {
			return false
		}
	}
	return true
}

func (r *shareResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state shareModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Identifier.ValueString()
	if err := r.api.remove(ctx, id, state.Name.ValueString()); err != nil {
		if s, getErr := r.api.get(ctx, id); getErr == nil && s == nil {
			return
		}
		resp.Diagnostics.AddError(fmt.Sprintf("Error deleting file share %q", state.Name.ValueString()), err.Error())
		return
	}
	if err := r.api.waitGone(ctx, id); err != nil {
		resp.Diagnostics.AddError("Error waiting for file share deletion", err.Error())
	}
}

func (r *shareResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

// orderLike returns the platform's access entries in the order of the
// configured ones, so a different answer order is not drift. Entries the
// configuration does not have follow in the platform's order.
func orderLike(configured, reported []accessModel) []accessModel {
	key := func(a accessModel) string {
		return a.VMIdentifier.ValueString() + "|" + a.CIDR.ValueString()
	}
	byKey := make(map[string]accessModel, len(reported))
	for _, a := range reported {
		byKey[key(a)] = a
	}
	out := make([]accessModel, 0, len(reported))
	for _, c := range configured {
		if a, ok := byKey[key(c)]; ok {
			out = append(out, a)
			delete(byKey, key(c))
		}
	}
	for _, a := range reported {
		if _, ok := byKey[key(a)]; ok {
			out = append(out, a)
		}
	}
	return out
}
