package server

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &serverResource{}
	_ resource.ResourceWithConfigure   = &serverResource{}
	_ resource.ResourceWithImportState = &serverResource{}
)

type serverResource struct {
	client *govpsie.Client
}

type serverResourceModel struct {
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

	ResourceIdentifier types.String `tfsdk:"resource_identifier"`
	OsIdentifier       types.String `tfsdk:"os_identifier"`

	BackupEnabled types.Int64    `tfsdk:"backup_enabled"`
	AddPublicIpV4 types.Int64    `tfsdk:"add_public_ip_v4"`
	AddPublicIpV6 types.Int64    `tfsdk:"add_public_ip_v6"`
	AddPrivateIp  types.Int64    `tfsdk:"add_private_ip"`
	Password      types.String   `tfsdk:"password"`
	DeleteReason  types.String   `tfsdk:"delete_reason"`
	DeleteNote    types.String   `tfsdk:"delete_note"`
	Timeouts      timeouts.Value `tfsdk:"timeouts"`
}

func NewServerResource() resource.Resource {
	return &serverResource{}
}

func (s *serverResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (s *serverResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backup_enabled": schema.Int64Attribute{
				Default:  int64default.StaticInt64(0),
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"add_public_ip_v4": schema.Int64Attribute{
				Default:  int64default.StaticInt64(0),
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"add_public_ip_v6": schema.Int64Attribute{
				Default:  int64default.StaticInt64(0),
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"add_private_ip": schema.Int64Attribute{
				Default:  int64default.StaticInt64(0),
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.Int64Attribute{
				Required: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"resource_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"os_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dc_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"notes": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"script_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"sshkey_id": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"boxsize_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"boximage_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"datacenter_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"node_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"boxdiscount_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"default_ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"default_ipv6": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"private_ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_autobackup": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"box_virtualization_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"ram": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"cpu": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"ssd": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"traffic": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"added_ip_addresses": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"initial_password": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_on": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dropped_on": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"is_active": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_deleted": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"power": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"is_custom": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"nr_added_ips": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"in_pcs": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"custom_price": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"payable_license": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"last_license_pay": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"is_locked": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_work_with_new_version": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_suspended": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_terminated": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"old_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"custom_iso_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},

				Optional: true,
			},
			"is_iso_image_bootable": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"has_ssl": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"last_action_date": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"is_created_from_legacy": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_smtp_allowed": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"weekly_backup": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monthly_backup": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"lib_iso_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},

				Optional: true,
			},
			"daily_snapshot": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"weekly_snapshot": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monthly_snap": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"last_action_in_min": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"firstname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"lastname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"username": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"state": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_fip_available": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_bucket_available": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},

			"category": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"fullname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vm_description": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"boxes_suspended": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_sata_available": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_ssd_available": schema.Int64Attribute{
				Computed: true,

				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"public_ip": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
				Sensitive: true,
			},
			"delete_reason": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"delete_note": schema.StringAttribute{
				Optional: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},

			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (s *serverResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (s *serverResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var createServerReq *govpsie.CreateServerRequest = &govpsie.CreateServerRequest{}
	createServerReq.Tags = []*string{}

	createServerReq.AddPrivateIp = plan.AddPrivateIp.ValueInt64Pointer()
	createServerReq.AddPublicIpV4 = plan.AddPublicIpV4.ValueInt64Pointer()
	createServerReq.AddPublicIpV6 = plan.AddPublicIpV6.ValueInt64Pointer()
	createServerReq.BackupEnabled = plan.BackupEnabled.ValueInt64Pointer()
	createServerReq.ResourceIdentifier = plan.ResourceIdentifier.ValueString()
	createServerReq.OsIdentifier = plan.OsIdentifier.ValueString()
	createServerReq.DcIdentifier = plan.DcIdentifier.ValueString()
	createServerReq.Hostname = plan.Hostname.ValueString()
	createServerReq.ProjectID = plan.ProjectID.ValueInt64()

	if !plan.SshKeyID.IsNull() {
		createServerReq.SshKeyIdentifier = plan.SshKeyID.ValueStringPointer()
	}

	if !plan.ScriptID.IsNull() {
		createServerReq.ScriptIdentifier = plan.ScriptID.ValueStringPointer()
	}

	if !plan.Notes.IsNull() || plan.Notes.ValueString() != "" {
		createServerReq.Notes = plan.Notes.ValueStringPointer()
	}

	err := s.client.Server.CreateServer(ctx, createServerReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating server", err.Error())
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, 20*time.Minute)

	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	for {
		// Check if the context has expired
		if ctx.Err() != nil {
			resp.Diagnostics.AddError("error waiting for resource to become ready", ctx.Err().Error())
			return
		}

		// Check resource status here using provider's API or SDK
		server, ready, err := s.checkResourceStatus(ctx, plan.Hostname.ValueString())
		if err != nil {
			//  return fmt.Errorf("Timeout waiting for resource to become ready")
			resp.Diagnostics.AddError("Error cheking status of resource", err.Error())
			return
		}

		if ready {

			plan.ID = types.Int64Value(server.ID)
			plan.Identifier = types.StringValue(server.Identifier)
			plan.UserID = types.Int64Value(server.UserID)
			plan.BoxSizeID = types.Int64Value(server.BoxSizeID)
			plan.BoxImageID = types.Int64Value(server.BoxImageID)
			plan.DataCenterID = types.Int64Value(server.DataCenterID)
			plan.NodeID = types.Int64Value(server.NodeID)
			plan.BoxdIsCountID = types.Int64PointerValue(server.BoxdIsCountID)
			plan.Hostname = types.StringValue(server.Hostname)
			plan.DefaultIP = types.StringValue(server.DefaultIP)
			plan.DefaultIPv6 = types.StringValue(server.DefaultIPv6)
			plan.PrivateIP = types.StringValue(server.PrivateIP)
			plan.IsAutoBackup = types.Int64Value(server.IsAutoBackup)
			plan.BoxVirtualization = types.StringValue(server.BoxVirtualization)
			plan.Ram = types.Int64Value(server.Ram)
			plan.Cpu = types.Int64Value(server.Cpu)
			plan.Ssd = types.Int64Value(server.Ssd)
			plan.Traffic = types.Int64Value(server.Traffic)
			plan.AddedIpAddresses = types.StringPointerValue(server.AddedIpAddresses)
			plan.InitialPassword = types.StringValue(server.InitialPassword)
			plan.Notes = types.StringPointerValue(server.Notes)
			plan.CreatedOn = types.StringValue(server.CreatedOn)
			plan.LastUpdated = types.StringValue(server.LastUpdated)
			plan.DroppedOn = types.StringPointerValue(server.DroppedOn)
			plan.IsActive = types.Int64Value(server.IsActive)
			plan.IsDeleted = types.Int64Value(server.IsDeleted)
			plan.Power = types.Int64Value(server.Power)
			plan.ProjectID = types.Int64Value(server.ProjectID)
			plan.IsCustom = types.Int64Value(server.IsCustom)
			plan.NrAddedIps = types.Int64Value(server.NrAddedIps)
			plan.InPcs = types.Int64Value(server.InPcs)
			plan.CustomPrice = types.Int64PointerValue(server.CustomPrice)
			plan.PayableLicense = types.Int64Value(server.PayableLicense)
			plan.LastLicensePay = types.StringPointerValue(server.LastLicensePay)
			plan.ScriptID = types.StringPointerValue(server.ScriptID)
			plan.SshKeyID = types.StringPointerValue(server.SshKeyID)
			plan.IsLocked = types.Int64Value(server.IsLocked)
			plan.IsWorkWithNew = types.Int64Value(server.IsWorkWithNew)
			plan.IsSuspended = types.Int64Value(server.IsSuspended)
			plan.IsTerminated = types.Int64Value(server.IsTerminated)
			plan.OldID = types.Int64Value(server.OldID)
			plan.CustomIsoID = types.Int64PointerValue(server.CustomIsoID)
			plan.IsIsoImageBootAble = types.Int64Value(server.IsIsoImageBootAble)
			plan.HasSsl = types.Int64Value(server.HasSsl)
			plan.LastActionDate = types.StringPointerValue(server.LastActionDate)
			plan.IsCreatedFromLegacy = types.Int64Value(server.IsCreatedFromLegacy)
			plan.IsSmtpAllowed = types.Int64Value(server.IsSmtpAllowed)
			plan.WeeklyBackup = types.Int64Value(server.WeeklyBackup)
			plan.MonthlyBackup = types.Int64Value(server.MonthlyBackup)
			plan.LibIsoID = types.Int64PointerValue(server.LibIsoID)
			plan.DailySnapshot = types.Int64Value(server.DailySnapshot)
			plan.WeeklySnapshot = types.Int64Value(server.WeeklySnapshot)
			plan.MonthlySnapshot = types.Int64Value(server.MonthlySnapshot)
			plan.LastActionInMin = types.Int64Value(server.LastActionInMin)
			plan.FirstName = types.StringValue(server.FirstName)
			plan.LastName = types.StringValue(server.LastName)
			plan.Username = types.StringValue(server.Username)
			plan.State = types.StringValue(server.State)
			plan.IsFipAvailable = types.Int64Value(server.IsFipAvailable)
			plan.IsBucketAvailable = types.Int64Value(server.IsBucketAvailable)
			plan.DcIdentifier = types.StringValue(server.DcIdentifier)
			plan.Category = types.StringValue(server.Category)
			plan.FullName = types.StringValue(server.FullName)
			plan.VmDescription = types.StringValue(server.VmDescription)
			plan.BoxesSuspended = types.Int64Value(server.BoxesSuspended)
			plan.IsSataAvailable = types.Int64Value(server.IsSataAvailable)
			plan.IsSsdAvailable = types.Int64Value(server.IsSsdAvailable)
			plan.PublicIp = types.StringPointerValue(server.PublicIp)

			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			return
		}
		time.Sleep(5 * time.Second)
	}

}

// Read refreshes the Terraform state with the latest data.
func (s *serverResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	server, err := s.GetServerByIdentifier(ctx, state.Identifier.ValueString())
	if err != nil {

		resp.Diagnostics.AddError(
			"Error reading vpsie server",
			"Cloudn't read vpsie server identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.Int64Value(server.ID)
	state.Identifier = types.StringValue(server.Identifier)
	state.UserID = types.Int64Value(server.UserID)
	state.BoxSizeID = types.Int64Value(server.BoxSizeID)
	state.BoxImageID = types.Int64Value(server.BoxImageID)
	state.DataCenterID = types.Int64Value(server.DataCenterID)
	state.NodeID = types.Int64Value(server.NodeID)
	state.BoxdIsCountID = types.Int64PointerValue(server.BoxdIsCountID)
	state.Hostname = types.StringValue(server.Hostname)
	state.DefaultIP = types.StringValue(server.DefaultIP)
	state.DefaultIPv6 = types.StringValue(server.DefaultIPv6)
	state.PrivateIP = types.StringValue(server.PrivateIP)
	state.IsAutoBackup = types.Int64Value(server.IsAutoBackup)
	state.BoxVirtualization = types.StringValue(server.BoxVirtualization)
	state.Ram = types.Int64Value(server.Ram)
	state.Cpu = types.Int64Value(server.Cpu)
	state.Ssd = types.Int64Value(server.Ssd)
	state.Traffic = types.Int64Value(server.Traffic)
	state.AddedIpAddresses = types.StringPointerValue(server.AddedIpAddresses)
	state.InitialPassword = types.StringValue(server.InitialPassword)
	state.Notes = types.StringPointerValue(server.Notes)
	state.CreatedOn = types.StringValue(server.CreatedOn)
	state.LastUpdated = types.StringValue(server.LastUpdated)
	state.DroppedOn = types.StringPointerValue(server.DroppedOn)
	state.IsActive = types.Int64Value(server.IsActive)
	state.IsDeleted = types.Int64Value(server.IsDeleted)
	state.Power = types.Int64Value(server.Power)
	state.ProjectID = types.Int64Value(server.ProjectID)
	state.IsCustom = types.Int64Value(server.IsCustom)
	state.NrAddedIps = types.Int64Value(server.NrAddedIps)
	state.InPcs = types.Int64Value(server.InPcs)
	state.CustomPrice = types.Int64PointerValue(server.CustomPrice)
	state.PayableLicense = types.Int64Value(server.PayableLicense)
	state.LastLicensePay = types.StringPointerValue(server.LastLicensePay)
	state.ScriptID = types.StringPointerValue(server.ScriptID)
	state.SshKeyID = types.StringPointerValue(server.SshKeyID)
	state.IsLocked = types.Int64Value(server.IsLocked)
	state.IsWorkWithNew = types.Int64Value(server.IsWorkWithNew)
	state.IsSuspended = types.Int64Value(server.IsSuspended)
	state.IsTerminated = types.Int64Value(server.IsTerminated)
	state.OldID = types.Int64Value(server.OldID)
	state.CustomIsoID = types.Int64PointerValue(server.CustomIsoID)
	state.IsIsoImageBootAble = types.Int64Value(server.IsIsoImageBootAble)
	state.HasSsl = types.Int64Value(server.HasSsl)
	state.LastActionDate = types.StringPointerValue(server.LastActionDate)
	state.IsCreatedFromLegacy = types.Int64Value(server.IsCreatedFromLegacy)
	state.IsSmtpAllowed = types.Int64Value(server.IsSmtpAllowed)
	state.WeeklyBackup = types.Int64Value(server.WeeklyBackup)
	state.MonthlyBackup = types.Int64Value(server.MonthlyBackup)
	state.LibIsoID = types.Int64PointerValue(server.LibIsoID)
	state.DailySnapshot = types.Int64Value(server.DailySnapshot)
	state.WeeklySnapshot = types.Int64Value(server.WeeklySnapshot)
	state.MonthlySnapshot = types.Int64Value(server.MonthlySnapshot)
	state.LastActionInMin = types.Int64Value(server.LastActionInMin)
	state.FirstName = types.StringValue(server.FirstName)
	state.LastName = types.StringValue(server.LastName)
	state.Username = types.StringValue(server.Username)
	state.State = types.StringValue(server.State)
	state.IsFipAvailable = types.Int64Value(server.IsFipAvailable)
	state.IsBucketAvailable = types.Int64Value(server.IsBucketAvailable)
	state.DcIdentifier = types.StringValue(server.DcIdentifier)
	state.Category = types.StringValue(server.Category)
	state.FullName = types.StringValue(server.FullName)
	state.VmDescription = types.StringValue(server.VmDescription)
	state.BoxesSuspended = types.Int64Value(server.BoxesSuspended)
	state.IsSataAvailable = types.Int64Value(server.IsSataAvailable)
	state.IsSsdAvailable = types.Int64Value(server.IsSsdAvailable)
	state.PublicIp = types.StringPointerValue(server.PublicIp)
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (s *serverResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state serverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan serverResourceModel
	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.Hostname.Equal(plan.Hostname) {
		err := s.client.Server.ChangeHostName(ctx, state.Identifier.ValueString(), plan.Hostname.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating server hostname",
				"Cloudn't update server hostname, unexpected error: "+err.Error(),
			)

			return
		}

		state.Hostname = plan.Hostname
	}

	if !state.Power.Equal(plan.Power) {
		if plan.Power.ValueInt64() == 1 {
			err := s.client.Server.StartServer(ctx, state.Identifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error starting server",
					"Cloudn't start server, unexpected error: "+err.Error(),
				)

				return
			}
		} else {
			err := s.client.Server.StopServer(ctx, state.Identifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error stopping server",
					"Cloudn't stop server, unexpected error: "+err.Error(),
				)

				return
			}
		}
		state.Power = plan.Power
	}

	if !state.IsLocked.Equal(plan.IsLocked) {
		if plan.IsLocked.ValueInt64() == 1 {
			err := s.client.Server.Lock(ctx, state.Identifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error locking server",
					"Cloudn't lock server, unexpected error: "+err.Error(),
				)

				return
			}
		} else {
			err := s.client.Server.UnLock(ctx, state.Identifier.ValueString())
			if err != nil {
				resp.Diagnostics.AddError(
					"Error unlocking server",
					"Cloudn't unlock server, unexpected error: "+err.Error(),
				)

				return
			}
		}

		state.IsLocked = plan.IsLocked
	}

	if !state.SshKeyID.Equal(plan.SshKeyID) {
		err := s.client.Server.AddSsh(ctx, state.Identifier.ValueString(), plan.SshKeyID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating server sshkey",
				"Cloudn't update server sshkey, unexpected error: "+err.Error(),
			)

			return
		}

		state.SshKeyID = plan.SshKeyID
	}

	if !state.ScriptID.Equal(plan.ScriptID) {
		err := s.client.Server.AddScript(ctx, state.Identifier.ValueString(), plan.ScriptID.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating server script",
				"Cloudn't update server script, unexpected error: "+err.Error(),
			)

			return
		}

		state.ScriptID = plan.ScriptID
	}
	// TODO add  ResizeServer -> cpu, ram, ssd, bandwidth

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Delete deletes the resource and removes the Terraform state on success.
func (s *serverResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if state.DeleteReason.IsNull() || state.Password.IsNull() || state.DeleteReason.ValueString() == "" || state.Password.ValueString() == "" {
		resp.Diagnostics.AddError(
			"Error deleting server",
			"Delete reason and password are required to delete server",
		)

		return
	}

	var deleteNote string
	if state.DeleteNote.IsNull() {
		deleteNote = ""
	} else {
		deleteNote = state.DeleteNote.ValueString()
	}
	err := s.client.Server.DeleteServer(ctx, state.Identifier.ValueString(), state.Password.ValueString(), state.DeleteReason.ValueString(), deleteNote)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting server",
			"Cloudn't delete server, unexpected error: "+err.Error(),
		)

		return
	}
}

func (s *serverResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (s *serverResource) checkResourceStatus(ctx context.Context, hostname string) (*govpsie.VmData, bool, error) {
	servers, err := s.client.Server.List(ctx, nil)
	if err != nil {
		return nil, false, err
	}

	for _, server := range servers {
		if server.Hostname == hostname {
			return &server, true, nil
		}
	}

	return nil, false, nil
}

func (s *serverResource) GetServerByIdentifier(ctx context.Context, identifier string) (*govpsie.VmData, error) {
	servers, err := s.client.Server.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, server := range servers {
		if server.Identifier == identifier {
			return &server, nil
		}
	}

	return nil, nil
}
