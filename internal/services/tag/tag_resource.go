package tag

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
	_ resource.Resource                = &tagResource{}
	_ resource.ResourceWithConfigure   = &tagResource{}
	_ resource.ResourceWithImportState = &tagResource{}
)

type tagResource struct {
	client *govpsie.Client
}

type tagResourceModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Color      types.String `tfsdk:"color"`
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
		MarkdownDescription: "Manages a VPSie resource tag (a reusable label with a name and color).",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the tag.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the tag.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"color": schema.StringAttribute{
				MarkdownDescription: "The color of the tag (for example a hex value like `#ff0000`).",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

// Create creates the resource and sets the initial Terraform state.
func (t *tagResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan tagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	identifier, err := t.client.Tags.Create(ctx, plan.Name.ValueString(), plan.Color.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating tag",
			"couldn't create tag, unexpected error: "+err.Error(),
		)

		return
	}

	// Some API versions do not return the identifier on create; fall back to a
	// lookup by name so the resource always ends up with a stable identifier.
	if identifier == "" {
		tag, lookupErr := t.getTagByName(ctx, plan.Name.ValueString())
		if lookupErr != nil {
			resp.Diagnostics.AddError(
				"Error creating tag",
				"tag created but couldn't resolve its identifier: "+lookupErr.Error(),
			)

			return
		}
		identifier = tag.Identifier
	}

	plan.Identifier = types.StringValue(identifier)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (t *tagResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state tagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tag, err := t.getTagByIdentifier(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading tag",
			"couldn't read tag "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}

	if tag == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	state.Name = types.StringValue(tag.Name)
	state.Color = types.StringValue(tag.Color)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is a no-op: name and color changes force replacement.
func (t *tagResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan tagResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete removes the tag from Terraform state. The VPSie API exposes no
// endpoint to delete a tag definition (only endpoints to detach a tag from a
// specific resource), so the tag definition itself remains in the account. A
// warning is emitted to make this explicit.
func (t *tagResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state tagResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.AddWarning(
		"Tag definition not deleted",
		"The VPSie API does not support deleting a tag definition. The tag "+
			state.Identifier.ValueString()+" has been removed from Terraform state but still "+
			"exists in your VPSie account and can be removed from the console.",
	)
}

func (t *tagResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (t *tagResource) getTagByName(ctx context.Context, name string) (*govpsie.Tag, error) {
	tags, err := t.client.Tags.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range tags {
		if tags[i].Name == name {
			return &tags[i], nil
		}
	}

	return nil, fmt.Errorf("tag %q not found", name)
}

func (t *tagResource) getTagByIdentifier(ctx context.Context, identifier string) (*govpsie.Tag, error) {
	tags, err := t.client.Tags.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range tags {
		if tags[i].Identifier == identifier {
			return &tags[i], nil
		}
	}

	return nil, nil
}
