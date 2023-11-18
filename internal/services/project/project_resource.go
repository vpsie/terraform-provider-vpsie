package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	_ resource.Resource                = &projectResource{}
	_ resource.ResourceWithConfigure   = &projectResource{}
	_ resource.ResourceWithImportState = &projectResource{}
)

type projectResource struct {
	client *govpsie.Client
}

type projectResourceModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Identifier  types.String `tfsdk:"identifier"`
	CreatedOn   types.String `tfsdk:"created_on"`
	CreatedBy   types.Int64  `tfsdk:"created_by"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	IsDefault   types.Int64  `tfsdk:"is_default"`
}

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

func (i *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (i *projectResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{

			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"is_default": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"updated_at": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Computed: true,
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
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (i *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report  this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	i.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (p *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	createProjectReq := govpsie.CreateProjectRequest{
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
	}

	err := p.client.Project.Create(ctx, &createProjectReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating project", err.Error())
		return
	}

	project, err := p.GetProjectByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting project", err.Error())
		return
	}

	plan.ID = types.Int64Value(int64(project.ID))
	plan.Identifier = types.StringValue(project.Identifier)
	plan.CreatedOn = types.StringValue(project.CreatedOn)
	plan.CreatedBy = types.Int64Value(int64(project.CreatedBy))
	plan.Description = types.StringValue(project.Description)
	plan.Name = types.StringValue(project.Name)
	plan.UpdatedAt = types.StringValue(project.UpdatedAt)
	plan.IsDefault = types.Int64Value(int64(project.IsDefault))

	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (p *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project, err := p.client.Project.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		if err.Error() == "project not found" {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading vpsie project",
			"couldn't read vpsie project identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.Int64Value(int64(project.ID))
	state.Identifier = types.StringValue(project.Identifier)
	state.CreatedOn = types.StringValue(project.CreatedOn)
	state.CreatedBy = types.Int64Value(int64(project.CreatedBy))
	state.Description = types.StringValue(project.Description)
	state.Name = types.StringValue(project.Name)
	state.UpdatedAt = types.StringValue(project.UpdatedAt)
	state.IsDefault = types.Int64Value(int64(project.IsDefault))

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (i *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (p *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := p.client.Project.Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project",
			"couldn't delete project, unexpected error: "+err.Error(),
		)

		return
	}
}

func (p *projectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (p *projectResource) GetProjectByName(ctx context.Context, name string) (*govpsie.Project, error) {
	projects, err := p.client.Project.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, project := range projects {
		if name == project.Name {
			project, err := p.client.Project.Get(ctx, project.Identifier)
			if err != nil {
				return nil, err
			}

			return project, nil
		}
	}

	return nil, fmt.Errorf("project with name %s not found", name)
}
