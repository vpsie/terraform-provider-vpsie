package kubernetes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

type kubernetesGroupDataSource struct {
	client *govpsie.Client
}

type kubernetesGroupDataSourceModel struct {
	KubernetesGroups []kubernetesGroupModel `tfsdk:"kubernetes_group"`
	ID               types.String           `tfsdk:"id"`
}

type kubernetesGroupModel struct {
	ID           types.Int64  `tfsdk:"id"`
	GroupName    types.String `tfsdk:"group_name"`
	UserID       types.Int64  `tfsdk:"user_id"`
	BoxsizeID    types.Int64  `tfsdk:"boxsize_id"`
	DatacenterID types.Int64  `tfsdk:"datacenter_id"`
	RAM          types.Int64  `tfsdk:"ram"`
	CPU          types.Int64  `tfsdk:"cpu"`
	Ssd          types.Int64  `tfsdk:"ssd"`
	Traffic      types.Int64  `tfsdk:"traffic"`
	Notes        types.String `tfsdk:"notes"`
	CreatedOn    types.String `tfsdk:"created_on"`
	LastUpdated  types.String `tfsdk:"last_updated"`
	DroppedOn    types.String `tfsdk:"dropped_on"`
	IsActive     types.Int64  `tfsdk:"is_active"`
	IsDeleted    types.Int64  `tfsdk:"is_deleted"`
	Identifier   types.String `tfsdk:"identifier"`
	ProjectID    types.Int64  `tfsdk:"project_id"`
	ClusterID    types.Int64  `tfsdk:"cluster_id"`
	NodesCount   types.Int64  `tfsdk:"nodes_count"`
	DcIdentifier types.String `tfsdk:"dc_identifier"`
}

// NewKubernetesGroupDataSource is is a helper function to create the data source.
func NewKubernetesGroupDataSource() datasource.DataSource {
	return &kubernetesGroupDataSource{}
}

// Metadata returns the data source type name.
func (k *kubernetesGroupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes_group"
}

// Schema defines the schema for the data source.
func (k *kubernetesGroupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"kubernetes_group": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"group_name": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"boxsize_id": schema.Int64Attribute{
							Computed: true,
						},
						"datacenter_id": schema.Int64Attribute{
							Computed: true,
						},
						"ram": schema.Int64Attribute{
							Computed: true,
						},
						"cpu": schema.Int64Attribute{
							Computed: true,
						},
						"ssd": schema.Int64Attribute{
							Computed: true,
						},
						"traffic": schema.Int64Attribute{
							Computed: true,
						},
						"notes": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
						"dropped_on": schema.StringAttribute{
							Computed: true,
						},
						"is_active": schema.Int64Attribute{
							Computed: true,
						},
						"is_deleted": schema.Int64Attribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"project_id": schema.Int64Attribute{
							Computed: true,
						},
						"cluster_id": schema.Int64Attribute{
							Computed: true,
						},
						"nodes_count": schema.Int64Attribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (k *kubernetesGroupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state kubernetesGroupDataSourceModel

	k8s, err := k.client.K8s.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Kubernetes",
			"Could not get Kubernetes, unexpected error: "+err.Error(),
		)

		return
	}

	for _, k8 := range k8s {

		k8sGroups, err := k.client.K8s.ListK8sGroups(ctx, k8.Identifier)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error Getting Kubernetes Groups",
				"Could not get Kubernetes Groups, unexpected error: "+err.Error(),
			)
		}

		for _, k8Group := range k8sGroups {
			state.KubernetesGroups = append(state.KubernetesGroups, kubernetesGroupModel{
				ID:           types.Int64Value(k8Group.ID),
				GroupName:    types.StringValue(k8Group.GroupName),
				UserID:       types.Int64Value(k8Group.UserID),
				BoxsizeID:    types.Int64Value(k8Group.BoxsizeID),
				DatacenterID: types.Int64Value(k8Group.DatacenterID),
				RAM:          types.Int64Value(k8Group.RAM),
				CPU:          types.Int64Value(k8Group.CPU),
				Ssd:          types.Int64Value(k8Group.Ssd),
				Traffic:      types.Int64Value(k8Group.Traffic),
				Notes:        types.StringValue(k8Group.Notes),
				CreatedOn:    types.StringValue(k8Group.CreatedOn.String()),
				LastUpdated:  types.StringValue(k8Group.LastUpdated.String()),
				DroppedOn:    types.StringValue(k8Group.DroppedOn.String()),
				IsActive:     types.Int64Value(k8Group.IsActive),
				IsDeleted:    types.Int64Value(k8Group.IsDeleted),
				Identifier:   types.StringValue(k8Group.Identifier),
				ProjectID:    types.Int64Value(k8Group.ProjectID),
				ClusterID:    types.Int64Value(k8Group.ClusterID),
				NodesCount:   types.Int64Value(k8Group.NodesCount),
				DcIdentifier: types.StringValue(k8Group.DcIdentifier),
			})
		}

	}
}

func (k *kubernetesGroupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	k.client = client
}
