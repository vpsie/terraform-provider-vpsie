package storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

var (
	_ datasource.DataSource              = &storageSnapshotDataSource{}
	_ datasource.DataSourceWithConfigure = &storageSnapshotDataSource{}
)

// storageSnapshotDataSource is the data source implementation.
type storageSnapshotDataSource struct {
	client *govpsie.Client
}

type storageSnapshotDataSourceModel struct {
	StorageSnapshots []storageSnapshotsModel `tfsdk:"storage_snapshots"`
	ID               types.String            `tfsdk:"id"`
}

type storageSnapshotsModel struct {
	ID          types.Int64  `tfsdk:"id"`
	Name        types.String `tfsdk:"name"`
	StorageID   types.Int64  `tfsdk:"storage_id"`
	Identifier  types.String `tfsdk:"identifier"`
	Size        types.Int64  `tfsdk:"size"`
	CreatedOn   types.String `tfsdk:"created_on"`
	UserID      types.Int64  `tfsdk:"user_id"`
	IsDeleted   types.Int64  `tfsdk:"is_deleted"`
	SnapshotKey types.String `tfsdk:"snapshot_key"`
	StorageName types.String `tfsdk:"storage_name"`
	StorageType types.String `tfsdk:"storage_type"`
	DiskFormat  types.String `tfsdk:"disk_format"`
	BoxID       types.Int64  `tfsdk:"box_id"`
	EntityType  types.String `tfsdk:"entity_type"`
}

// NewstorageSnapshotDataSource is a helper function to create the data source.
func NewStorageSnapshotDataSource() datasource.DataSource {
	return &storageSnapshotDataSource{}
}

// Metadata returns the data source type name.
func (s *storageSnapshotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_snapshots"
}

// Schema defines the schema for the data source.
func (s *storageSnapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"storage_snapshots": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"storage_id": schema.Int64Attribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"size": schema.Int64Attribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"is_deleted": schema.Int64Attribute{
							Computed: true,
						},
						"snapshot_key": schema.StringAttribute{
							Computed: true,
						},
						"storage_name": schema.StringAttribute{
							Computed: true,
						},
						"storage_type": schema.StringAttribute{
							Computed: true,
						},
						"disk_format": schema.StringAttribute{
							Computed: true,
						},
						"box_id": schema.Int64Attribute{
							Computed: true,
						},
						"entity_type": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *storageSnapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state storageSnapshotDataSourceModel

	storageSnapshots, err := s.client.Storage.ListSnapshots(ctx)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting Storage Snapshots",
			"Could not get storage snapshots for storage: "+err.Error(),
		)

		return
	}

	for _, storageSnap := range storageSnapshots {
		storageSnapState := storageSnapshotsModel{
			ID:          types.Int64Value(int64(storageSnap.ID)),
			Name:        types.StringValue(storageSnap.Name),
			StorageID:   types.Int64Value(int64(storageSnap.StorageID)),
			Identifier:  types.StringValue(storageSnap.Identifier),
			Size:        types.Int64Value(int64(storageSnap.Size)),
			CreatedOn:   types.StringValue(storageSnap.CreatedOn.String()),
			UserID:      types.Int64Value(int64(storageSnap.UserID)),
			IsDeleted:   types.Int64Value(int64(storageSnap.IsDeleted)),
			SnapshotKey: types.StringValue(storageSnap.SnapshotKey),
			StorageName: types.StringValue(storageSnap.StorageName),
			StorageType: types.StringValue(storageSnap.StorageType),
			DiskFormat:  types.StringValue(storageSnap.DiskFormat),
			BoxID:       types.Int64Value(int64(storageSnap.BoxID)),
			EntityType:  types.StringValue(storageSnap.EntityType),
		}

		state.StorageSnapshots = append(state.StorageSnapshots, storageSnapState)

	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *storageSnapshotDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
