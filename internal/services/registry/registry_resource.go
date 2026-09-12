package registry

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
	_ resource.Resource                = &registryResource{}
	_ resource.ResourceWithConfigure   = &registryResource{}
	_ resource.ResourceWithImportState = &registryResource{}
)

type registryResource struct {
	client *govpsie.Client
}

type registryResourceModel struct {
	Identifier           types.String `tfsdk:"identifier"`
	Name                 types.String `tfsdk:"name"`
	DatacenterIdentifier types.String `tfsdk:"datacenter_identifier"`
	PlanIdentifier       types.String `tfsdk:"plan_identifier"`
	Status               types.String `tfsdk:"status"`
	CreatedOn            types.String `tfsdk:"created_on"`
}

// NewRegistryResource is a helper function to create the resource.
func NewRegistryResource() resource.Resource {
	return &registryResource{}
}

func (r *registryResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registry"
}

func (r *registryResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a VPSie container registry.",
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				MarkdownDescription: "The unique identifier of the registry.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the registry.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datacenter_identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier of the datacenter that hosts the registry.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"plan_identifier": schema.StringAttribute{
				MarkdownDescription: "The identifier of the resource plan for the registry.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"status": schema.StringAttribute{
				MarkdownDescription: "The current status of the registry.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_on": schema.StringAttribute{
				MarkdownDescription: "The creation timestamp of the registry.",
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (r *registryResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *registryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan registryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Registry.Create(ctx, plan.Name.ValueString(), plan.DatacenterIdentifier.ValueString(), plan.PlanIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating registry",
			"couldn't create registry, unexpected error: "+err.Error(),
		)

		return
	}

	registry, err := r.getRegistryByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating registry",
			"registry created but couldn't resolve its identifier: "+err.Error(),
		)

		return
	}

	plan.Identifier = types.StringValue(registry.Identifier)
	plan.Status = types.StringValue(registry.Status)
	plan.CreatedOn = types.StringValue(registry.CreatedOn)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Read refreshes the Terraform state with the latest data.
func (r *registryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state registryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	registry, err := r.findRegistry(ctx, state.Identifier.ValueString(), state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading registry",
			"couldn't read registry "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}

	if registry == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	if registry.Name != "" {
		state.Name = types.StringValue(registry.Name)
	}
	if registry.Status != "" {
		state.Status = types.StringValue(registry.Status)
	}
	if registry.CreatedOn != "" {
		state.CreatedOn = types.StringValue(registry.CreatedOn)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// Update is a no-op: all configurable attributes force replacement.
func (r *registryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan registryResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *registryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state registryResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.Registry.Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting registry",
			"couldn't delete registry "+state.Identifier.ValueString()+": "+err.Error(),
		)

		return
	}
}

func (r *registryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (r *registryResource) getRegistryByName(ctx context.Context, name string) (*govpsie.Registry, error) {
	registries, err := r.client.Registry.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range registries {
		if registries[i].Name == name {
			return &registries[i], nil
		}
	}

	return nil, fmt.Errorf("registry %q not found", name)
}

// findRegistry matches a registry by identifier first, then falls back to name.
func (r *registryResource) findRegistry(ctx context.Context, identifier, name string) (*govpsie.Registry, error) {
	registries, err := r.client.Registry.List(ctx)
	if err != nil {
		return nil, err
	}

	for i := range registries {
		if identifier != "" && registries[i].Identifier == identifier {
			return &registries[i], nil
		}
	}

	for i := range registries {
		if name != "" && registries[i].Name == name {
			return &registries[i], nil
		}
	}

	return nil, nil
}
