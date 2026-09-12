package servergroup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &serverGroupDataSource{}
	_ datasource.DataSourceWithConfigure = &serverGroupDataSource{}
)

type serverGroupDataSource struct {
	client *govpsie.Client
}

type serverGroupDataSourceModel struct {
	ServerGroups []serverGroupModel `tfsdk:"server_groups"`
	ID           types.String       `tfsdk:"id"`
}

type serverGroupModel struct {
	Identifier       types.String `tfsdk:"identifier"`
	GroupName        types.String `tfsdk:"group_name"`
	GroupDescription types.String `tfsdk:"group_description"`
	IsDistributed    types.Bool   `tfsdk:"is_distributed"`
	VMCounts         types.Int64  `tfsdk:"vm_counts"`
	CreatedOn        types.String `tfsdk:"created_on"`
}

// NewServerGroupDataSource is a helper function to create the data source.
func NewServerGroupDataSource() datasource.DataSource {
	return &serverGroupDataSource{}
}

func (s *serverGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_groups"
}

func (s *serverGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Lists all VPSie server (VM) groups.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"server_groups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier":        schema.StringAttribute{Computed: true},
						"group_name":        schema.StringAttribute{Computed: true},
						"group_description": schema.StringAttribute{Computed: true},
						"is_distributed":    schema.BoolAttribute{Computed: true},
						"vm_counts":         schema.Int64Attribute{Computed: true},
						"created_on":        schema.StringAttribute{Computed: true},
					},
				},
			},
		},
	}
}

func (s *serverGroupDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serverGroupDataSourceModel

	groups, err := s.client.ServerGroup.List(ctx)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get server groups",
			"An unexpected error occurred when getting server groups: "+err.Error(),
		)

		return
	}

	for _, group := range groups {
		state.ServerGroups = append(state.ServerGroups, serverGroupModel{
			Identifier:       types.StringValue(group.Identifier),
			GroupName:        types.StringValue(group.GroupName),
			GroupDescription: types.StringValue(group.GroupDescription),
			IsDistributed:    types.BoolValue(group.IsDistributed),
			VMCounts:         types.Int64Value(group.VMCounts),
			CreatedOn:        types.StringValue(group.CreatedOn),
		})
	}

	state.ID = types.StringValue("server_groups")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (s *serverGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	s.client = client
}
