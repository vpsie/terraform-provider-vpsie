package manageddb

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &managedDBDataSource{}
	_ datasource.DataSourceWithConfigure = &managedDBDataSource{}
)

type managedDBDataSource struct {
	client *govpsie.Client
}

type managedDBDataSourceModel struct {
	ManagedDatabases []managedDBModel `tfsdk:"managed_databases"`
	ID               types.String     `tfsdk:"id"`
}

type managedDBModel struct {
	Identifier  types.String `tfsdk:"identifier"`
	ClusterName types.String `tfsdk:"cluster_name"`
	Nickname    types.String `tfsdk:"nickname"`
	NodesCount  types.Int64  `tfsdk:"nodes_count"`
	CPU         types.Int64  `tfsdk:"cpu"`
	RAM         types.Int64  `tfsdk:"ram"`
	Traffic     types.Int64  `tfsdk:"traffic"`
	CreatedOn   types.String `tfsdk:"created_on"`
}

// NewManagedDatabaseDataSource is a helper function to create the data source.
func NewManagedDatabaseDataSource() datasource.DataSource {
	return &managedDBDataSource{}
}

func (m *managedDBDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_managed_databases"
}

func (m *managedDBDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all VPSie managed database clusters.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"managed_databases": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":   schema.StringAttribute{Computed: true},
						"cluster_name": schema.StringAttribute{Computed: true},
						"nickname":     schema.StringAttribute{Computed: true},
						"nodes_count":  schema.Int64Attribute{Computed: true},
						"cpu":          schema.Int64Attribute{Computed: true},
						"ram":          schema.Int64Attribute{Computed: true},
						"traffic":      schema.Int64Attribute{Computed: true},
						"created_on":   schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (m *managedDBDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state managedDBDataSourceModel

	clusters, err := m.client.ManagedDB.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get managed databases",
			"An unexpected error occurred when getting managed databases: "+err.Error(),
		)

		return
	}

	for _, cluster := range clusters {
		state.ManagedDatabases = append(state.ManagedDatabases, managedDBModel{
			Identifier:  types.StringValue(cluster.Identifier),
			ClusterName: types.StringValue(cluster.ClusterName),
			Nickname:    types.StringValue(cluster.Nickname),
			NodesCount:  types.Int64Value(cluster.NodesCount),
			CPU:         types.Int64Value(cluster.CPU),
			RAM:         types.Int64Value(cluster.RAM),
			Traffic:     types.Int64Value(cluster.Traffic),
			CreatedOn:   types.StringValue(cluster.CreatedOn),
		})
	}

	state.ID = types.StringValue("managed_databases")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (m *managedDBDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	m.client = client
}
