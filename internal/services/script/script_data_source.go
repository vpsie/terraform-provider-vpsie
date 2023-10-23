package script

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
)

type scriptDataSource struct {
	client *govpsie.Client
}

type scriptDataSourceModel struct {
	Scripts []scriptsModel `tfsdk:"scripts"`
	ID      types.String   `tfsdk:"id"`
}

type scriptsModel struct {
	Identifier    types.String `tfsdk:"identifier"`
	UserID        types.Int64  `tfsdk:"user_id"`
	BoxID         types.Int64  `tfsdk:"box_id"`
	BoxIdentifier types.String `tfsdk:"box_identifier"`
	ScriptName    types.String `tfsdk:"script_name"`
	Script        types.String `tfsdk:"script"`
	CreatedOn     types.String `tfsdk:"created_on"`
	CreatedBy     types.String `tfsdk:"created_by"`
}

// NewScriptDataSource is a helper function to create the data source.
func NewScriptDataSource() datasource.DataSource {
	return &scriptDataSource{}
}

// Metadata returns the data source type name.
func (s *scriptDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_scripts"
}

// Schema defines the schema for the data source.
func (s *scriptDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"scripts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"box_id": schema.Int64Attribute{
							Computed: true,
						},
						"box_identifier": schema.StringAttribute{
							Computed: true,
						},
						"script_name": schema.StringAttribute{
							Computed: true,
						},
						"script": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *scriptDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state scriptDataSourceModel

	scripts, err := s.client.Scripts.GetScripts(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get scripts",
			"An unexpected error occurred when getting scripts: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Got scripts", map[string]interface{}{"scripts": scripts})

	for _, script := range scripts {

		scriptState := scriptsModel{
			UserID:        types.Int64Value(int64(script.UserID)),
			BoxID:         types.Int64Value(int64(script.BoxID)),
			BoxIdentifier: types.StringValue(script.BoxIdentifier),
			ScriptName:    types.StringValue(script.ScriptName),
			Script:        types.StringValue(script.Script),
			CreatedOn:     types.StringValue(script.CreatedOn.String()),
			Identifier:    types.StringValue(script.Identifier),
			CreatedBy:     types.StringValue(script.CreatedBy),
		}

		state.Scripts = append(state.Scripts, scriptState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *scriptDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	s.client = client
}
