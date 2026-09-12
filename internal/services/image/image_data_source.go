package image

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type imageDataSource struct {
	client *govpsie.Client
}

type imageDataSourceModel struct {
	Images []imagesModel `tfsdk:"images"`
	ID     types.String  `tfsdk:"id"`
}

type imagesModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Identifier     types.String `tfsdk:"identifier"`
	UserID         types.Int64  `tfsdk:"user_id"`
	DataCenterID   types.Int64  `tfsdk:"datacenter_id"`
	ImageSize      types.Int64  `tfsdk:"image_size"`
	OriginalName   types.String `tfsdk:"original_name"`
	FetchedFromUrl types.String `tfsdk:"fetched_from_url"`
	ImageHash      types.String `tfsdk:"image_hash"`
	ImageLabel     types.String `tfsdk:"image_label"`
	CreatedOn      types.String `tfsdk:"created_on"`
	Deleted        types.Int64  `tfsdk:"deleted"`
	DcName         types.String `tfsdk:"dc_name"`
	DcIdentifier   types.String `tfsdk:"dc_identifier"`
	CreatedBy      types.String `tfsdk:"created_by"`
}

// NewImageDataSource is a helper function to create the data source.
func NewImageDataSource() datasource.DataSource {
	return &imageDataSource{}
}

// Metadata returns the data source type name.
func (i *imageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_images"
}

// Schema defines the schema for the data source.
func (i *imageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"images": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"datacenter_id": schema.Int64Attribute{
							Computed: true,
						},
						"image_size": schema.Int64Attribute{
							Computed: true,
						},
						"original_name": schema.StringAttribute{
							Computed: true,
						},
						"fetched_from_url": schema.StringAttribute{
							Computed: true,
						},
						"image_hash": schema.StringAttribute{
							Computed: true,
						},
						"image_label": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"deleted": schema.Int64Attribute{
							Computed: true,
						},
						"dc_name": schema.StringAttribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
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
func (i *imageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state imageDataSourceModel

	images, err := i.client.Image.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Images",
			"Could not get Images, unexpected error: "+err.Error(),
		)

		return
	}

	for _, image := range images {
		imageState := imagesModel{
			ID:             types.Int64Value(int64(image.ID)),
			Identifier:     types.StringValue(image.Identifier),
			UserID:         types.Int64Value(int64(image.UserID)),
			DataCenterID:   types.Int64Value(int64(image.DatacenterID)),
			ImageSize:      types.Int64Value(image.ImageSize),
			OriginalName:   types.StringValue(image.OriginalName),
			FetchedFromUrl: types.StringValue(image.FetchedFromURL),
			ImageHash:      types.StringValue(image.ImageHash),
			ImageLabel:     types.StringValue(image.ImageLabel),
			CreatedOn:      types.StringValue(image.CreatedOn.String()),
			Deleted:        types.Int64Value(int64(image.Deleted)),
			DcName:         types.StringValue(image.DcName),
			DcIdentifier:   types.StringValue(image.DcIdentifier),
			CreatedBy:      types.StringValue(image.CreatedBy),
		}

		state.Images = append(state.Images, imageState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
func (i *imageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	i.client = client
}
