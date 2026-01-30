package storage

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
	_ resource.Resource                = &storageSnapshotResource{}
	_ resource.ResourceWithConfigure   = &storageSnapshotResource{}
	_ resource.ResourceWithImportState = &storageSnapshotResource{}
)

type storageSnapshotResource struct {
	client *govpsie.Client
}

type storageSnapshotResourceModel struct {
	ID                types.Int64  `tfsdk:"id"`
	StorageID         types.Int64  `tfsdk:"storage_id"`
	Identifier        types.String `tfsdk:"identifier"`
	Name              types.String `tfsdk:"name"`
	Size              types.Int64  `tfsdk:"size"`
	CreatedOn         types.String `tfsdk:"created_on"`
	UserID            types.Int64  `tfsdk:"user_id"`
	IsDeleted         types.Int64  `tfsdk:"is_deleted"`
	SnapshotKey       types.String `tfsdk:"snapshot_key"`
	StorageName       types.String `tfsdk:"storage_name"`
	StorageType       types.String `tfsdk:"storage_type"`
	DiskFormat        types.String `tfsdk:"disk_format"`
	BoxID             types.Int64  `tfsdk:"box_id"`
	EntityType        types.String `tfsdk:"entity_type"`
	StorageIdentifier types.String `tfsdk:"storage_identifier"`
	Type              types.String `tfsdk:"type"`
}

func NewStorageSnapshotResource() resource.Resource {
	return &storageSnapshotResource{}
}

func (s *storageSnapshotResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_snapshot"
}

func (s *storageSnapshotResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_id": schema.Int64Attribute{
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
			"size": schema.Int64Attribute{
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
			"user_id": schema.Int64Attribute{
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
			"snapshot_key": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"storage_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"disk_format": schema.StringAttribute{
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
			"entity_type": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (s *storageSnapshotResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (s *storageSnapshotResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageSnapshotResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Storage.CreateSnapshot(ctx, plan.StorageIdentifier.ValueString(), plan.Name.ValueString(), plan.Type.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating snapshot", err.Error())
		return
	}

	snapshot, err := s.GetStorageSnapshot(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error getting snapshot", err.Error())
		return
	}

	plan.ID = types.Int64Value(int64(snapshot.ID))
	plan.StorageID = types.Int64Value(int64(snapshot.StorageID))
	plan.Identifier = types.StringValue(snapshot.Identifier)
	plan.Size = types.Int64Value(int64(snapshot.Size))
	plan.CreatedOn = types.StringValue(snapshot.CreatedOn.String())
	plan.BoxID = types.Int64Value(int64(snapshot.BoxID))
	plan.EntityType = types.StringValue(snapshot.EntityType)
	plan.DiskFormat = types.StringValue(snapshot.DiskFormat)
	plan.IsDeleted = types.Int64Value(int64(snapshot.IsDeleted))
	plan.SnapshotKey = types.StringValue(snapshot.SnapshotKey)
	plan.StorageName = types.StringValue(snapshot.StorageName)
	plan.UserID = types.Int64Value(int64(snapshot.UserID))
	plan.StorageType = types.StringValue(snapshot.StorageType)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *storageSnapshotResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageSnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	snapshot, err := s.GetStorageSnapshotByIdentifier(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading storage snapshot",
			"couldn't read vpsie storage snapshot identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state

	state.ID = types.Int64Value(int64(snapshot.ID))
	state.StorageID = types.Int64Value(int64(snapshot.StorageID))
	state.Identifier = types.StringValue(snapshot.Identifier)
	state.Size = types.Int64Value(int64(snapshot.Size))
	state.CreatedOn = types.StringValue(snapshot.CreatedOn.String())
	state.BoxID = types.Int64Value(int64(snapshot.BoxID))
	state.EntityType = types.StringValue(snapshot.EntityType)
	state.DiskFormat = types.StringValue(snapshot.DiskFormat)
	state.IsDeleted = types.Int64Value(int64(snapshot.IsDeleted))
	state.SnapshotKey = types.StringValue(snapshot.SnapshotKey)
	state.StorageName = types.StringValue(snapshot.StorageName)
	state.UserID = types.Int64Value(int64(snapshot.UserID))
	state.StorageType = types.StringValue(snapshot.StorageType)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (s *storageSnapshotResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageSnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Storage.DeleteSnapshot(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting storage snapshot",
			"couldn't delete storage snapshot, unexpected error: "+err.Error(),
		)

		return
	}

}

func (s *storageSnapshotResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state storageSnapshotResourceModel
	var plan storageSnapshotResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	diags = req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !state.Name.Equal(plan.Name) {
		err := s.client.Storage.UpdateSnapshotName(ctx, plan.Identifier.String(), plan.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating storage name",
				"couldn't update storage name, unexpected error: "+err.Error(),
			)

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

func (s *storageSnapshotResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (s *storageSnapshotResource) GetStorageSnapshot(ctx context.Context, name string) (govpsie.StorageSnapShot, error) {
	snapshots, err := s.client.Storage.ListSnapshots(ctx, nil)
	if err != nil {
		return govpsie.StorageSnapShot{}, err
	}

	for _, snap := range snapshots {
		if snap.Name == name {
			return snap, nil
		}
	}

	return govpsie.StorageSnapShot{}, fmt.Errorf("snapshot not found")
}

func (s *storageSnapshotResource) GetStorageSnapshotByIdentifier(ctx context.Context, identifier string) (govpsie.StorageSnapShot, error) {
	snapshots, err := s.client.Storage.ListSnapshots(ctx, nil)
	if err != nil {
		return govpsie.StorageSnapShot{}, err
	}

	for _, snap := range snapshots {
		if snap.Identifier == identifier {
			return snap, nil
		}
	}

	return govpsie.StorageSnapShot{}, fmt.Errorf("snapshot not found")
}
