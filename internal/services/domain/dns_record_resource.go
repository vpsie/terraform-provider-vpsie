package domain

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

var (
	_ resource.Resource                = &dnsRecordResource{}
	_ resource.ResourceWithConfigure   = &dnsRecordResource{}
	_ resource.ResourceWithImportState = &dnsRecordResource{}
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
		MarkdownDescription: "Manages a single DNS record on a VPSie domain. Fully supported record " +
			"types are `A`, `AAAA`, `CNAME`, `TXT`, and `NS` (the record is addressed by " +
			"`name`, `type`, and `content`). Because the underlying API edits records a whole " +
			"record set (name + type) at a time, avoid managing several records that share the " +
			"same `name` and `type` unless each has distinct `content`; destroying multiple " +
			"records of one set in a single run can race — use `depends_on` to serialize them.",
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

	// Individual DNS records have no id; they are read back from the parent
	// domain, whose records are grouped by type and carry fully-qualified names.
	domain, err := d.client.Domain.GetDomainByIdentifier(ctx, state.DomainIdentifier.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading DNS record",
			"couldn't read parent domain: "+err.Error(),
		)
		return
	}

	// The parent domain is gone, so the record is too; drop it from state.
	if domain == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	wantName := normalizeRecordName(state.Name.ValueString(), domain.Domain)
	wantContent := strings.TrimSuffix(state.Content.ValueString(), ".")

	// Collect the records sharing this record's (normalized) name and type —
	// i.e. the rrset the managed record belongs to.
	var rrset []govpsie.DomainRecord
	for _, record := range domain.RecordsByType(state.Type.ValueString()) {
		if strings.EqualFold(record.Name, wantName) {
			rrset = append(rrset, record)
		}
	}

	// The whole rrset is gone, so the managed record is too; drop it from state.
	if len(rrset) == 0 {
		resp.State.RemoveResource(ctx)
		return
	}

	// Prefer an exact content match. The API normalizes/reshapes content for
	// several record types (IPv6 compaction, TXT quote stripping, MX/CAA/SRV
	// field splitting), so when the rrset holds a single record we adopt it even
	// if its stored content differs; only when several records share the
	// name+type do we rely on content to disambiguate.
	var match *govpsie.DomainRecord
	for i := range rrset {
		if strings.TrimSuffix(rrset[i].Content, ".") == wantContent {
			match = &rrset[i]
			break
		}
	}
	if match == nil && len(rrset) == 1 {
		match = &rrset[0]
	}

	// Keep the config-owned values as the user wrote them (short name,
	// un-normalized content) to avoid perpetual drift, and only refresh the
	// server-computed TTL. When content can't be disambiguated in a multi-record
	// rrset, leave state untouched rather than forcing a spurious recreate.
	if match != nil {
		state.TTL = types.Int64Value(int64(match.TTL))
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}

// normalizeRecordName mirrors the cloud/api checkName logic: a short record name
// is qualified with the domain, while "@"/"*" and already-qualified names are
// mapped to their canonical form.
func normalizeRecordName(name, domain string) string {
	switch name {
	case "@":
		return domain
	case "*":
		return "*." + domain
	}
	lname := strings.ToLower(name)
	ldomain := strings.ToLower(domain)
	if lname == ldomain || strings.HasSuffix(lname, "."+ldomain) {
		return name
	}
	return name + "." + domain
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

// ImportState imports a DNS record using the composite id
// "domain_identifier/type/name/content" (a record has no standalone id, so all
// four fields are needed to identify it uniquely). Read then refreshes the TTL.
func (d *dnsRecordResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	parts := strings.SplitN(req.ID, "/", 4)
	if len(parts) != 4 || parts[0] == "" || parts[1] == "" || parts[2] == "" || parts[3] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import id in the format \"domain_identifier/type/name/content\", got: %q", req.ID),
		)
		return
	}

	state := dnsRecordResourceModel{
		DomainIdentifier: types.StringValue(parts[0]),
		Type:             types.StringValue(parts[1]),
		Name:             types.StringValue(parts[2]),
		Content:          types.StringValue(parts[3]),
		ID:               types.StringValue(fmt.Sprintf("%s/%s/%s", parts[0], parts[1], parts[2])),
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
