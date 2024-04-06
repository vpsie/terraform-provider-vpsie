package kubernetes

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type kubernetesDataSource struct {
	client *govpsie.Client
}

type kubernetesDataSourceModel struct {
	Kubernetes []kubernetesModel `tfsdk:"kubernetes"`
	ID         types.String      `tfsdk:"id"`
}

type kubernetesModel struct {
	ClusterName  types.String  `tfsdk:"cluster_name"`
	Identifier   types.String  `tfsdk:"identifier"`
	MasterCount  types.Int64   `tfsdk:"count"`
	CreatedOn    types.String  `tfsdk:"created_on"`
	UpdatedOn    types.String  `tfsdk:"updated_on"`
	CreatedBy    types.String  `tfsdk:"created_by"`
	NickName     types.String  `tfsdk:"nickname"`
	Cpu          types.Int64   `tfsdk:"cpu"`
	Ram          types.Int64   `tfsdk:"ram"`
	Traffic      types.Int64   `tfsdk:"traffic"`
	Color        types.String  `tfsdk:"color"`
	Price        types.Float64 `tfsdk:"price"`
	ManagerCount types.Int64   `tfsdk:"manager_count"`
	SlaveCount   types.Int64   `tfsdk:"slave_count"`
}

// NewKubernetesDataSource is a helper function to create the data source.
func NewKubernetesDataSource() datasource.DataSource {
	return &kubernetesDataSource{}
}

// Metadata returns the data source type name.
func (i *kubernetesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kubernetes"
}

// Schema defines the schema for the data source.
func (i *kubernetesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"kubernetes": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"cluster_name": schema.StringAttribute{
							Computed: true,
						},
						"master_count": schema.Int64Attribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"updated_on": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
						"nickname": schema.StringAttribute{
							Computed: true,
						},
						"cpu": schema.Int64Attribute{
							Computed: true,
						},
						"ram": schema.Int64Attribute{
							Computed: true,
						},
						"traffic": schema.Int64Attribute{
							Computed: true,
						},
						"price": schema.Float64Attribute{
							Computed: true,
						},
						"manager_count": schema.Int64Attribute{
							Computed: true,
						},
						"slave_count": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (k *kubernetesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state kubernetesDataSourceModel

	kubernetes, err := k.client.K8s.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Kubernetes",
			"Could not get Kubernetes, unexpected error: "+err.Error(),
		)

		return
	}

	for _, k8s := range kubernetes {
		k8sState := kubernetesModel{
			Identifier:   types.StringValue(k8s.Identifier),
			ClusterName:  types.StringValue(k8s.ClusterName),
			MasterCount:  types.Int64Value(int64(k8s.Count)),
			CreatedOn:    types.StringValue(k8s.CreatedOn),
			UpdatedOn:    types.StringValue(k8s.UpdatedOn),
			CreatedBy:    types.StringValue(k8s.CreatedBy),
			NickName:     types.StringValue(k8s.NickName),
			Cpu:          types.Int64Value(int64(k8s.Cpu)),
			Ram:          types.Int64Value(int64(k8s.Ram)),
			Traffic:      types.Int64Value(int64(k8s.Traffic)),
			Color:        types.StringValue(k8s.Color),
			Price:        types.Float64Value(k8s.Price),
			ManagerCount: types.Int64Value(int64(k8s.ManagerCount)),
			SlaveCount:   types.Int64Value(int64(k8s.SlaveCount)),
		}

		state.Kubernetes = append(state.Kubernetes, k8sState)
	}
}
func (k *kubernetesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	k.client = client
}
