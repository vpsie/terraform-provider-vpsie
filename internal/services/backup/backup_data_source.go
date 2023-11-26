package backup

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type backupDataSource struct {
	client *govpsie.Client
}

type backupDataSourceModel struct {
	Backups []backupsModel `tfsdk:"backups"`
	ID      types.String   `tfsdk:"id"`
}

type backupsModel struct {
	Identifier   types.String `tfsdk:"identifier"`
	CreatedOn    types.String `tfsdk:"created_on"`
	DcIdentifier types.String `tfsdk:"dc_identifier"`
	CreatedBy    types.String `tfsdk:"created_by"`
	HostName     types.String `tfsdk:"hostname"`
	Name         types.String `tfsdk:"name"`
	Note         types.String `tfsdk:"note"`
	BackupKey    types.String `tfsdk:"backup_key"`
	State        types.String `tfsdk:"state"`
	VMIdentifier types.String `tfsdk:"vm_identifier"`
	BoxID        types.Int64  `tfsdk:"box_id"`
	BackupSHA1   types.String `tfsdk:"backupsha1"`
	OSFullName   types.String `tfsdk:"os_full_name"`
	VMCategory   types.String `tfsdk:"vm_category"`
}

// NewBackupDataSource is a helper function to create the data source.
func NewBackupDataSource() datasource.DataSource {
	return &backupDataSource{}
}

// Metadata returns the data source type name.
func (b *backupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "backups"
}

// Schema defines the schema for the data source.
func (b *backupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"backups": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
							Computed: true,
						},
						"created_by": schema.StringAttribute{
							Computed: true,
						},
						"hostname": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"note": schema.StringAttribute{
							Computed: true,
						},
						"backup_key": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"vm_identifier": schema.StringAttribute{
							Computed: true,
						},
						"box_id": schema.Int64Attribute{
							Computed: true,
						},
						"backupsha1": schema.StringAttribute{
							Computed: true,
						},
						"os_full_name": schema.StringAttribute{
							Computed: true,
						},
						"vm_category": schema.StringAttribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (b *backupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state backupDataSourceModel

	backups, err := b.client.Backup.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Getting backups",
			"Could not get backups, unexpected error: "+err.Error(),
		)

		return
	}

	for _, backup := range backups {
		backupState := backupsModel{
			Identifier:   types.StringValue(backup.Identifier),
			CreatedOn:    types.StringValue(backup.CreatedOn),
			DcIdentifier: types.StringValue(backup.DcIdentifier),
			CreatedBy:    types.StringValue(backup.CreatedBy),
			HostName:     types.StringValue(backup.HostName),
			Name:         types.StringValue(backup.Name),
			Note:         types.StringValue(backup.Note),
			BackupKey:    types.StringValue(backup.BackupKey),
			State:        types.StringValue(backup.State),
			VMIdentifier: types.StringValue(backup.VMIdentifier),
			BoxID:        types.Int64Value(int64(backup.BoxID)),
			BackupSHA1:   types.StringValue(backup.BackupSHA1),
			OSFullName:   types.StringValue(backup.OSFullName),
			VMCategory:   types.StringValue(backup.VMCategory),
		}

		state.Backups = append(state.Backups, backupState)
	}
}
func (b *backupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

	b.client = client
}
