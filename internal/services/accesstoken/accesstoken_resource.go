package accesstoken

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource              = &accessTokenResource{}
	_ resource.ResourceWithConfigure = &accessTokenResource{}
)

type accessTokenResource struct {
	client *govpsie.Client
}

type accessTokenResourceModel struct {
	Identifier     types.String `tfsdk:"identifier"`
	Name           types.String `tfsdk:"name"`
	AccessToken    types.String `tfsdk:"access_token"`
	ExpirationDate types.String `tfsdk:"expiration_date"`
	CreatedOn      types.String `tfsdk:"created_on"`
}

func NewAccessTokenResource() resource.Resource {
	return &accessTokenResource{}
}

func (a *accessTokenResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_access_token"
}

func (a *accessTokenResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"identifier": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"access_token": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"expiration_date": schema.StringAttribute{
				Required: true,
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

func (a *accessTokenResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	a.client = client
}

func (a *accessTokenResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan accessTokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := a.client.AccessToken.Create(ctx, plan.Name.ValueString(), plan.AccessToken.ValueString(), plan.ExpirationDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error creating access token", err.Error())
		return
	}

	token, err := a.GetTokenByName(ctx, plan.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error reading access token after creation", err.Error())
		return
	}

	plan.Identifier = types.StringValue(token.AccessTokenIdentifier)
	plan.CreatedOn = types.StringValue(token.CreatedOn)

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (a *accessTokenResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state accessTokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	tokens, err := a.client.AccessToken.List(ctx, nil)
	if err != nil {
		resp.Diagnostics.AddError("Error reading access tokens", err.Error())
		return
	}

	found := false
	for _, token := range tokens {
		if token.AccessTokenIdentifier == state.Identifier.ValueString() {
			state.Name = types.StringValue(token.Name)
			state.ExpirationDate = types.StringValue(token.ExpirationDate)
			state.CreatedOn = types.StringValue(token.CreatedOn)
			found = true
			break
		}
	}

	if !found {
		resp.State.RemoveResource(ctx)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (a *accessTokenResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan accessTokenResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state accessTokenResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := a.client.AccessToken.Update(ctx, state.Identifier.ValueString(), plan.Name.ValueString(), plan.ExpirationDate.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Error updating access token", err.Error())
		return
	}

	state.Name = plan.Name
	state.ExpirationDate = plan.ExpirationDate

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (a *accessTokenResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state accessTokenResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := a.client.AccessToken.Delete(ctx, state.Identifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting access token",
			"couldn't delete access token, unexpected error: "+err.Error(),
		)
	}
}

func (a *accessTokenResource) GetTokenByName(ctx context.Context, name string) (*govpsie.AccessToken, error) {
	tokens, err := a.client.AccessToken.List(ctx, nil)
	if err != nil {
		return nil, err
	}

	for _, token := range tokens {
		if token.Name == name {
			return &token, nil
		}
	}

	return nil, fmt.Errorf("access token with name %s not found", name)
}
