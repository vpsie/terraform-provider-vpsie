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
	_ resource.Resource                = &storageResource{}
	_ resource.ResourceWithConfigure   = &storageResource{}
	_ resource.ResourceWithImportState = &storageResource{}
)

type storageResource struct {
	client *govpsie.Client
}

type storageResourceModel struct {
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

func NewStorageResource() resource.Resource {
	return &storageResource{}
}

func (s *storageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage"
}

func (s *storageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
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
			"dc_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"description": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				Required: true,
			},
			"storage_type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"disk_format": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"is_automatic": schema.Int64Attribute{
				Optional: true,
				Computed: true,
			},
			"user_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"box_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
				Optional: true,
			},
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"user_template_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"storage_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"disk_key": schema.StringAttribute{
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
			"vm_identifier": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"hostname": schema.StringAttribute{
				Computed: true,
				Optional: true,
			},
			"os_identifier": schema.StringAttribute{
				Computed: true,
			},
			"state": schema.StringAttribute{
				Computed: true,
			},
			"bus_device": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bus_number": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (s *storageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (s *storageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var storageReq *govpsie.StorageCreateRequest = &govpsie.StorageCreateRequest{}

	storageReq.Name = plan.Name.ValueString()
	storageReq.Size = int(plan.Size.ValueInt64())
	storageReq.Description = plan.Description.ValueString()
	storageReq.DcIdentifier = plan.DcIdentifier.ValueString()
	storageReq.DiskFormat = plan.DiskFormat.ValueString()
	storageReq.StorageType = plan.StorageType.ValueString()

	err := s.client.Storage.CreateVolume(ctx, storageReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating storage", err.Error())
		return
	}

	storage, err := s.GetVolumeByName(ctx, plan.Name.ValueString())

	if err != nil {
		resp.Diagnostics.AddError("Erorr fetching storage by name", err.Error())
		return
	}

	plan.BoxID = types.Int64Value(int64(storage.BoxID))
	plan.Identifier = types.StringValue(storage.Identifier)
	plan.DiskKey = types.StringValue(storage.DiskKey)
	plan.CreatedOn = types.StringValue(storage.CreatedOn)
	plan.VmIdentifier = types.StringValue(storage.VmIdentifier)
	plan.Hostname = types.StringValue(storage.Hostname)
	plan.OsIdentifier = types.StringValue(storage.OsIdentifier)
	plan.State = types.StringValue(storage.State)
	plan.BusDevice = types.StringValue(storage.BusDevice)
	plan.BusNumber = types.Int64Value(int64(storage.BusNumber))
	plan.ID = types.Int64Value(int64(storage.ID))
	plan.UserID = types.Int64Value(int64(storage.UserID))
	plan.UserTemplateID = types.Int64Value(int64(storage.UserTemplateID))
	plan.IsAutomatic = types.Int64Value(int64(storage.IsAutomatic))
	plan.StorageID = types.Int64Value(int64(storage.StorageID))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *storageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// storage, err := s.client.Storage.Get(ctx, state.Identifier.ValueString())
	storage, err := s.GetVolumeByIdentifier(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading vpsie storage",
			"couldn't read vpsie storage identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state

	state.BoxID = types.Int64Value(int64(storage.BoxID))
	state.DiskKey = types.StringValue(storage.DiskKey)
	state.CreatedOn = types.StringValue(storage.CreatedOn)
	state.VmIdentifier = types.StringValue(storage.VmIdentifier)
	state.Hostname = types.StringValue(storage.Hostname)
	state.OsIdentifier = types.StringValue(storage.OsIdentifier)
	state.State = types.StringValue(storage.State)
	state.BusDevice = types.StringValue(storage.BusDevice)
	state.BusNumber = types.Int64Value(int64(storage.BusNumber))
	state.UserID = types.Int64Value(int64(storage.UserID))
	state.UserTemplateID = types.Int64Value(int64(storage.UserTemplateID))
	state.IsAutomatic = types.Int64Value(int64(storage.IsAutomatic))
	state.StorageType = types.StringValue(storage.StorageType)
	state.DiskFormat = types.StringValue(storage.DiskFormat)
	state.Size = types.Int64Value(int64(storage.Size))
	state.Description = types.StringValue(storage.Description)
	state.DcIdentifier = types.StringValue(storage.DcIdentifier)
	state.Name = types.StringValue(storage.Name)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}

// Update updates the resource and sets the updated Terraform state on success.
func (s *storageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var nameState, sizeState, namePlan, sizePlan, identifier types.String

	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("name"), &nameState)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("size"), &sizeState)...)
	resp.Diagnostics.Append(req.State.GetAttribute(ctx, path.Root("identifier"), &identifier)...)

	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("name"), &namePlan)...)
	resp.Diagnostics.Append(req.Plan.GetAttribute(ctx, path.Root("size"), &sizePlan)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if !namePlan.Equal(nameState) {
		err := s.client.Storage.UpdateName(ctx, identifier.ValueString(), namePlan.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating storage name",
				"couldn't update storage name, unexpected error: "+err.Error(),
			)
		}

		resp.State.SetAttribute(ctx, path.Root("name"), namePlan)

	}

	if !sizePlan.Equal(sizeState) {
		err := s.client.Storage.UpdateSize(ctx, identifier.ValueString(), sizePlan.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating storage size",
				"couldn't update storage size, unexpected error: "+err.Error(),
			)
		}

		resp.State.SetAttribute(ctx, path.Root("size"), sizePlan)
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (s *storageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Storage.Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting storage",
			"couldn't delete storage, unexpected error: "+err.Error(),
		)

		return
	}
}

func (s *storageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (s *storageResource) GetVolumeByName(ctx context.Context, name string) (*govpsie.Storage, error) {
	volumes, err := s.client.Storage.ListAll(ctx, &govpsie.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		if volume.Name == name {
			return &volume, nil
		}
	}

	return nil, fmt.Errorf("volume with name %s not found", name)
}

func (s *storageResource) GetVolumeByIdentifier(ctx context.Context, identifier string) (*govpsie.Storage, error) {
	volumes, err := s.client.Storage.ListAll(ctx, &govpsie.ListOptions{})
	if err != nil {
		return nil, err
	}

	for _, volume := range volumes {
		if volume.Identifier == identifier {
			return &volume, nil
		}
	}

	return nil, fmt.Errorf("volume with identifier %s not found", identifier)
}
