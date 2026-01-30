package storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &storageAttachmentResource{}
	_ resource.ResourceWithConfigure   = &storageAttachmentResource{}
	_ resource.ResourceWithImportState = &storageAttachmentResource{}
)

type storageAttachmentResource struct {
	client *govpsie.Client
}

type storageAttachmentResourceModel struct {
	VmIdentifier      types.String `tfsdk:"vm_identifier"`
	StorageIdentifier types.String `tfsdk:"storage_identifier"`
	VmType            types.String `tfsdk:"vm_type"`
}

func NewStorageAttachmentResource() resource.Resource {
	return &storageAttachmentResource{}
}

func (s *storageAttachmentResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_storage_attachement"
}

func (s *storageAttachmentResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"vm_identifier": schema.StringAttribute{
				Required: true,
			},
			"storage_identifier": schema.StringAttribute{
				Required: true,
			},
			"vm_type": schema.StringAttribute{
				Default:  stringdefault.StaticString("vm"),
				Optional: true,
				Computed: true,
			},
		},
	}
}

func (s *storageAttachmentResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
func (s *storageAttachmentResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan storageAttachmentResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Storage.AttachToServer(ctx, plan.StorageIdentifier.ValueString(), plan.VmIdentifier.ValueString(), plan.VmType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error attaching storage", err.Error())
		return
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (s *storageAttachmentResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state storageAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	storage, err := s.GetStorageSnapshotByIdentifier(ctx, state.StorageIdentifier.ValueString())
	if err != nil {
		if err.Error() == "snapshot not found" {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError("Error reading storage snapshot", err.Error())
		return
	}

	if storage.Identifier == "" || storage.Identifier != state.StorageIdentifier.ValueString() {
		tflog.Debug(ctx, "storage attachement %s was not found removing from state")
		resp.State.RemoveResource(ctx)
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (s *storageAttachmentResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state storageAttachmentResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := s.client.Storage.DetachToServer(ctx, state.StorageIdentifier.ValueString(), state.VmIdentifier.ValueString(), state.VmType.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error detaching storage", err.Error())
		return
	}
}

func (s *storageAttachmentResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("storage_identifier"), req, resp)
}

func (s *storageAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
}

func (s *storageAttachmentResource) GetStorageSnapshotByIdentifier(ctx context.Context, identifier string) (govpsie.StorageSnapShot, error) {
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
