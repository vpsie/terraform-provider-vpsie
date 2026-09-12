package tag

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &tagDataSource{}
	_ datasource.DataSourceWithConfigure = &tagDataSource{}
)

type tagDataSource struct {
	client *govpsie.Client
}

type tagDataSourceModel struct {
	Tags []tagModel   `tfsdk:"tags"`
	ID   types.String `tfsdk:"id"`
}

type tagModel struct {
	Identifier types.String `tfsdk:"identifier"`
	Name       types.String `tfsdk:"name"`
	Color      types.String `tfsdk:"color"`
}

// NewTagDataSource is a helper function to create the data source.
func NewTagDataSource() datasource.DataSource {
	return &tagDataSource{}
}

func (t *tagDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_tags"
}

func (t *tagDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all VPSie resource tags.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"tags": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"color": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (t *tagDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state tagDataSourceModel

	tags, err := t.client.Tags.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get tags",
			"An unexpected error occurred when getting tags: "+err.Error(),
		)

		return
	}

	for _, tag := range tags {
		state.Tags = append(state.Tags, tagModel{
			Identifier: types.StringValue(tag.Identifier),
			Name:       types.StringValue(tag.Name),
			Color:      types.StringValue(tag.Color),
		})
	}

	state.ID = types.StringValue("tags")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (t *tagDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	t.client = client
}
