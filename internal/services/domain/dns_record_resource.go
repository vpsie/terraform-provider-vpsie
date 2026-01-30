package domain

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource              = &dnsRecordResource{}
	_ resource.ResourceWithConfigure = &dnsRecordResource{}
)

type dnsRecordResource struct {
	client *govpsie.Client
}

type dnsRecordResourceModel struct {
	ID               types.String `tfsdk:"id"`
	DomainIdentifier types.String `tfsdk:"domain_identifier"`
	Name             types.String `tfsdk:"name"`
	Content          types.String `tfsdk:"content"`
	Type             types.String `tfsdk:"type"`
	TTL              types.Int64  `tfsdk:"ttl"`
}

func NewDnsRecordResource() resource.Resource {
	return &dnsRecordResource{}
}

func (d *dnsRecordResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_dns_record"
}

func (d *dnsRecordResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"domain_identifier": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
			},
			"content": schema.StringAttribute{
				Required: true,
			},
			"type": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"ttl": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

func (d *dnsRecordResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	d.client = client
}

func (d *dnsRecordResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan dnsRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ttl := 3600
	if !plan.TTL.IsNull() && !plan.TTL.IsUnknown() {
		ttl = int(plan.TTL.ValueInt64())
	}

	createReq := govpsie.CreateDnsRecordReq{
		DomainIdentifier: plan.DomainIdentifier.ValueString(),
		Record: govpsie.Record{
			Name:    plan.Name.ValueString(),
			Content: plan.Content.ValueString(),
			Type:    plan.Type.ValueString(),
			TTL:     ttl,
		},
	}

	err := d.client.Domain.CreateDnsRecord(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Error creating DNS record", err.Error())
		return
	}

	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", plan.DomainIdentifier.ValueString(), plan.Type.ValueString(), plan.Name.ValueString()))
	plan.TTL = types.Int64Value(int64(ttl))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *dnsRecordResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state dnsRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// DNS records don't have a dedicated Get API — state is maintained from Create/Update
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

func (d *dnsRecordResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan dnsRecordResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state dnsRecordResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ttl := 3600
	if !plan.TTL.IsNull() && !plan.TTL.IsUnknown() {
		ttl = int(plan.TTL.ValueInt64())
	}

	oldTTL := 3600
	if !state.TTL.IsNull() && !state.TTL.IsUnknown() {
		oldTTL = int(state.TTL.ValueInt64())
	}

	updateReq := &govpsie.UpdateDnsRecordReq{
		DomainIdentifier: plan.DomainIdentifier.ValueString(),
		Current: govpsie.Record{
			Name:    state.Name.ValueString(),
			Content: state.Content.ValueString(),
			Type:    state.Type.ValueString(),
			TTL:     oldTTL,
		},
		New: govpsie.Record{
			Name:    plan.Name.ValueString(),
			Content: plan.Content.ValueString(),
			Type:    plan.Type.ValueString(),
			TTL:     ttl,
		},
	}

	err := d.client.Domain.UpdateDnsRecord(ctx, updateReq)
	if err != nil {
		resp.Diagnostics.AddError("Error updating DNS record", err.Error())
		return
	}

	plan.TTL = types.Int64Value(int64(ttl))
	plan.ID = types.StringValue(fmt.Sprintf("%s/%s/%s", plan.DomainIdentifier.ValueString(), plan.Type.ValueString(), plan.Name.ValueString()))

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
}

func (d *dnsRecordResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state dnsRecordResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	ttl := 3600
	if !state.TTL.IsNull() && !state.TTL.IsUnknown() {
		ttl = int(state.TTL.ValueInt64())
	}

	record := &govpsie.Record{
		Name:    state.Name.ValueString(),
		Content: state.Content.ValueString(),
		Type:    state.Type.ValueString(),
		TTL:     ttl,
	}

	err := d.client.Domain.DeleteDnsRecord(ctx, state.DomainIdentifier.ValueString(), record)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting DNS record",
			"couldn't delete DNS record, unexpected error: "+err.Error(),
		)
	}
}
