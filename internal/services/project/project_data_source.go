package project

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type projectDataSource struct {
	client *govpsie.Client
}

type projectDataSourceModel struct {
	Projects []projectsModel `tfsdk:"projects"`
	ID       types.String    `tfsdk:"id"`
}

type projectsModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Identifier  types.String `tfsdk:"identifier"`
	CreatedOn   types.String `tfsdk:"created_on"`
	CreatedBy   types.Int64  `tfsdk:"created_by"`
	Description types.String `tfsdk:"description"`
	Name        types.String `tfsdk:"name"`
	UpdatedAt   types.String `tfsdk:"updated_at"`
	IsDefault   types.Int64  `tfsdk:"is_default"`
}

// NewProjectDataSource is a helper function to create the data source.
func NewProjectDataSource() datasource.DataSource {
	return &projectDataSource{}
}

// Metadata returns the data source type name.
func (p *projectDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_projects"
}

// Schema defines the schema for the data source.
func (p *projectDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"projects": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"is_default": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"updated_at": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (p *projectDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state projectDataSourceModel

	projects, err := p.client.Project.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Projects",
			"Could not get Projects, unexpected error: "+err.Error(),
		)

		return
	}

	for _, project := range projects {
		projectState := projectsModel{
			ID:          types.Int64Value(int64(project.ID)),
			Identifier:  types.StringValue(project.Identifier),
			CreatedOn:   types.StringValue(project.CreatedOn),
			CreatedBy:   types.Int64Value(int64(project.CreatedBy)),
			Description: types.StringValue(project.Description),
			Name:        types.StringValue(project.Name),
			UpdatedAt:   types.StringValue(project.UpdatedAt),
			IsDefault:   types.Int64Value(int64(project.IsDefault)),
		}

		state.Projects = append(state.Projects, projectState)
	}
}
func (p *projectDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
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

	p.client = client
}
