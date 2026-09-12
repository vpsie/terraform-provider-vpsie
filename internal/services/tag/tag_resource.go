package tag

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
	_ resource.Resource                = &tagResource{}
	_ resource.ResourceWithConfigure   = &tagResource{}
	_ resource.ResourceWithImportState = &tagResource{}
)

type tagResource struct {
	client *govpsie.Client
}

type tagResourceModel struct {
	Entity             types.String `tfsdk:"entity"`
	ResourceIdentifier types.String `tfsdk:"resource_identifier"`
	Tags               types.List   `tfsdk:"tags"`
}

// NewTagResource is a helper function to create the resource.
func NewTagResource() resource.Resource {
	return &tagResource{}
}

func (t *tagResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tag"
}

func (t *tagResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Applies a set of tags to a VPSie resource (entity). Tags are labels " +
			"attached to an existing resource such as a server, VPC, storage volume or ssh key.",
		Attributes: map[string]schema.Attribute{
			"entity": schema.StringAttribute{
				MarkdownDescription: "The type of resource the tags are applied to (for example " +
					"`boxes`, `vpc`, `storages`, `ssh_keys`, `dns_domains`, `lbs`, `k8s`, " +
					"`container_registry`, `managed_db_clusters`).",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"resource_identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier of the resource the tags are applied to.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"tags": schema.ListAttribute{
				MarkdownDescription: "The list of tag names applied to the resource.",
				Required:            true,
				ElementType:         types.StringType,
			},
		},
	}
}

func (t *tagResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	t.client = client
}

func (t *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Use Edit (replace) so the resource is authoritative over the entity's tag
	// set: after apply the entity has exactly the configured tags, matching Read.
	err := t.client.Tags.Edit(ctx, plan.Entity.ValueString(), plan.ResourceIdentifier.ValueString(), tags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error applying tags",
			"couldn't apply tags to "+plan.Entity.ValueString()+" "+plan.ResourceIdentifier.ValueString()+": "+err.Error(),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (t *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	names, err := t.client.Tags.ListForEntity(ctx, state.Entity.ValueString(), state.ResourceIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tags",
			"couldn't read tags for "+state.Entity.ValueString()+" "+state.ResourceIdentifier.ValueString()+": "+err.Error(),
		)

		return
	}

	if len(names) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	list, diags := types.ListValueFrom(ctx, types.StringType, names)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	state.Tags = list

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (t *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var tags []string
	resp.Diagnostics.Append(plan.Tags.ElementsAs(ctx, &tags, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := t.client.Tags.Edit(ctx, plan.Entity.ValueString(), plan.ResourceIdentifier.ValueString(), tags)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating tags",
			"couldn't update tags for "+plan.Entity.ValueString()+" "+plan.ResourceIdentifier.ValueString()+": "+err.Error(),
		)

		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (t *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := t.client.Tags.Delete(ctx, state.Entity.ValueString(), state.ResourceIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error removing tags",
			"couldn't remove tags from "+state.Entity.ValueString()+" "+state.ResourceIdentifier.ValueString()+": "+err.Error(),
		)

		return
	}
}

// ImportState imports a tag attachment using the composite ID "entity,resource_identifier".
func (t *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.Split(req.ID, ",")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			`Expected import ID in the format "entity,resource_identifier".`,
		)

		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("entity"), parts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("resource_identifier"), parts[1])...)
}
