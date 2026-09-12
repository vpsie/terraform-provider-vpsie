package registry

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &registryDataSource{}
	_ datasource.DataSourceWithConfigure = &registryDataSource{}
)

type registryDataSource struct {
	client *govpsie.Client
}

type registryDataSourceModel struct {
	Registries []registryModel `tfsdk:"registries"`
	ID         types.String    `tfsdk:"id"`
}

type registryModel struct {
	Identifier     types.String `tfsdk:"identifier"`
	Name           types.String `tfsdk:"name"`
	DatacenterName types.String `tfsdk:"datacenter_name"`
	Status         types.String `tfsdk:"status"`
	CreatedOn      types.String `tfsdk:"created_on"`
}

// NewRegistryDataSource is a helper function to create the data source.
func NewRegistryDataSource() datasource.DataSource {
	return &registryDataSource{}
}

func (r *registryDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_registries"
}

func (r *registryDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all VPSie container registries.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"registries": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":      schema.StringAttribute{Computed: true},
						"name":            schema.StringAttribute{Computed: true},
						"datacenter_name": schema.StringAttribute{Computed: true},
						"status":          schema.StringAttribute{Computed: true},
						"created_on":      schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (r *registryDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state registryDataSourceModel

	registries, err := r.client.Registry.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get registries",
			"An unexpected error occurred when getting registries: "+err.Error(),
		)

		return
	}

	for _, registry := range registries {
		state.Registries = append(state.Registries, registryModel{
			Identifier:     types.StringValue(registry.Identifier),
			Name:           types.StringValue(registry.Name),
			DatacenterName: types.StringValue(registry.DatacenterName),
			Status:         types.StringValue(registry.Status),
			CreatedOn:      types.StringValue(registry.CreatedOn),
		})
	}

	state.ID = types.StringValue("registries")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (r *registryDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	r.client = client
}
