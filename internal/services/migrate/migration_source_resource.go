package migrate

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/mapplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &migrationSourceResource{}
	_ resource.ResourceWithConfigure   = &migrationSourceResource{}
	_ resource.ResourceWithImportState = &migrationSourceResource{}
)

type migrationSourceResource struct {
	api *sourceAPI
}

type migrationSourceModel struct {
	Identifier           types.String `tfsdk:"identifier"`
	Type                 types.String `tfsdk:"type"`
	Name                 types.String `tfsdk:"name"`
	DatacenterIdentifier types.String `tfsdk:"datacenter_identifier"`
	Settings             types.Map    `tfsdk:"settings"`
	Credentials          types.Map    `tfsdk:"credentials"`
	Host                 types.String `tfsdk:"host"`
	Username             types.String `tfsdk:"username"`
	Status               types.String `tfsdk:"status"`
	CreatedAt            types.String `tfsdk:"created_at"`
}

func NewMigrationSourceResource() resource.Resource {
	return &migrationSourceResource{}
}

func (r *migrationSourceResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_migration_source"
}

func (r *migrationSourceResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	computed := func(desc string) schema.StringAttribute {
		return schema.StringAttribute{
			MarkdownDescription: desc,
			Computed:            true,
			PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
		}
	}
	resp.Schema = schema.Schema{
		MarkdownDescription: "A Migrate source: the platform a migration reads servers from (for example VMware ESXi or vCenter, " +
			"Nutanix AHV, an image file or an SSH-reachable host). The platform registers the source when it is created, " +
			"which tests the connection; a source that cannot be reached is not created. SSH-based sources " +
			"(`kvm-libvirt`, `any-server-rescue`, `xcp-ng`) also need their host key confirmed in the portal before a job can run.",
		Attributes: map[string]schema.Attribute{
			"identifier": computed("The source identifier."),
			"type": schema.StringAttribute{
				MarkdownDescription: "The source type, for example `vmware-esxi`, `vmware-vcenter`, `image-file`, `proxmox-ve`, " +
					"`proxmox-pbs`, `xcp-ng`, `nutanix-ahv`, `kvm-libvirt`, `any-server-rescue` or `openstack`. The types " +
					"available to the account are listed by `GET /migrate/source-types`. Changing it replaces the source.",
				Required:      true,
				PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The source name. Changing it replaces the source.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"datacenter_identifier": schema.StringAttribute{
				MarkdownDescription: "The datacenter whose nodes run the source's jobs. Changing it replaces the source.",
				Required:            true,
				PlanModifiers:       []planmodifier.String{stringplanmodifier.RequiresReplace()},
			},
			"settings": schema.MapAttribute{
				MarkdownDescription: "The connector's non-secret fields as strings (for example `host`, `port`, `username`). " +
					"Which fields a type takes is defined by the platform per type. Changing a setting replaces the source.",
				Optional:      true,
				ElementType:   types.StringType,
				PlanModifiers: []planmodifier.Map{mapplanmodifier.RequiresReplace()},
			},
			"credentials": schema.MapAttribute{
				MarkdownDescription: "The connector's secret fields as strings (for example `password`). Stored sensitive in " +
					"the state and never read back from the platform. Changing them rotates the credential in place after the " +
					"platform has tested the new one; give every secret field of the type, as the platform expects.",
				Optional:    true,
				Sensitive:   true,
				ElementType: types.StringType,
			},
			"host":       computed("The host the source connects to, as the platform records it."),
			"username":   computed("The user the source signs in as, as the platform records it."),
			"status":     schema.StringAttribute{MarkdownDescription: "The source status.", Computed: true},
			"created_at": computed("When the source was created (RFC 3339)."),
		},
	}
}

func (r *migrationSourceResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.api = &sourceAPI{client: client}
}

func stringMap(ctx context.Context, m types.Map) (map[string]string, error) {
	out := map[string]string{}
	if m.IsNull() || m.IsUnknown() {
		return out, nil
	}
	if diags := m.ElementsAs(ctx, &out, false); diags.HasError() {
		return nil, fmt.Errorf("%s", diags.Errors()[0].Detail())
	}
	return out, nil
}

func (r *migrationSourceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan migrationSourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}
	settings, err := stringMap(ctx, plan.Settings)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("settings"), "Invalid settings", err.Error())
		return
	}
	credentials, err := stringMap(ctx, plan.Credentials)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("credentials"), "Invalid credentials", err.Error())
		return
	}
	body, err := createBody(plan.Type.ValueString(), plan.Name.ValueString(), plan.DatacenterIdentifier.ValueString(), settings, credentials)
	if err != nil {
		resp.Diagnostics.AddError("Invalid migration source", err.Error())
		return
	}
	src, err := r.api.create(ctx, body)
	if err != nil {
		resp.Diagnostics.AddError("Error creating migration source", err.Error())
		return
	}
	applySource(&plan, src)
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

// applySource copies what the platform reports. Settings and credentials stay
// as configured: the platform returns neither in full.
func applySource(m *migrationSourceModel, src *migrationSource) {
	m.Identifier = types.StringValue(src.Identifier)
	m.Host = types.StringValue(src.Host)
	m.Username = types.StringValue(src.Username)
	m.Status = types.StringValue(src.Status)
	m.CreatedAt = types.StringValue(src.CreatedAt)
}

func (r *migrationSourceResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state migrationSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	src, err := r.api.get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading migration source", err.Error())
		return
	}
	if src == nil {
		resp.State.RemoveResource(ctx)
		return
	}
	state.Type = types.StringValue(src.Type)
	state.Name = types.StringValue(src.Name)
	if src.DatacenterIdentifier != "" {
		state.DatacenterIdentifier = types.StringValue(src.DatacenterIdentifier)
	}
	applySource(&state, src)
	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *migrationSourceResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state migrationSourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	credentials, err := stringMap(ctx, plan.Credentials)
	if err != nil {
		resp.Diagnostics.AddAttributeError(path.Root("credentials"), "Invalid credentials", err.Error())
		return
	}
	if len(credentials) == 0 {
		resp.Diagnostics.AddAttributeError(path.Root("credentials"), "Credentials cannot be removed",
			"A source keeps its credential; to change a source without credentials, replace it.")
		return
	}
	if err := r.api.rotateCredentials(ctx, state.Identifier.ValueString(), credentials); err != nil {
		resp.Diagnostics.AddError(fmt.Sprintf("Error rotating credentials (%s)", sortedKeys(credentials)), err.Error())
		return
	}
	plan.Identifier = state.Identifier
	plan.Host, plan.Username, plan.Status, plan.CreatedAt = state.Host, state.Username, state.Status, state.CreatedAt
	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *migrationSourceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state migrationSourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	if err := r.api.delete(ctx, state.Identifier.ValueString()); err != nil {
		src, getErr := r.api.get(ctx, state.Identifier.ValueString())
		if getErr == nil && src == nil {
			return
		}
		resp.Diagnostics.AddError("Error deleting migration source", err.Error())
	}
}

func (r *migrationSourceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}
