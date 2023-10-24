package snapshot

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
)

type serverSnapshotDataSource struct {
	client *govpsie.Client
}

type serverSnapshotDataSourceModel struct {
	ServerSnapshots []serverSnapshotsModel `tfsdk:"server_snapshots"`
	ID              types.String           `tfsdk:"id"`
}

type serverSnapshotsModel struct {
	Identifier   types.String `tfsdk:"identifier"`
	Hostname     types.String `tfsdk:"hostname"`
	Name         types.String `tfsdk:"name"`
	BackupKey    types.String `tfsdk:"backup_key"`
	State        types.String `tfsdk:"state"`
	DcIdentifier types.String `tfsdk:"dc_identifier"`
	Daily        types.Int64  `tfsdk:"daily"`
	IsSnapshot   types.Int64  `tfsdk:"is_snapshot"`
	VmIdentifier types.String `tfsdk:"vm_identifier"`
	BackupSHA1   types.String `tfsdk:"backupsha1"`
	IsDeletedVM  types.Int64  `tfsdk:"is_deleted_vm"`
	CreatedOn    types.String `tfsdk:"created_on"`
	Note         types.String `tfsdk:"note"`
	BackupSize   types.Int64  `tfsdk:"backup_size"`
	DcName       types.String `tfsdk:"dc_name"`
	Weekly       types.Int64  `tfsdk:"weekly"`
	Monthly      types.Int64  `tfsdk:"monthly"`
	BoxID        types.Int64  `tfsdk:"box_id"`
	GlobalBackup types.Int64  `tfsdk:"global_backup"`
	OsIdentifier types.String `tfsdk:"os_identifier"`
	OsFullName   types.String `tfsdk:"os_full_name"`
	VMCategory   types.String `tfsdk:"vm_category"`
	VMSSD        types.Int64  `tfsdk:"vm_ssd"`
}

// NewServerSnapshotDataSource is a helper function to create the data source.
func NewServerSnapshotDataSource() datasource.DataSource {
	return &serverSnapshotDataSource{}
}

// Metadata returns the data source type name.
func (s *serverSnapshotDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_snapshots"
}

// Schema defines the schema for the data source.
func (s *serverSnapshotDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"server_snapshots": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"identifier": schema.StringAttribute{
							Computed: true,
						},
						"hostname": schema.StringAttribute{
							Computed: true,
						},
						"name": schema.StringAttribute{
							Computed: true,
						},
						"backup_key": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
							Computed: true,
						},
						"box_id": schema.Int64Attribute{
							Computed: true,
						},
						"daily": schema.Int64Attribute{
							Computed: true,
						},
						"is_snapshot": schema.Int64Attribute{
							Computed: true,
						},
						"vm_identifier": schema.StringAttribute{
							Computed: true,
						},
						"backupsha1": schema.StringAttribute{
							Computed: true,
						},
						"is_deleted_vm": schema.Int64Attribute{
							Computed: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"note": schema.StringAttribute{
							Computed: true,
						},
						"backup_size": schema.Int64Attribute{
							Computed: true,
						},
						"dc_name": schema.StringAttribute{
							Computed: true,
						},
						"weekly": schema.Int64Attribute{
							Computed: true,
						},
						"monthly": schema.Int64Attribute{
							Computed: true,
						},
						"global_backup": schema.Int64Attribute{
							Computed: true,
						},
						"os_identifier": schema.StringAttribute{
							Computed: true,
						},
						"os_full_name": schema.StringAttribute{
							Computed: true,
						},
						"vm_category": schema.StringAttribute{
							Computed: true,
						},
						"vm_ssd": schema.Int64Attribute{
							Computed: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *serverSnapshotDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serverSnapshotDataSourceModel

	snapshots, err := s.client.Snapshot.List(ctx, &govpsie.ListOptions{Page: 1, PerPage: 1000})
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to get snapshots",
			"An unexpected error occurred when getting snapshots: "+err.Error(),
		)
		return
	}

	tflog.Info(ctx, "Got snapshots", map[string]interface{}{"snapshots": snapshots})

	for _, snapshot := range snapshots {

		snapshotState := serverSnapshotsModel{
			Identifier:   types.StringValue(snapshot.Identifier),
			Hostname:     types.StringValue(snapshot.Hostname),
			Name:         types.StringValue(snapshot.Name),
			BackupKey:    types.StringValue(snapshot.BackupKey),
			State:        types.StringValue(snapshot.State),
			DcIdentifier: types.StringValue(snapshot.DcIdentifier),
			Daily:        types.Int64Value(snapshot.Daily),
			IsSnapshot:   types.Int64Value(snapshot.IsSnapshot),
			VmIdentifier: types.StringValue(snapshot.VmIdentifier),
			BackupSHA1:   types.StringValue(snapshot.BackupSHA1),
			IsDeletedVM:  types.Int64Value(snapshot.IsDeletedVM),
			CreatedOn:    types.StringValue(snapshot.CreatedOn.String()),
			Note:         types.StringValue(snapshot.Note),
			BackupSize:   types.Int64Value(snapshot.BackupSize),
			DcName:       types.StringValue(snapshot.DcName),
			Weekly:       types.Int64Value(snapshot.Weekly),
			Monthly:      types.Int64Value(snapshot.Monthly),
			BoxID:        types.Int64Value(snapshot.BoxID),
			GlobalBackup: types.Int64Value(snapshot.GlobalBackup),
			OsIdentifier: types.StringValue(snapshot.OsIdentifier),
			OsFullName:   types.StringValue(snapshot.OsFullName),
			VMCategory:   types.StringValue(snapshot.VMCategory),
			VMSSD:        types.Int64Value(snapshot.VMSSD),
		}

		state.ServerSnapshots = append(state.ServerSnapshots, snapshotState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (s *serverSnapshotDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
