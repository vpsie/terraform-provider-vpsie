package backup

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
	_ resource.Resource                = &backupResource{}
	_ resource.ResourceWithConfigure   = &backupResource{}
	_ resource.ResourceWithImportState = &backupResource{}
)

type backupResource struct {
	client *govpsie.Client
}

type backupResourceModel struct {
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

func NewBackupResource() resource.Resource {
	return &backupResource{}
}

func (s *backupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_backup"
}

func (s *backupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
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
			"dc_identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"created_by": schema.StringAttribute{
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
			},
			"note": schema.StringAttribute{
				Optional: true,
			},
			"backup_key": schema.StringAttribute{
				Required: true,
			},
			"state": schema.StringAttribute{
				Computed:      true,
				PlanModifiers: []planmodifier.String{},
			},
			"vm_identifier": schema.StringAttribute{
				Required: true,
			},
			"box_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"backupsha1": schema.StringAttribute{
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
		},
	}
}

func (b *backupResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create creates the resource and sets the initial Terraform state.
func (b *backupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan backupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := b.client.Backup.CreateBackups(ctx, plan.VMIdentifier.ValueString(), plan.Name.ValueString(), plan.Note.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating backup", err.Error())
		return
	}

	backup, err := b.GetBackupByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting backup", err.Error())
		return
	}

	// Overwrite items with refreshed state

	plan.Identifier = types.StringValue(backup.Identifier)
	plan.CreatedOn = types.StringValue(backup.CreatedOn)
	plan.DcIdentifier = types.StringValue(backup.DcIdentifier)
	plan.CreatedBy = types.StringValue(backup.CreatedBy)
	plan.HostName = types.StringValue(backup.HostName)
	plan.Name = types.StringValue(backup.Name)
	plan.Note = types.StringValue(backup.Note)
	plan.BackupKey = types.StringValue(backup.BackupKey)
	plan.State = types.StringValue(backup.State)
	plan.VMIdentifier = types.StringValue(backup.VMIdentifier)
	plan.BoxID = types.Int64Value(int64(backup.BoxID))
	plan.BackupSHA1 = types.StringValue(backup.BackupSHA1)
	plan.OSFullName = types.StringValue(backup.OSFullName)
	plan.VMCategory = types.StringValue(backup.VMCategory)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *backupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state backupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	backup, err := s.client.Backup.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading vpsie backups",
			"couldn't read vpsie backups identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state

	state.Name = types.StringValue(backup.Name)
	state.CreatedOn = types.StringValue(backup.CreatedOn)
	state.DcIdentifier = types.StringValue(backup.DcIdentifier)
	state.CreatedBy = types.StringValue(backup.CreatedBy)
	state.HostName = types.StringValue(backup.HostName)
	state.Note = types.StringValue(backup.Note)
	state.BackupKey = types.StringValue(backup.BackupKey)
	state.State = types.StringValue(backup.State)
	state.VMIdentifier = types.StringValue(backup.VMIdentifier)
	state.BoxID = types.Int64Value(int64(backup.BoxID))
	state.BackupSHA1 = types.StringValue(backup.BackupSHA1)
	state.OSFullName = types.StringValue(backup.OSFullName)
	state.VMCategory = types.StringValue(backup.VMCategory)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (s *backupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan backupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state backupResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Name.Equal(state.Name) {
		err := s.client.Backup.Rename(ctx, state.Identifier.ValueString(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Error renaming backup", err.Error())
			return
		}

		state.Name = plan.Name
		diags = resp.State.Set(ctx, &state)
		resp.Diagnostics.Append(diags...)
		if resp.Diagnostics.HasError() {
			return
		}
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (b *backupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state backupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := b.client.Backup.DeleteBackup(ctx, state.Identifier.ValueString(), "terraform delete", "terraform delete")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting backup",
			"couldn't delete backup, unexpected error: "+err.Error(),
		)

		return
	}
}

func (b *backupResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (b *backupResource) GetBackupByName(ctx context.Context, backupName string) (*govpsie.Backup, error) {
	backups, err := b.client.Backup.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, backup := range backups {
		if backupName == backup.Name {
			return &backup, nil
		}
	}

	return nil, fmt.Errorf("backup with name %s not found", backupName)
}
