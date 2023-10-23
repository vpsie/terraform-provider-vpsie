package image

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
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
	_ resource.Resource                = &imageResource{}
	_ resource.ResourceWithConfigure   = &imageResource{}
	_ resource.ResourceWithImportState = &imageResource{}
)

type imageResource struct {
	client *govpsie.Client
}

type imageResourceModel struct {
	ID             types.Int64    `tfsdk:"id"`
	Identifier     types.String   `tfsdk:"identifier"`
	UserID         types.Int64    `tfsdk:"user_id"`
	DataCenterID   types.Int64    `tfsdk:"datacenter_id"`
	ImageSize      types.Int64    `tfsdk:"image_size"`
	OriginalName   types.String   `tfsdk:"original_name"`
	FetchedFromUrl types.String   `tfsdk:"fetched_from_url"`
	ImageHash      types.String   `tfsdk:"image_hash"`
	ImageLabel     types.String   `tfsdk:"image_label"`
	CreatedOn      types.String   `tfsdk:"created_on"`
	Deleted        types.Int64    `tfsdk:"deleted"`
	DcName         types.String   `tfsdk:"dc_name"`
	DcIdentifier   types.String   `tfsdk:"dc_identifier"`
	CreatedBy      types.String   `tfsdk:"created_by"`
	Timeouts       timeouts.Value `tfsdk:"timeouts"`
}

func NewImageResource() resource.Resource {
	return &imageResource{}
}

func (i *imageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_image"
}

func (i *imageResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"image_label": schema.StringAttribute{
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
			"fetched_from_url": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"user_id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"image_size": schema.Int64Attribute{
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
			"original_name": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"deleted": schema.Int64Attribute{
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
			"image_hash": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"dc_name": schema.StringAttribute{
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
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create: true,
			}),
		},
	}
}

func (i *imageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	i.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (i *imageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan imageResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := i.client.Image.CreateImages(ctx, plan.DcIdentifier.ValueString(), plan.ImageLabel.ValueString(), plan.FetchedFromUrl.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating image", err.Error())
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
		image, ready, err := i.checkResourceStatus(ctx, plan.ImageLabel.ValueString())
		if err != nil {
			//  return fmt.Errorf("Timeout waiting for resource to become ready")
			resp.Diagnostics.AddError("Error cheking status of resource", err.Error())
			return
		}

		if ready {
			plan.ID = types.Int64Value(int64(image.ID))
			plan.Identifier = types.StringValue(image.Identifier)
			plan.UserID = types.Int64Value(int64(image.UserID))
			plan.DataCenterID = types.Int64Value(int64(image.DatacenterID))
			plan.ImageSize = types.Int64Value(int64(image.ImageSize))
			plan.OriginalName = types.StringValue(image.OriginalName)
			plan.FetchedFromUrl = types.StringValue(image.FetchedFromURL)
			plan.ImageHash = types.StringValue(image.ImageHash)
			plan.ImageLabel = types.StringValue(image.ImageLabel)
			plan.CreatedOn = types.StringValue(image.CreatedOn.String())
			plan.Deleted = types.Int64Value(int64(image.Deleted))
			plan.DcName = types.StringValue(image.DcName)
			plan.DcIdentifier = types.StringValue(image.DcIdentifier)
			plan.CreatedBy = types.StringValue(image.CreatedBy)

			diags = resp.State.Set(ctx, plan)
			resp.Diagnostics.Append(diags...)
			if resp.Diagnostics.HasError() {
				return
			}

			return
		}

		// Wait for a delay before retrying
		time.Sleep(5 * time.Second)
	}

}

// Read refreshes the Terraform state with the latest data.
func (i *imageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	image, err := i.client.Image.GetImage(ctx, state.Identifier.ValueString())
	if err != nil {
		if err.Error() == "image not found" {
			resp.State.RemoveResource(ctx)
			return
		}

		resp.Diagnostics.AddError(
			"Error reading vpsie image",
			"Cloudn't read vpsie image identifier "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.Int64Value(int64(image.ID))
	state.Identifier = types.StringValue(image.Identifier)
	state.UserID = types.Int64Value(int64(image.UserID))
	state.DataCenterID = types.Int64Value(int64(image.DatacenterID))
	state.ImageSize = types.Int64Value(int64(image.ImageSize))
	state.OriginalName = types.StringValue(image.OriginalName)
	state.FetchedFromUrl = types.StringValue(image.FetchedFromURL)
	state.ImageHash = types.StringValue(image.ImageHash)
	state.ImageLabel = types.StringValue(image.ImageLabel)
	state.CreatedOn = types.StringValue(image.CreatedOn.String())
	state.Deleted = types.Int64Value(int64(image.Deleted))
	state.DcName = types.StringValue(image.DcName)
	state.DcIdentifier = types.StringValue(image.DcIdentifier)
	state.CreatedBy = types.StringValue(image.CreatedBy)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (i *imageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Delete deletes the resource and removes the Terraform state on success.
func (i *imageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state imageResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := i.client.Image.DeleteImage(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting image",
			"Cloudn't delete image, unexpected error: "+err.Error(),
		)

		return
	}
}

func (i *imageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (i *imageResource) checkResourceStatus(ctx context.Context, imageLabel string) (*govpsie.CustomImage, bool, error) {
	images, err := i.client.Image.List(ctx, nil)
	if err != nil {
		return nil, false, err
	}

	for _, image := range images {
		if image.ImageLabel == imageLabel {
			return &image, true, nil
		}
	}

	return nil, false, nil
}
