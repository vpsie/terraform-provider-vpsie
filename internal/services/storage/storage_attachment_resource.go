package storage

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"

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
				MarkdownDescription: "Identifier of the server to attach the volume to. " +
					"Changing it detaches and re-attaches the volume, so it forces a new attachment.",
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"storage_identifier": schema.StringAttribute{
				MarkdownDescription: "Identifier of the storage volume. Changing it forces a new attachment.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vm_type": schema.StringAttribute{
				MarkdownDescription: "Type of target the volume attaches to. Changing it forces a new attachment.",
				Default:             stringdefault.StaticString("vm"),
				Optional:            true,
				Computed:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
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

	// Look the volume up in the STORAGE listing. This used to search the
	// storage-SNAPSHOT listing, where a volume identifier can never appear, so
	// the lookup always failed and every refresh silently dropped the
	// attachment from state -- after which the next apply tried to attach again
	// and was rejected with "This storage is currently attached to a Server".
	storage, err := s.getStorageByIdentifier(ctx, state.StorageIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading storage", err.Error())
		return
	}

	if storage == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// The volume still exists; make sure it is still attached to the server
	// this resource manages. The listing reports the attachment either as the
	// VM identifier or, on endpoints that omit it, as a non-zero box id.
	attached := storage.BoxID != 0
	if storage.VmIdentifier != "" {
		attached = storage.VmIdentifier == state.VmIdentifier.ValueString()
	}

	if !attached {
		resp.State.RemoveResource(ctx)
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
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

// Update is never called: every attribute is RequiresReplace, because moving a
// volume means detaching it and attaching it again. Leaving this body empty
// silently accepted the new values into state without touching the API.
func (s *storageAttachmentResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	resp.Diagnostics.AddError(
		"Storage attachment update is not supported",
		"Every attribute of vpsie_storage_attachement forces replacement. Reaching this point "+
			"means the schema and this method have drifted apart; please report it.",
	)
}

// getStorageByIdentifier returns the storage volume with the given identifier,
// or nil when it no longer exists.
func (s *storageAttachmentResource) getStorageByIdentifier(ctx context.Context, identifier string) (*govpsie.Storage, error) {
	storages, err := s.client.Storage.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for i := range storages {
		if storages[i].Identifier == identifier {
			return &storages[i], nil
		}
	}

	return nil, nil
}
