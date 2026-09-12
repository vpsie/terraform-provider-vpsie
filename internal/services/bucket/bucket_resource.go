package bucket

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &bucketResource{}
	_ resource.ResourceWithConfigure   = &bucketResource{}
	_ resource.ResourceWithImportState = &bucketResource{}
)

type bucketResource struct {
	client *govpsie.Client
}

type bucketResourceModel struct {
	Identifier   types.String `tfsdk:"identifier"`
	BucketName   types.String `tfsdk:"bucket_name"`
	ProjectID    types.String `tfsdk:"project_id"`
	DataCenterID types.String `tfsdk:"datacenter_id"`
	FileListing  types.Bool   `tfsdk:"file_listing"`
	AccessKey    types.String `tfsdk:"access_key"`
	SecretKey    types.String `tfsdk:"secret_key"`
	EndPoint     types.String `tfsdk:"endpoint"`
	State        types.String `tfsdk:"state"`
	CreatedBy    types.String `tfsdk:"created_by"`
	CreatedOn    types.String `tfsdk:"created_on"`
}

func NewBucketResource() resource.Resource {
	return &bucketResource{}
}

func (b *bucketResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bucket"
}

func (b *bucketResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"bucket_name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"datacenter_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"file_listing": schema.BoolAttribute{
				Optional: true,
			},
			"access_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"secret_key": schema.StringAttribute{
				Computed:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"endpoint": schema.StringAttribute{
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
			"created_by": schema.StringAttribute{
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
		},
	}
}

func (b *bucketResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configuration Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	b.client = client
}

func (b *bucketResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bucketResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	fileListing := false
	if !plan.FileListing.IsNull() {
		fileListing = plan.FileListing.ValueBool()
	}

	createReq := &govpsie.CreateBucketReq{
		BucketName:   plan.BucketName.ValueString(),
		FileListing:  fileListing,
		ProjectId:    plan.ProjectID.ValueString(),
		DataCenterId: plan.DataCenterID.ValueString(),
	}

	err := b.client.Bucket.Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating bucket", err.Error())
		return
	}

	bucket, err := b.GetBucketByName(ctx, plan.BucketName.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading bucket after creation", err.Error())
		return
	}

	plan.Identifier = types.StringValue(bucket.Identifier)
	plan.AccessKey = types.StringValue(bucket.AccessKey)
	plan.SecretKey = types.StringValue(bucket.SecretKey)
	plan.EndPoint = types.StringValue(bucket.EndPoint)
	plan.State = types.StringValue(bucket.State)
	plan.CreatedBy = types.StringValue(bucket.CreatedBy)
	plan.CreatedOn = types.StringValue(bucket.CreatedOn)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (b *bucketResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bucketResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	bucket, err := b.client.Bucket.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading bucket",
			"couldn't read bucket "+state.Identifier.ValueString()+": "+err.Error(),
		)
		return
	}

	state.BucketName = types.StringValue(bucket.BucketName)
	state.AccessKey = types.StringValue(bucket.AccessKey)
	state.SecretKey = types.StringValue(bucket.SecretKey)
	state.EndPoint = types.StringValue(bucket.EndPoint)
	state.State = types.StringValue(bucket.State)
	state.CreatedBy = types.StringValue(bucket.CreatedBy)
	state.CreatedOn = types.StringValue(bucket.CreatedOn)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (b *bucketResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan bucketResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state bucketResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.FileListing.Equal(state.FileListing) && !plan.FileListing.IsNull() {
		_, err := b.client.Bucket.ToggleFileListing(ctx, state.Identifier.ValueString(), plan.FileListing.ValueBool())
		if err != nil {
			resp.Diagnostics.AddError("Error toggling file listing", err.Error())
			return
		}
		state.FileListing = plan.FileListing
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (b *bucketResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bucketResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := b.client.Bucket.Delete(ctx, state.Identifier.ValueString(), "terraform-destroy", "terraform-destroy")
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting bucket",
			"couldn't delete bucket, unexpected error: "+err.Error(),
		)
	}
}

func (b *bucketResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
}

func (b *bucketResource) GetBucketByName(ctx context.Context, name string) (*govpsie.Bucket, error) {
	buckets, err := b.client.Bucket.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, bucket := range buckets {
		if bucket.BucketName == name {
			return &bucket, nil
		}
	}

	return nil, fmt.Errorf("bucket with name %s not found", name)
}
