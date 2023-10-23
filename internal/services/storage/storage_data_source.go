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
	_ datasource.DataSource              = &storageDataSource{}
	_ datasource.DataSourceWithConfigure = &storageDataSource{}
)

// storageDataSource is the data source implementation.
type storageDataSource struct {
	client *govpsie.Client
}

type storageDataSourceModel struct {
	Storages []storagesModel `tfsdk:"storages"`
	ID       types.String    `tfsdk:"id"`
}

type storagesModel struct {
	ID             types.Int64  `tfsdk:"id"`
	Name           types.String `tfsdk:"name"`
	Description    types.String `tfsdk:"description"`
	UserID         types.Int64  `tfsdk:"user_id"`
	BoxID          types.Int64  `tfsdk:"box_id"`
	Identifier     types.String `tfsdk:"identifier"`
	UserTemplateID types.Int64  `tfsdk:"user_template_id"`
	StorageType    types.String `tfsdk:"storage_type"`
	DiskFormat     types.String `tfsdk:"disk_format"`
	IsAutomatic    types.Int64  `tfsdk:"is_automatic"`
	Size           types.Int64  `tfsdk:"size"`
	StorageID      types.Int64  `tfsdk:"storage_id"`
	DiskKey        types.String `tfsdk:"disk_key"`
	CreatedOn      types.String `tfsdk:"created_on"`
	VmIdentifier   types.String `tfsdk:"vm_identifier"`
	Hostname       types.String `tfsdk:"hostname"`
	OsIdentifier   types.String `tfsdk:"os_identifier"`
	State          types.String `tfsdk:"state"`
	DcIdentifier   types.String `tfsdk:"dc_identifier"`
	BusDevice      types.String `tfsdk:"bus_device"`
	BusNumber      types.Int64  `tfsdk:"bus_number"`
}

// NewStorageDataSource is a helper function to create the data source.
func NewStorageDataSource() datasource.DataSource {
	return &storageDataSource{}
}

// Metadata returns the data source type name.
func (s *storageDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storages"
}

// Schema defines the schema for the data source.
func (s *storageDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"storages": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.Int64Attribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"description": schema.StringAttribute{
							Computed: true,
						},
						"user_id": schema.Int64Attribute{
							Computed: true,
						},
						"box_id": schema.Int64Attribute{
							Computed: true,
						},
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"user_template_id": schema.Int64Attribute{
							Computed: true,
						},
						"storage_type": schema.StringAttribute{
							Computed: true,
						},
						"disk_format": schema.StringAttribute{
							Computed: true,
						},
						"is_automatic": schema.Int64Attribute{
							Computed: true,
						},
						"size": schema.Int64Attribute{
							Computed: true,
						},
						"storage_id": schema.Int64Attribute{
							Computed: true,
						},
						"disk_key": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"vm_identifier": schema.StringAttribute{
							Computed: true,
						},
						"hostname": schema.StringAttribute{
							Computed: true,
						},
						"os_identifier": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
							Computed: true,
						},
						"bus_device": schema.StringAttribute{
							Computed: true,
						},
						"bus_number": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *storageDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state storageDataSourceModel

	storages, err := s.client.Storage.ListAll(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read Vpsie Storages",
			err.Error(),
		)

		return
	}

	for _, storage := range storages {
		storageState := storagesModel{
			ID:             types.Int64Value(int64(storage.ID)),
			Name:           types.StringValue(storage.Name),
			Description:    types.StringValue(storage.Description),
			UserID:         types.Int64Value(int64(storage.UserID)),
			BoxID:          types.Int64Value(int64(storage.BoxID)),
			Identifier:     types.StringValue(storage.Identifier),
			UserTemplateID: types.Int64Value(int64(storage.UserTemplateID)),
			StorageType:    types.StringValue(storage.StorageType),
			DiskFormat:     types.StringValue(storage.DiskFormat),
			IsAutomatic:    types.Int64Value(int64(storage.IsAutomatic)),
			Size:           types.Int64Value(int64(storage.Size)),
			StorageID:      types.Int64Value(int64(storage.StorageID)),
			DiskKey:        types.StringValue(storage.DiskKey),
			CreatedOn:      types.StringValue(storage.CreatedOn),
			VmIdentifier:   types.StringValue(storage.VmIdentifier),
			Hostname:       types.StringValue(storage.Hostname),
			OsIdentifier:   types.StringValue(storage.OsIdentifier),
			State:          types.StringValue(storage.State),
			DcIdentifier:   types.StringValue(storage.DcIdentifier),
			BusDevice:      types.StringValue(storage.BusDevice),
			BusNumber:      types.Int64Value(int64(storage.BusNumber)),
		}

		state.Storages = append(state.Storages, storageState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (s *storageDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
