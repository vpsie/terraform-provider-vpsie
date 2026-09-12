package bucket

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

type bucketDataSource struct {
	client *govpsie.Client
}

type bucketDataSourceModel struct {
	ID      types.String  `tfsdk:"id"`
	Buckets []bucketModel `tfsdk:"buckets"`
}

type bucketModel struct {
	Identifier  types.String `tfsdk:"identifier"`
	BucketName  types.String `tfsdk:"bucket_name"`
	AccessKey   types.String `tfsdk:"access_key"`
	SecretKey   types.String `tfsdk:"secret_key"`
	EndPoint    types.String `tfsdk:"endpoint"`
	State       types.String `tfsdk:"state"`
	Country     types.String `tfsdk:"country"`
	CreatedBy   types.String `tfsdk:"created_by"`
	CreatedOn   types.String `tfsdk:"created_on"`
	ProjectName types.String `tfsdk:"project_name"`
}

func NewBucketDataSource() datasource.DataSource {
	return &bucketDataSource{}
}

func (d *bucketDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_buckets"
}

func (d *bucketDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"buckets": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"bucket_name": schema.StringAttribute{
							Computed: true,
						},
						"access_key": schema.StringAttribute{
							Computed:  true,
							Sensitive: true,
						},
						"secret_key": schema.StringAttribute{
							Computed:  true,
							Sensitive: true,
						},
						"endpoint": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"country": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"project_name": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

func (d *bucketDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *bucketDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state bucketDataSourceModel

	buckets, err := d.client.Bucket.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Buckets",
			"Could not get buckets, unexpected error: "+err.Error(),
		)
		return
	}

	for _, b := range buckets {
		state.Buckets = append(state.Buckets, bucketModel{
			Identifier:  types.StringValue(b.Identifier),
			BucketName:  types.StringValue(b.BucketName),
			AccessKey:   types.StringValue(b.AccessKey),
			SecretKey:   types.StringValue(b.SecretKey),
			EndPoint:    types.StringValue(b.EndPoint),
			State:       types.StringValue(b.State),
			Country:     types.StringValue(b.Country),
			CreatedBy:   types.StringValue(b.CreatedBy),
			CreatedOn:   types.StringValue(b.CreatedOn),
			ProjectName: types.StringValue(b.ProjectName),
		})
	}

	state.ID = types.StringValue("buckets")

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
