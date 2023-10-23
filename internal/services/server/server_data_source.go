package server

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/vpsie/govpsie"
)

type serverDataSource struct {
	client *govpsie.Client
}

type serverDataSourceModel struct {
	Servers []serversModel `tfsdk:"servers"`
	ID      types.String   `tfsdk:"id"`
}

type serversModel struct {
	ID                  types.Int64  `tfsdk:"id"`
	Identifier          types.String `tfsdk:"identifier"`
	UserID              types.Int64  `tfsdk:"user_id"`
	BoxSizeID           types.Int64  `tfsdk:"boxsize_id"`
	BoxImageID          types.Int64  `tfsdk:"boximage_id"`
	DataCenterID        types.Int64  `tfsdk:"datacenter_id"`
	NodeID              types.Int64  `tfsdk:"node_id"`
	BoxdIsCountID       types.Int64  `tfsdk:"boxdiscount_id"`
	Hostname            types.String `tfsdk:"hostname"`
	DefaultIP           types.String `tfsdk:"default_ip"`
	DefaultIPv6         types.String `tfsdk:"default_ipv6"`
	PrivateIP           types.String `tfsdk:"private_ip"`
	IsAutoBackup        types.Int64  `tfsdk:"is_autobackup"`
	BoxVirtualization   types.String `tfsdk:"box_virtualization_id"`
	Ram                 types.Int64  `tfsdk:"ram"`
	Cpu                 types.Int64  `tfsdk:"cpu"`
	Ssd                 types.Int64  `tfsdk:"ssd"`
	Traffic             types.Int64  `tfsdk:"traffic"`
	AddedIpAddresses    types.String `tfsdk:"added_ip_addresses"`
	InitialPassword     types.String `tfsdk:"initial_password"`
	Notes               types.String `tfsdk:"notes"`
	CreatedOn           types.String `tfsdk:"created_on"`
	LastUpdated         types.String `tfsdk:"last_updated"`
	DroppedOn           types.String `tfsdk:"dropped_on"`
	IsActive            types.Int64  `tfsdk:"is_active"`
	IsDeleted           types.Int64  `tfsdk:"is_deleted"`
	Power               types.Int64  `tfsdk:"power"`
	ProjectID           types.Int64  `tfsdk:"project_id"`
	IsCustom            types.Int64  `tfsdk:"is_custom"`
	NrAddedIps          types.Int64  `tfsdk:"nr_added_ips"`
	InPcs               types.Int64  `tfsdk:"in_pcs"`
	CustomPrice         types.Int64  `tfsdk:"custom_price"`
	PayableLicense      types.Int64  `tfsdk:"payable_license"`
	LastLicensePay      types.String `tfsdk:"last_license_pay"`
	ScriptID            types.String `tfsdk:"script_id"`
	SshKeyID            types.String `tfsdk:"sshkey_id"`
	IsLocked            types.Int64  `tfsdk:"is_locked"`
	IsWorkWithNew       types.Int64  `tfsdk:"is_work_with_new_version"`
	IsSuspended         types.Int64  `tfsdk:"is_suspended"`
	IsTerminated        types.Int64  `tfsdk:"is_terminated"`
	OldID               types.Int64  `tfsdk:"old_id"`
	CustomIsoID         types.Int64  `tfsdk:"custom_iso_id"`
	IsIsoImageBootAble  types.Int64  `tfsdk:"is_iso_image_bootable"`
	HasSsl              types.Int64  `tfsdk:"has_ssl"`
	LastActionDate      types.String `tfsdk:"last_action_date"`
	IsCreatedFromLegacy types.Int64  `tfsdk:"is_created_from_legacy"`
	IsSmtpAllowed       types.Int64  `tfsdk:"is_smtp_allowed"`
	WeeklyBackup        types.Int64  `tfsdk:"weekly_backup"`
	MonthlyBackup       types.Int64  `tfsdk:"monthly_backup"`
	LibIsoID            types.Int64  `tfsdk:"lib_iso_id"`
	DailySnapshot       types.Int64  `tfsdk:"daily_snapshot"`
	WeeklySnapshot      types.Int64  `tfsdk:"weekly_snapshot"`
	MonthlySnapshot     types.Int64  `tfsdk:"monthly_snap"`
	LastActionInMin     types.Int64  `tfsdk:"last_action_in_min"`
	FirstName           types.String `tfsdk:"firstname"`
	LastName            types.String `tfsdk:"lastname"`
	Username            types.String `tfsdk:"username"`
	State               types.String `tfsdk:"state"`
	IsFipAvailable      types.Int64  `tfsdk:"is_fip_available"`
	IsBucketAvailable   types.Int64  `tfsdk:"is_bucket_available"`
	DcIdentifier        types.String `tfsdk:"dc_identifier"`
	Category            types.String `tfsdk:"category"`
	FullName            types.String `tfsdk:"fullname"`
	VmDescription       types.String `tfsdk:"vm_description"`
	BoxesSuspended      types.Int64  `tfsdk:"boxes_suspended"`
	IsSataAvailable     types.Int64  `tfsdk:"is_sata_available"`
	IsSsdAvailable      types.Int64  `tfsdk:"is_ssd_available"`
	PublicIp            types.String `tfsdk:"public_ip"`
}

// NewServerDataSource is a helper function to create the data source.
func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

// Metadata returns the data source type name.
func (s *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_servers"
}

// Schema defines the schema for the data source.
func (s *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"servers": schema.ListNestedAttribute{
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
						"boxsize_id": schema.Int64Attribute{
							Computed: true,
						},
						"boximage_id": schema.Int64Attribute{
							Computed: true,
						},
						"datacenter_id": schema.Int64Attribute{
							Computed: true,
						},
						"node_id": schema.Int64Attribute{
							Computed: true,
						},
						"boxdiscount_id": schema.Int64Attribute{
							Computed: true,
							Optional: true,
						},
						"hostname": schema.StringAttribute{
							Computed: true,
						},
						"default_ip": schema.StringAttribute{
							Computed: true,
						},
						"default_ipv6": schema.StringAttribute{
							Computed: true,
						},
						"private_ip": schema.StringAttribute{
							Computed: true,
						},
						"is_autobackup": schema.Int64Attribute{
							Computed: true,
						},
						"box_virtualization_id": schema.StringAttribute{
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
						"added_ip_addresses": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
						"initial_password": schema.StringAttribute{
							Computed: true,
						},
						"notes": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
						"created_on": schema.StringAttribute{
							Computed: true,
						},
						"last_updated": schema.StringAttribute{
							Computed: true,
						},
						"dropped_on": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
						"is_active": schema.Int64Attribute{
							Computed: true,
						},
						"is_deleted": schema.Int64Attribute{
							Computed: true,
						},
						"power": schema.Int64Attribute{
							Computed: true,
						},
						"project_id": schema.Int64Attribute{
							Computed: true,
						},
						"is_custom": schema.Int64Attribute{
							Computed: true,
						},
						"nr_added_ips": schema.Int64Attribute{
							Computed: true,
						},
						"in_pcs": schema.Int64Attribute{
							Computed: true,
						},
						"custom_price": schema.Int64Attribute{
							Computed: true,
							Optional: true,
						},
						"payable_license": schema.Int64Attribute{
							Computed: true,
						},
						"last_license_pay": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
						"script_id": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
						"sshkey_id": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
						"is_locked": schema.Int64Attribute{
							Computed: true,
						},
						"is_work_with_new_version": schema.Int64Attribute{
							Computed: true,
						},
						"is_suspended": schema.Int64Attribute{
							Computed: true,
						},
						"is_terminated": schema.Int64Attribute{
							Computed: true,
						},
						"old_id": schema.Int64Attribute{
							Computed: true,
						},
						"custom_iso_id": schema.Int64Attribute{
							Computed: true,
							Optional: true,
						},
						"is_iso_image_bootable": schema.Int64Attribute{
							Computed: true,
						},
						"has_ssl": schema.Int64Attribute{
							Computed: true,
						},
						"last_action_date": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
						"is_created_from_legacy": schema.Int64Attribute{
							Computed: true,
						},
						"is_smtp_allowed": schema.Int64Attribute{
							Computed: true,
						},
						"weekly_backup": schema.Int64Attribute{
							Computed: true,
						},
						"monthly_backup": schema.Int64Attribute{
							Computed: true,
						},
						"lib_iso_id": schema.Int64Attribute{
							Computed: true,
							Optional: true,
						},
						"daily_snapshot": schema.Int64Attribute{
							Computed: true,
						},
						"weekly_snapshot": schema.Int64Attribute{
							Computed: true,
						},
						"monthly_snap": schema.Int64Attribute{
							Computed: true,
						},
						"last_action_in_min": schema.Int64Attribute{
							Computed: true,
						},
						"firstname": schema.StringAttribute{
							Computed: true,
						},
						"lastname": schema.StringAttribute{
							Computed: true,
						},
						"username": schema.StringAttribute{
							Computed: true,
						},
						"state": schema.StringAttribute{
							Computed: true,
						},
						"is_fip_available": schema.Int64Attribute{
							Computed: true,
						},
						"is_bucket_available": schema.Int64Attribute{
							Computed: true,
						},
						"dc_identifier": schema.StringAttribute{
							Computed: true,
						},
						"category": schema.StringAttribute{
							Computed: true,
						},
						"fullname": schema.StringAttribute{
							Computed: true,
						},
						"vm_description": schema.StringAttribute{
							Computed: true,
						},
						"boxes_suspended": schema.Int64Attribute{
							Computed: true,
						},
						"is_sata_available": schema.Int64Attribute{
							Computed: true,
						},
						"is_ssd_available": schema.Int64Attribute{
							Computed: true,
						},
						"public_ip": schema.StringAttribute{
							Computed: true,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *serverDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serverDataSourceModel

	servers, err := s.client.Server.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading servers",
			"Could not read servers: "+err.Error(),
		)

		return
	}

	for _, server := range servers {
		var boxdIsCountId *int64
		var customPrice *int64
		var scriptID *string
		var sshKeyID *string
		var customIsoID *int64
		var libIsoID *int64

		if server.BoxdIsCountID != nil {
			value := int64(*server.BoxdIsCountID)
			boxdIsCountId = &value
		}

		if server.CustomPrice != nil {
			value := int64(*server.CustomPrice)
			customPrice = &value
		}

		if server.ScriptID != nil {
			value := string(*server.ScriptID)
			scriptID = &value
		}

		if server.SshKeyID != nil {
			value := string(*server.SshKeyID)
			sshKeyID = &value
		}

		if server.CustomIsoID != nil {
			value := int64(*server.CustomIsoID)
			customIsoID = &value
		}

		if server.LibIsoID != nil {
			value := int64(*server.LibIsoID)
			libIsoID = &value
		}

		serverState := serversModel{
			ID:                  types.Int64Value(int64(server.ID)),
			Identifier:          types.StringValue(server.Identifier),
			UserID:              types.Int64Value(int64(server.UserID)),
			BoxSizeID:           types.Int64Value(int64(server.BoxSizeID)),
			BoxImageID:          types.Int64Value(int64(server.BoxImageID)),
			DataCenterID:        types.Int64Value(int64(server.DataCenterID)),
			NodeID:              types.Int64Value(int64(server.NodeID)),
			BoxdIsCountID:       types.Int64PointerValue(boxdIsCountId),
			Hostname:            types.StringValue(server.Hostname),
			DefaultIP:           types.StringValue(server.DefaultIP),
			DefaultIPv6:         types.StringValue(server.DefaultIPv6),
			PrivateIP:           types.StringValue(server.PrivateIP),
			IsAutoBackup:        types.Int64Value(int64(server.IsAutoBackup)),
			BoxVirtualization:   types.StringValue(server.BoxVirtualization),
			Ram:                 types.Int64Value(int64(server.Ram)),
			Cpu:                 types.Int64Value(int64(server.Cpu)),
			Ssd:                 types.Int64Value(int64(server.Ssd)),
			Traffic:             types.Int64Value(int64(server.Traffic)),
			AddedIpAddresses:    types.StringPointerValue(server.AddedIpAddresses),
			InitialPassword:     types.StringValue(server.InitialPassword),
			Notes:               types.StringPointerValue(server.Notes),
			CreatedOn:           types.StringValue(server.CreatedOn),
			LastUpdated:         types.StringValue(server.LastUpdated),
			DroppedOn:           types.StringPointerValue(server.DroppedOn),
			IsActive:            types.Int64Value(int64(server.IsActive)),
			IsDeleted:           types.Int64Value(int64(server.IsDeleted)),
			Power:               types.Int64Value(int64(server.Power)),
			ProjectID:           types.Int64Value(int64(server.ProjectID)),
			IsCustom:            types.Int64Value(int64(server.IsCustom)),
			NrAddedIps:          types.Int64Value(int64(server.NrAddedIps)),
			InPcs:               types.Int64Value(int64(server.InPcs)),
			CustomPrice:         types.Int64PointerValue(customPrice),
			PayableLicense:      types.Int64Value(int64(server.PayableLicense)),
			LastLicensePay:      types.StringPointerValue(server.LastLicensePay),
			ScriptID:            types.StringPointerValue(scriptID),
			SshKeyID:            types.StringPointerValue(sshKeyID),
			IsLocked:            types.Int64Value(int64(server.IsLocked)),
			IsWorkWithNew:       types.Int64Value(int64(server.IsWorkWithNew)),
			IsSuspended:         types.Int64Value(int64(server.IsSuspended)),
			IsTerminated:        types.Int64Value(int64(server.IsTerminated)),
			OldID:               types.Int64Value(int64(server.OldID)),
			CustomIsoID:         types.Int64PointerValue(customIsoID),
			IsIsoImageBootAble:  types.Int64Value(int64(server.IsIsoImageBootAble)),
			HasSsl:              types.Int64Value(int64(server.HasSsl)),
			LastActionDate:      types.StringPointerValue(server.LastActionDate),
			IsCreatedFromLegacy: types.Int64Value(int64(server.IsCreatedFromLegacy)),
			IsSmtpAllowed:       types.Int64Value(int64(server.IsSmtpAllowed)),
			WeeklyBackup:        types.Int64Value(int64(server.WeeklyBackup)),
			MonthlyBackup:       types.Int64Value(int64(server.MonthlyBackup)),
			LibIsoID:            types.Int64PointerValue(libIsoID),
			DailySnapshot:       types.Int64Value(int64(server.DailySnapshot)),
			WeeklySnapshot:      types.Int64Value(int64(server.WeeklySnapshot)),
			MonthlySnapshot:     types.Int64Value(int64(server.MonthlySnapshot)),
			LastActionInMin:     types.Int64Value(int64(server.LastActionInMin)),
			FirstName:           types.StringValue(server.FirstName),
			LastName:            types.StringValue(server.LastName),
			Username:            types.StringValue(server.Username),
			State:               types.StringValue(server.State),
			IsFipAvailable:      types.Int64Value(int64(server.IsFipAvailable)),
			IsBucketAvailable:   types.Int64Value(int64(server.IsBucketAvailable)),
			DcIdentifier:        types.StringValue(server.DcIdentifier),
			Category:            types.StringValue(server.Category),
			FullName:            types.StringValue(server.FullName),
			VmDescription:       types.StringValue(server.VmDescription),
			BoxesSuspended:      types.Int64Value(int64(server.BoxesSuspended)),
			IsSataAvailable:     types.Int64Value(int64(server.IsSataAvailable)),
			IsSsdAvailable:      types.Int64Value(int64(server.IsSsdAvailable)),
			PublicIp:            types.StringPointerValue(server.PublicIp),
		}

		state.Servers = append(state.Servers, serverState)
	}

	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

func (s *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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
