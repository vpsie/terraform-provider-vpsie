package snapshot

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &serverSnapshotResource{}
	_ resource.ResourceWithConfigure   = &serverSnapshotResource{}
	_ resource.ResourceWithImportState = &serverSnapshotResource{}
)

type serverSnapshotResource struct {
	client *govpsie.Client
}

type serverSnapshotResourceModel struct {
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

func NewServerSnapshotResource() resource.Resource {
	return &serverSnapshotResource{}
}

func (s *serverSnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server_snapshot"
}

func (s *serverSnapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"hostname": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backup_key": schema.StringAttribute{
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
			"dc_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"box_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"daily": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"is_snapshot": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"vm_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"backupsha1": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_deleted_vm": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"created_on": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"note": schema.StringAttribute{
				Required: true,
			},
			"backup_size": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"dc_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"weekly": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"monthly": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"global_backup": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"os_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"os_full_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vm_category": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"vm_ssd": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (s *serverSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (s *serverSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan serverSnapshotResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	note := ""
	if !plan.Note.IsNull() {
		note = plan.Note.ValueString()
	}

	err := s.client.Snapshot.Create(ctx, plan.Name.ValueString(), plan.VmIdentifier.ValueString(), note)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating vpsie server snapshots",
			"couldn't create vpsie server snapshots "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	snapshot, err := s.GetSnapshotByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating vpsie server snapshots",
			"couldn't create vpsie server snapshots "+plan.Name.ValueString()+": "+err.Error(),
		)
		return
	}

	plan.Hostname = types.StringValue(snapshot.Hostname)
	plan.Name = types.StringValue(snapshot.Name)
	plan.BackupKey = types.StringValue(snapshot.BackupKey)
	plan.State = types.StringValue(snapshot.State)
	plan.DcIdentifier = types.StringValue(snapshot.DcIdentifier)
	plan.Daily = types.Int64Value(snapshot.Daily)
	plan.IsSnapshot = types.Int64Value(snapshot.IsSnapshot)
	plan.VmIdentifier = types.StringValue(snapshot.VmIdentifier)
	plan.BackupSHA1 = types.StringValue(snapshot.BackupSHA1)
	plan.IsDeletedVM = types.Int64Value(snapshot.IsDeletedVM)
	plan.CreatedOn = types.StringValue(snapshot.CreatedOn.String())
	plan.Note = types.StringValue(snapshot.Note)
	plan.BackupSize = types.Int64Value(snapshot.BackupSize)
	plan.DcName = types.StringValue(snapshot.DcName)
	plan.Weekly = types.Int64Value(snapshot.Weekly)
	plan.Monthly = types.Int64Value(snapshot.Monthly)
	plan.BoxID = types.Int64Value(snapshot.BoxID)
	plan.GlobalBackup = types.Int64Value(snapshot.GlobalBackup)
	plan.OsIdentifier = types.StringValue(snapshot.OsIdentifier)
	plan.OsFullName = types.StringValue(snapshot.OsFullName)
	plan.VMCategory = types.StringValue(snapshot.VMCategory)
	plan.VMSSD = types.Int64Value(snapshot.VMSSD)
	plan.Identifier = types.StringValue(snapshot.Identifier)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *serverSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state serverSnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshot, err := s.client.Snapshot.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading vpsie server snapshots",
			"couldn't read vpsie server snapshots identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state

	state.Hostname = types.StringValue(snapshot.Hostname)
	state.Name = types.StringValue(snapshot.Name)
	state.BackupKey = types.StringValue(snapshot.BackupKey)
	state.State = types.StringValue(snapshot.State)
	state.DcIdentifier = types.StringValue(snapshot.DcIdentifier)
	state.Daily = types.Int64Value(snapshot.Daily)
	state.IsSnapshot = types.Int64Value(snapshot.IsSnapshot)
	state.VmIdentifier = types.StringValue(snapshot.VmIdentifier)
	state.BackupSHA1 = types.StringValue(snapshot.BackupSHA1)
	state.IsDeletedVM = types.Int64Value(snapshot.IsDeletedVM)
	state.CreatedOn = types.StringValue(snapshot.CreatedOn.String())
	state.Note = types.StringValue(snapshot.Note)
	state.BackupSize = types.Int64Value(snapshot.BackupSize)
	state.DcName = types.StringValue(snapshot.DcName)
	state.Weekly = types.Int64Value(snapshot.Weekly)
	state.Monthly = types.Int64Value(snapshot.Monthly)
	state.BoxID = types.Int64Value(snapshot.BoxID)
	state.GlobalBackup = types.Int64Value(snapshot.GlobalBackup)
	state.OsIdentifier = types.StringValue(snapshot.OsIdentifier)
	state.OsFullName = types.StringValue(snapshot.OsFullName)
	state.VMCategory = types.StringValue(snapshot.VMCategory)
	state.VMSSD = types.Int64Value(snapshot.VMSSD)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (s *serverSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan serverSnapshotResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Snapshot.Update(ctx, plan.Identifier.ValueString(), plan.Note.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error updating vpsie server snapshots",
			"couldn't update vpsie server snapshots "+plan.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (s *serverSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state serverSnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Snapshot.Delete(ctx, state.Identifier.ValueString(), "no reason", "")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting server snapshot",
			"couldn't delete server snapshot, unexpected error: "+err.Error(),
		)

		return
	}
}

func (s *serverSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (s *serverSnapshotResource) GetSnapshotByName(ctx context.Context, snapshotName string) (*govpsie.Snapshot, error) {
	snapshots, err := s.client.Snapshot.List(ctx, &govpsie.ListOptions{Page: 0, PerPage: 10000})
	if err != nil {
		return nil, err
	}

	for _, snapshot := range snapshots {
		if snapshotName == snapshot.Name {
			return &snapshot, nil
		}
	}

	return nil, fmt.Errorf("snapshot with name %s not found", snapshotName)
}
