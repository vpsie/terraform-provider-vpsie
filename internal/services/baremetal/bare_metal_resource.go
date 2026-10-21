package baremetal

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/setvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
)

const (
	defaultCreateTimeout = 60 * time.Minute
	defaultUpdateTimeout = 60 * time.Minute
	defaultDeleteTimeout = 30 * time.Minute

	// statusOrdered is the status of a build-to-order server that has not
	// been delivered yet.
	statusOrdered = "ordered"
)

var (
	_ resource.Resource                = &bareMetalResource{}
	_ resource.ResourceWithConfigure   = &bareMetalResource{}
	_ resource.ResourceWithImportState = &bareMetalResource{}
)

type bareMetalResource struct {
	client *govpsie.Client
}

type bareMetalResourceModel struct {
	ID                types.String   `tfsdk:"id"`
	Identifier        types.String   `tfsdk:"identifier"`
	DcIdentifier      types.String   `tfsdk:"dc_identifier"`
	PlanIdentifier    types.String   `tfsdk:"plan_identifier"`
	OsIdentifier      types.String   `tfsdk:"os_identifier"`
	Hostname          types.String   `tfsdk:"hostname"`
	Password          types.String   `tfsdk:"password"`
	SSHKeyIdentifiers types.Set      `tfsdk:"ssh_key_identifiers"`
	RemoteAccess      types.Object   `tfsdk:"remote_access"`
	DNSZoneIdentifier types.String   `tfsdk:"dns_zone_identifier"`
	ProjectIdentifier types.String   `tfsdk:"project_identifier"`
	ContractTerm      types.String   `tfsdk:"contract_term"`
	LinkSpeedMbps     types.Int64    `tfsdk:"link_speed_mbps"`
	UserData          types.String   `tfsdk:"user_data"`
	ScriptIdentifiers types.List     `tfsdk:"script_identifiers"`
	PrivateNetworkIDs types.Set      `tfsdk:"private_network_ids"`
	ReleaseReason     types.String   `tfsdk:"release_reason"`
	ReleaseNote       types.String   `tfsdk:"release_note"`
	OrderStatus       types.String   `tfsdk:"order_status"`
	DeliveryHours     types.Int64    `tfsdk:"delivery_hours"`
	IPAddress         types.String   `tfsdk:"ip_address"`
	Netmask           types.String   `tfsdk:"netmask"`
	Gateway           types.String   `tfsdk:"gateway"`
	Status            types.String   `tfsdk:"status"`
	OsPhase           types.String   `tfsdk:"os_phase"`
	PowerState        types.String   `tfsdk:"power_state"`
	OsName            types.String   `tfsdk:"os_name"`
	OsFamily          types.String   `tfsdk:"os_family"`
	DeployedAt        types.String   `tfsdk:"deployed_at"`
	Timeouts          timeouts.Value `tfsdk:"timeouts"`
}

type remoteAccessModel struct {
	Enabled            types.Bool   `tfsdk:"enabled"`
	WinRM              types.Bool   `tfsdk:"winrm"`
	FirewallIdentifier types.String `tfsdk:"firewall_identifier"`
}

var remoteAccessAttrTypes = map[string]attr.Type{
	"enabled":             types.BoolType,
	"winrm":               types.BoolType,
	"firewall_identifier": types.StringType,
}

// NewBareMetalResource returns the vpsie_bare_metal resource.
func NewBareMetalResource() resource.Resource {
	return &bareMetalResource{}
}

func (r *bareMetalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_bare_metal"
}

func (r *bareMetalResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	install := func(desc string) string { return desc + " " + installInputDescription }

	resp.Schema = schema.Schema{
		MarkdownDescription: "Orders a dedicated (bare-metal) server, installs its operating system and keeps " +
			"its hostname, project, administrator password and private networks in step with the configuration. " +
			"Destroying the resource releases the server, which wipes its disks; this needs `account_password` " +
			"on the provider.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The server identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"identifier": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "The server identifier.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"dc_identifier": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: install("Identifier of the datacenter to order the server in."),
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 64)},
				PlanModifiers:       []planmodifier.String{replaceInstallInput()},
			},
			"plan_identifier": schema.StringAttribute{
				Required: true,
				MarkdownDescription: install("Identifier of the hardware plan, from the `vpsie_bare_metal_plans` data source. " +
					"The API does not return it, so an imported server takes it from the configuration."),
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 64)},
				PlanModifiers: []planmodifier.String{replaceInstallInput()},
			},
			"os_identifier": schema.StringAttribute{
				Required: true,
				MarkdownDescription: install("Identifier of the operating system (`bmos_...`), from the `vpsie_bare_metal_os` " +
					"data source. The API does not return it, so an imported server takes it from the configuration."),
				Validators:    []validator.String{stringvalidator.LengthBetween(1, 64)},
				PlanModifiers: []planmodifier.String{replaceInstallInput()},
			},
			"hostname": schema.StringAttribute{
				Required:            true,
				MarkdownDescription: "Hostname of the server, at most 253 characters. Changed in place.",
				Validators: []validator.String{
					stringvalidator.LengthBetween(1, 253),
					stringvalidator.RegexMatches(hostnamePattern, "must be a hostname of letters, digits, dots and hyphens"),
				},
			},
			"password": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
				MarkdownDescription: "Administrator (root) password: 8 to 128 printable ASCII characters, without double quotes, " +
					"backslashes or double curly braces. Changing it resets the password in place: the server powers off, " +
					"boots a reset job and starts again, which takes several minutes.",
				Validators: []validator.String{passwordValidator{}},
			},
			"ssh_key_identifiers": schema.SetAttribute{
				Optional:            true,
				ElementType:         types.StringType,
				MarkdownDescription: install("Identifiers of the account's SSH keys to install (at most 10)."),
				Validators: []validator.Set{
					setvalidator.SizeAtMost(10),
					setvalidator.ValueStringsAre(stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9-]+$`), "must be an SSH key identifier")),
				},
				PlanModifiers: []planmodifier.Set{replaceInstallInput()},
			},
			"remote_access": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				MarkdownDescription: install("Remote-access firewall rules written at install time. Without the block the " +
					"server opens SSH (or RDP on Windows) to the internet."),
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(true),
						MarkdownDescription: "Open SSH (Linux) or RDP (Windows). Defaults to `true`.",
					},
					"winrm": schema.BoolAttribute{
						Optional:            true,
						Computed:            true,
						Default:             booldefault.StaticBool(false),
						MarkdownDescription: "Also open WinRM on a Windows server. Defaults to `false`.",
					},
					"firewall_identifier": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Identifier of an account firewall group to apply instead of the open rules.",
					},
				},
				PlanModifiers: []planmodifier.Object{objectplanmodifier.UseStateForUnknown(), replaceInstallInput()},
			},
			"dns_zone_identifier": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: install("Identifier of an account DNS zone in which to create the server's record."),
				PlanModifiers:       []planmodifier.String{replaceInstallInput()},
			},
			"project_identifier": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: "Identifier of the project the server belongs to. Changed in place (the server moves).",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 45)},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"contract_term": schema.StringAttribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: install("Billing term: `hourly` or `monthly`, one of the plan's `contract_terms`. Defaults to the plan's term."),
				Validators:          []validator.String{stringvalidator.OneOf("hourly", "monthly")},
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown(), replaceInstallInput()},
			},
			"link_speed_mbps": schema.Int64Attribute{
				Optional:            true,
				Computed:            true,
				MarkdownDescription: install("Uplink speed in Mbit/s, one of the plan's `link_speed_options`. Defaults to the plan's speed."),
				Validators:          []validator.Int64{int64validator.OneOf(1000, 10000, 25000, 40000, 100000)},
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown(), replaceInstallInput()},
			},
			"user_data": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
				MarkdownDescription: install("Cloud-init user data run on first boot: a document starting with `#cloud-config` " +
					"or a script starting with `#!`, at most 64 KB."),
				Validators:    []validator.String{userDataValidator{}},
				PlanModifiers: []planmodifier.String{replaceInstallInput()},
			},
			"script_identifiers": schema.ListAttribute{
				Optional:    true,
				ElementType: types.StringType,
				MarkdownDescription: install("Identifiers of the account's saved scripts to run on first boot, in the listed " +
					"order (at most 5, each once). Changing the order also orders a new server."),
				Validators: []validator.List{
					listvalidator.SizeAtMost(5),
					listvalidator.UniqueValues(),
					listvalidator.ValueStringsAre(
						stringvalidator.LengthAtMost(64),
						stringvalidator.RegexMatches(regexp.MustCompile(`^[A-Za-z0-9-]+$`), "must be a script identifier"),
					),
				},
				PlanModifiers: []planmodifier.List{replaceInstallInput()},
			},
			"private_network_ids": schema.SetAttribute{
				Optional:    true,
				Computed:    true,
				ElementType: types.Int64Type,
				MarkdownDescription: "IDs of the VPCs (`vpsie_vpc.id`) to attach, at most 8. Attached and detached in place; " +
					"Terraform waits until each change is done. Without the attribute Terraform leaves the attachments alone.",
				Validators: []validator.Set{
					setvalidator.SizeAtMost(8),
					setvalidator.ValueInt64sAre(int64validator.AtLeast(1)),
				},
				PlanModifiers: []planmodifier.Set{setplanmodifier.UseStateForUnknown()},
			},
			"release_reason": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Reason recorded when Terraform releases the server. Defaults to `Destroyed by Terraform`.",
				Validators:          []validator.String{stringvalidator.LengthBetween(1, 256)},
			},
			"release_note": schema.StringAttribute{
				Optional:            true,
				MarkdownDescription: "Note recorded when Terraform releases the server.",
				Validators:          []validator.String{stringvalidator.LengthAtMost(512)},
			},
			"order_status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "`sold` once the server is on the account, `pending_build` while a build-to-order " +
					"server is assembled.",
				PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"delivery_hours": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Hours a build-to-order server takes to deliver, as quoted at order time.",
				PlanModifiers:       []planmodifier.Int64{int64planmodifier.UseStateForUnknown()},
			},
			"ip_address": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Primary public IPv4 address.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"netmask": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Netmask of the primary address.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"gateway": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Gateway of the primary address.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"status": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "State shown in the portal: `ordered`, `installing`, `reinstalling`, `ready`, `failed`, " +
					"`powered_off`, `deleting` or `suspended`.",
			},
			"os_phase": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "OS state: `installing`, `reinstalling`, `password_reset`, `running` or `failed`.",
			},
			"power_state": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Power state reported by the server's BMC.",
			},
			"os_name": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Name of the installed operating system.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"os_family": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "`linux` or `windows`.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"deployed_at": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "When the operating system was installed.",
				PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
			},
			"timeouts": timeouts.Attributes(ctx, timeouts.Opts{
				Create:            true,
				Update:            true,
				Delete:            true,
				CreateDescription: "How long to wait for the order and the OS install (default 60m). An order whose answer is lost (no answer, 502/503/504, Cloudflare 52x) is resent with the same idempotency key within this time; one still unconfirmed at the end, or refused after a lost answer, is looked up by hostname and adopted.",
				UpdateDescription: "How long to wait for a password reset and private network changes (default 60m).",
				DeleteDescription: "How long to wait for the release (default 30m).",
			}),
		},
	}
}

func (r *bareMetalResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*govpsie.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *govpsie.Client, got %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

// deployRequest builds the order from the plan.
func deployRequest(ctx context.Context, plan *bareMetalResourceModel) (*govpsie.BareMetalDeployRequest, diag.Diagnostics) {
	var diags diag.Diagnostics

	req := &govpsie.BareMetalDeployRequest{
		DCIdentifier:      plan.DcIdentifier.ValueString(),
		PlanIdentifier:    plan.PlanIdentifier.ValueString(),
		OSIdentifier:      plan.OsIdentifier.ValueString(),
		Hostname:          plan.Hostname.ValueString(),
		Password:          plan.Password.ValueString(),
		DNSZoneIdentifier: plan.DNSZoneIdentifier.ValueString(),
		ProjectIdentifier: plan.ProjectIdentifier.ValueString(),
		ContractTerm:      plan.ContractTerm.ValueString(),
		LinkSpeedMbps:     int(plan.LinkSpeedMbps.ValueInt64()),
		UserData:          plan.UserData.ValueString(),
	}

	if !plan.SSHKeyIdentifiers.IsNull() && !plan.SSHKeyIdentifiers.IsUnknown() {
		diags.Append(plan.SSHKeyIdentifiers.ElementsAs(ctx, &req.SSHKeyIdentifiers, false)...)
	}
	if !plan.ScriptIdentifiers.IsNull() && !plan.ScriptIdentifiers.IsUnknown() {
		diags.Append(plan.ScriptIdentifiers.ElementsAs(ctx, &req.ScriptIdentifiers, false)...)
	}
	if !plan.RemoteAccess.IsNull() && !plan.RemoteAccess.IsUnknown() {
		var ra remoteAccessModel
		diags.Append(plan.RemoteAccess.As(ctx, &ra, objectAsOptions)...)
		req.RemoteAccess = &govpsie.BareMetalRemoteAccess{
			Enabled:            ra.Enabled.ValueBool(),
			WinRM:              ra.WinRM.ValueBool(),
			FirewallIdentifier: ra.FirewallIdentifier.ValueStringPointer(),
		}
	}

	return req, diags
}

// defaultRemoteAccess is what the API applies when an order has no
// remoteAccess: SSH or RDP open, no WinRM, no firewall group.
func defaultRemoteAccess() types.Object {
	return types.ObjectValueMust(remoteAccessAttrTypes, map[string]attr.Value{
		"enabled":             types.BoolValue(true),
		"winrm":               types.BoolValue(false),
		"firewall_identifier": types.StringNull(),
	})
}

func remoteAccessValue(ra *govpsie.BareMetalRemoteAccess) types.Object {
	if ra == nil {
		return types.ObjectNull(remoteAccessAttrTypes)
	}

	return types.ObjectValueMust(remoteAccessAttrTypes, map[string]attr.Value{
		"enabled":             types.BoolValue(ra.Enabled),
		"winrm":               types.BoolValue(ra.WinRM),
		"firewall_identifier": types.StringPointerValue(ra.FirewallIdentifier),
	})
}

// networkIDs returns the IDs of the attached (or attaching) private networks.
func networkIDs(list *govpsie.BareMetalPrivateNetworks) []int64 {
	ids := []int64{}
	if list == nil {
		return ids
	}
	for _, item := range list.Items {
		if item.Status == govpsie.BareMetalNetworkAttached || item.Status == govpsie.BareMetalNetworkAttaching {
			ids = append(ids, item.NetworkID)
		}
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })

	return ids
}

func int64Set(ids []int64) types.Set {
	elems := make([]attr.Value, 0, len(ids))
	for _, id := range ids {
		elems = append(elems, types.Int64Value(id))
	}

	return types.SetValueMust(types.Int64Type, elems)
}

func setInt64s(ctx context.Context, set types.Set) ([]int64, diag.Diagnostics) {
	ids := []int64{}
	if set.IsNull() || set.IsUnknown() {
		return ids, nil
	}
	diags := set.ElementsAs(ctx, &ids, false)

	return ids, diags
}

// objectAsOptions converts the remote_access object as is.
var objectAsOptions = basetypes.ObjectAsOptions{}

func isSet(v attr.Value) bool {
	return !v.IsNull() && !v.IsUnknown()
}

// applyServer copies what the API returns about the server into m.
//
// Install-time inputs (plan, OS, SSH keys, user data, scripts, DNS zone,
// remote access, contract term, link speed) are written only where m has no
// value yet, that is after an import or for a value the configuration left
// to the API: a change made outside Terraform must not plan a new server.
func applyServer(m *bareMetalResourceModel, bm *govpsie.BareMetal) {
	if bm.Identifier != "" {
		m.Identifier = types.StringValue(bm.Identifier)
	}
	m.ID = m.Identifier

	if !isSet(m.DcIdentifier) && bm.DatacenterIdentifier != "" {
		m.DcIdentifier = types.StringValue(bm.DatacenterIdentifier)
	}
	if bm.Name != "" {
		m.Hostname = types.StringValue(bm.Name)
	}
	m.ProjectIdentifier = types.StringValue(bm.ProjectIdentifier)
	if bm.PrivateNetworks != nil || !isSet(m.PrivateNetworkIDs) {
		m.PrivateNetworkIDs = int64Set(networkIDs(bm.PrivateNetworks))
	}

	if !isSet(m.ContractTerm) {
		m.ContractTerm = types.StringNull()
		if bm.Contract != nil && bm.Contract.Term != "" {
			m.ContractTerm = types.StringValue(bm.Contract.Term)
		}
	}
	if !isSet(m.LinkSpeedMbps) {
		m.LinkSpeedMbps = types.Int64Null()
		if bm.LinkSpeedMbps > 0 {
			m.LinkSpeedMbps = types.Int64Value(int64(bm.LinkSpeedMbps))
		}
	}
	if !isSet(m.RemoteAccess) {
		m.RemoteAccess = remoteAccessValue(bm.RemoteAccess)
	}

	m.OrderStatus = types.StringValue(govpsie.BareMetalOrderSold)
	if m.DeliveryHours.IsUnknown() {
		m.DeliveryHours = types.Int64Null()
	}
	m.IPAddress = types.StringValue(bm.IPAddress)
	m.Netmask = types.StringValue(bm.Netmask)
	m.Gateway = types.StringValue(bm.Gateway)
	m.Status = types.StringValue(bm.DisplayState)
	m.OsPhase = types.StringValue(bm.OSPhase)
	m.PowerState = types.StringValue(bm.PowerState)
	m.OsName = types.StringValue(bm.OSName)
	m.OsFamily = types.StringValue(bm.OSFamily)
	m.DeployedAt = types.StringValue(bm.DeployedAt)
}

// applyOrder fills m from the order's answer, for a server that cannot be
// read yet (build to order) or whose wait failed. Every value the plan left
// unknown becomes known, so the state can be saved.
func applyOrder(m *bareMetalResourceModel, order *govpsie.BareMetalDeployment) {
	m.Identifier = types.StringValue(order.Identifier)
	m.ID = m.Identifier
	m.OrderStatus = types.StringValue(order.Status)
	m.DeliveryHours = types.Int64Value(int64(order.DeliveryHours))

	if !isSet(m.ContractTerm) {
		m.ContractTerm = types.StringNull()
		if order.ContractTerm != "" {
			m.ContractTerm = types.StringValue(order.ContractTerm)
		}
	}
	if !isSet(m.LinkSpeedMbps) {
		m.LinkSpeedMbps = types.Int64Null()
		if order.LinkSpeedMbps > 0 {
			m.LinkSpeedMbps = types.Int64Value(int64(order.LinkSpeedMbps))
		}
	}
	if !isSet(m.RemoteAccess) {
		m.RemoteAccess = defaultRemoteAccess()
	}
	if m.ProjectIdentifier.IsUnknown() {
		m.ProjectIdentifier = types.StringValue("")
	}
	if m.PrivateNetworkIDs.IsUnknown() {
		m.PrivateNetworkIDs = int64Set(nil)
	}

	status := statusOrdered
	if order.Status != govpsie.BareMetalOrderPendingBuild {
		status = "installing"
	}
	if m.IPAddress.IsUnknown() {
		m.IPAddress = types.StringValue(order.IPAddress)
	}
	for _, v := range []*types.String{&m.Netmask, &m.Gateway, &m.PowerState, &m.OsName, &m.OsFamily, &m.DeployedAt} {
		if v.IsUnknown() {
			*v = types.StringValue("")
		}
	}
	if m.Status.IsUnknown() {
		m.Status = types.StringValue(status)
	}
	if m.OsPhase.IsUnknown() {
		m.OsPhase = types.StringValue(order.OSPhase)
	}
}

func (r *bareMetalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan bareMetalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, createTimeout)
	defer cancel()

	order, diags := deployRequest(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	wantNetworks, diags := setInt64s(ctx, plan.PrivateNetworkIDs)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	deployment, adopted, err := deploy(ctx, r.client, order)
	if err != nil {
		resp.Diagnostics.AddError("Error ordering bare-metal server", err.Error())
		return
	}
	tflog.Info(ctx, "bare-metal server ordered", map[string]any{"identifier": deployment.Identifier, "status": deployment.Status, "adopted": adopted})

	// From here on the server exists: every exit saves the state, so a
	// failed wait leaves a tainted resource that the next apply replaces or
	// destroy releases, never an untracked server.
	plannedHostname, plannedNetworks := plan.Hostname, plan.PrivateNetworkIDs
	applyOrder(&plan, deployment)

	if adopted {
		resp.Diagnostics.AddWarning("Bare-metal order confirmed by lookup",
			fmt.Sprintf("The API did not confirm the order, but server %s named %q appeared in datacenter %s after it was sent. "+
				"Terraform adopted it instead of ordering another.", deployment.Identifier, plan.Hostname.ValueString(), plan.DcIdentifier.ValueString()))
	}
	if ctx.Err() != nil {
		// Found by the last lookup after the create timeout: no time is left
		// to wait for the install, so the server is saved as it reads now.
		readCtx, cancelRead := context.WithTimeout(context.WithoutCancel(ctx), orderLookupTimeout)
		defer cancelRead()
		plan.PrivateNetworkIDs = types.SetUnknown(types.Int64Type)
		r.refreshAfterError(readCtx, &plan, nil)
		resp.Diagnostics.AddError("Bare-metal server found after the create timeout",
			fmt.Sprintf("Server %s was found on the account after the create timeout and is saved in state as tainted. "+
				"Once its OS is running, keep it with terraform untaint; otherwise the next apply replaces it.", deployment.Identifier))
		resp.Diagnostics.Append(resp.State.Set(readCtx, &plan)...)
		return
	}

	if deployment.Status == govpsie.BareMetalOrderPendingBuild {
		resp.Diagnostics.AddWarning("Build-to-order server ordered",
			fmt.Sprintf("Server %s is built to order and delivered in about %d hours. Terraform keeps the order in state; "+
				"the OS is installed on delivery, and the next apply after delivery attaches the private networks.",
				deployment.Identifier, deployment.DeliveryHours))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}

	bm, err := waitForOS(ctx, r.client, deployment.Identifier, "the OS install")
	if bm != nil {
		applyServer(&plan, bm)
	}
	if err != nil {
		resp.Diagnostics.AddError("Error waiting for the bare-metal server",
			fmt.Sprintf("Server %s was ordered, but %s. It is saved in state as tainted; the next apply replaces it "+
				"and destroy releases it.", deployment.Identifier, err))
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
		return
	}
	// The reads during the wait saw the server before its networks were
	// attached; finish keeps the configured hostname and networks.
	plan.Hostname = plannedHostname

	for _, id := range wantNetworks {
		if err := attachNetwork(ctx, r.client, deployment.Identifier, id); err != nil {
			resp.Diagnostics.AddError("Error attaching a private network", err.Error())
			plan.PrivateNetworkIDs = types.SetUnknown(types.Int64Type)
			r.refreshAfterError(ctx, &plan, nil)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
	}

	r.finish(ctx, &plan, nil, plannedNetworks, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// refreshAfterError reads the server once more after a failed change, so the
// saved state shows what really applied (the private networks attached, the
// hostname). When the read fails too, the values the plan left unknown are
// taken from prior (the state before the change) or, on create, left empty.
func (r *bareMetalResource) refreshAfterError(ctx context.Context, m, prior *bareMetalResourceModel) {
	bm, err := r.client.BareMetal.Get(ctx, m.Identifier.ValueString())
	if err == nil {
		applyServer(m, bm)
		return
	}

	tflog.Warn(ctx, "could not read the bare-metal server after an error", map[string]any{"error": err.Error()})
	if prior != nil {
		keepComputed(m, prior)
		return
	}
	m.PrivateNetworkIDs = int64Set(nil)
}

// finish reads the server after a create or update. The hostname and the
// planned private networks, when configured, are kept as planned: the calls
// and waits before confirmed them, and Terraform requires the applied value
// of a configured attribute to equal the planned one. Networks left out of
// the configuration are taken from the read.
func (r *bareMetalResource) finish(ctx context.Context, m, prior *bareMetalResourceModel, plannedNetworks types.Set, diags *diag.Diagnostics) {
	hostname := m.Hostname
	bm, err := r.client.BareMetal.Get(ctx, m.Identifier.ValueString())
	if err != nil {
		diags.AddError("Error reading bare-metal server", err.Error())
		if prior != nil {
			keepComputed(m, prior)
		}
		return
	}
	applyServer(m, bm)
	m.Hostname = hostname
	if isSet(plannedNetworks) {
		m.PrivateNetworkIDs = plannedNetworks
	}
}

func (r *bareMetalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state bareMetalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	bm, err := r.client.BareMetal.Get(ctx, state.Identifier.ValueString())
	if err != nil {
		if govpsie.IsNotFound(err) {
			if state.OrderStatus.ValueString() == govpsie.BareMetalOrderPendingBuild {
				// A build-to-order server is not readable until delivery.
				resp.Diagnostics.AddWarning("Bare-metal server not delivered yet",
					fmt.Sprintf("Server %s is still being built; Terraform keeps the order in state.", state.Identifier.ValueString()))
				return
			}
			tflog.Warn(ctx, "bare-metal server not found, removing it from state", map[string]any{"identifier": state.Identifier.ValueString()})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Error reading bare-metal server", err.Error())
		return
	}

	applyServer(&state, bm)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *bareMetalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state bareMetalResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, updateTimeout)
	defer cancel()

	id := state.Identifier.ValueString()
	plan.Identifier = state.Identifier
	plan.ID = state.Identifier

	// saveProgress records what was applied before an error, so the next
	// plan retries only what is left.
	saveProgress := func() {
		r.refreshAfterError(ctx, &plan, &state)
		resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
	}

	if !plan.Hostname.Equal(state.Hostname) {
		if err := r.client.BareMetal.UpdateHostname(ctx, id, plan.Hostname.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error changing the bare-metal hostname", err.Error())
			resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
			return
		}
	}

	if isSet(plan.ProjectIdentifier) && !plan.ProjectIdentifier.Equal(state.ProjectIdentifier) {
		if err := r.client.BareMetal.MoveToProject(ctx, id, plan.ProjectIdentifier.ValueString()); err != nil {
			resp.Diagnostics.AddError("Error moving the bare-metal server to another project", err.Error())
			plan.ProjectIdentifier = types.StringUnknown()
			plan.Password, plan.PrivateNetworkIDs = state.Password, types.SetUnknown(types.Int64Type)
			saveProgress()
			return
		}
	}

	// A password that is null in state belongs to an imported server: the
	// configured one is adopted without a reset.
	if !state.Password.IsNull() && !plan.Password.Equal(state.Password) {
		err := retryWhileBusy(ctx, "the password reset", func() error {
			_, err := r.client.BareMetal.ResetPassword(ctx, id, &govpsie.BareMetalPasswordResetRequest{Password: plan.Password.ValueString()})
			return err
		}, codeBusy)
		if err == nil {
			_, err = waitForOS(ctx, r.client, id, "the password reset")
		}
		if err != nil {
			resp.Diagnostics.AddError("Error resetting the bare-metal password", err.Error())
			plan.Password, plan.PrivateNetworkIDs = state.Password, types.SetUnknown(types.Int64Type)
			saveProgress()
			return
		}
	}

	if isSet(plan.PrivateNetworkIDs) && !plan.PrivateNetworkIDs.Equal(state.PrivateNetworkIDs) {
		if err := r.applyNetworks(ctx, id, plan.PrivateNetworkIDs, state.PrivateNetworkIDs); err != nil {
			resp.Diagnostics.AddError("Error changing the bare-metal private networks", err.Error())
			plan.PrivateNetworkIDs = types.SetUnknown(types.Int64Type)
			saveProgress()
			return
		}
	}

	if state.OrderStatus.ValueString() == govpsie.BareMetalOrderPendingBuild {
		// Not delivered: nothing to read yet; the changes above were
		// state-only (adopted inputs, release reason).
		if _, err := r.client.BareMetal.Get(ctx, id); govpsie.IsNotFound(err) {
			keepComputed(&plan, &state)
			resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
			return
		}
	}

	r.finish(ctx, &plan, &state, plan.PrivateNetworkIDs, &resp.Diagnostics)
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// keepComputed copies the computed values of state into the unknown values
// of plan.
func keepComputed(plan, state *bareMetalResourceModel) {
	for _, pair := range []struct{ p, s *types.String }{
		{&plan.Status, &state.Status}, {&plan.OsPhase, &state.OsPhase}, {&plan.PowerState, &state.PowerState},
		{&plan.ProjectIdentifier, &state.ProjectIdentifier},
	} {
		if pair.p.IsUnknown() {
			*pair.p = *pair.s
		}
	}
	if plan.PrivateNetworkIDs.IsUnknown() {
		plan.PrivateNetworkIDs = state.PrivateNetworkIDs
	}
}

// applyNetworks detaches the networks only in state and attaches those only
// in the plan, one at a time: the API runs one change per server.
func (r *bareMetalResource) applyNetworks(ctx context.Context, id string, planned, current types.Set) error {
	want, diags := setInt64s(ctx, planned)
	if diags.HasError() {
		return fmt.Errorf("reading the planned private networks: %v", diags)
	}
	have, diags := setInt64s(ctx, current)
	if diags.HasError() {
		return fmt.Errorf("reading the attached private networks: %v", diags)
	}

	wantSet := make(map[int64]bool, len(want))
	for _, n := range want {
		wantSet[n] = true
	}
	haveSet := make(map[int64]bool, len(have))
	for _, n := range have {
		haveSet[n] = true
	}

	for _, n := range have {
		if !wantSet[n] {
			if err := detachNetwork(ctx, r.client, id, n); err != nil {
				return err
			}
		}
	}
	for _, n := range want {
		if !haveSet[n] {
			if err := attachNetwork(ctx, r.client, id, n); err != nil {
				return err
			}
		}
	}

	return nil
}

func (r *bareMetalResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state bareMetalResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}
	id := state.Identifier.ValueString()

	// Release is confirmed with the ACCOUNT password, never the server's.
	if r.client.AccountPassword == "" {
		resp.Diagnostics.AddError("Cannot release the bare-metal server without the account password",
			fmt.Sprintf("Releasing %s requires the password of the VPSie account that owns it. Set `account_password` "+
				"on the provider (or the VPSIE_ACCOUNT_PASSWORD environment variable) and try again.", id))
		return
	}

	deleteTimeout, diags := state.Timeouts.Delete(ctx, defaultDeleteTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, deleteTimeout)
	defer cancel()

	if _, err := r.client.BareMetal.Get(ctx, id); err != nil {
		if !govpsie.IsNotFound(err) {
			resp.Diagnostics.AddError("Error reading bare-metal server", err.Error())
			return
		}
		if state.OrderStatus.ValueString() == govpsie.BareMetalOrderPendingBuild {
			resp.Diagnostics.AddError("Cannot release a server that is still being built",
				fmt.Sprintf("Server %s is built to order and not delivered yet, and the API can release it only after "+
					"delivery. Wait for delivery and run destroy again, or ask support to cancel the order and then "+
					"remove it from state with `terraform state rm`.", id))
			return
		}
		// Already released.
		return
	}

	reason := state.ReleaseReason.ValueString()
	if reason == "" {
		reason = releaseReasonDefault
	}
	release := &govpsie.BareMetalReleaseRequest{
		Password: r.client.AccountPassword,
		Reason:   reason,
		Note:     state.ReleaseNote.ValueString(),
	}
	err := retryWhileBusy(ctx, "the release", func() error {
		return r.client.BareMetal.Release(ctx, id, release)
	}, codeBusy)
	if err != nil {
		if govpsie.IsNotFound(err) {
			return
		}
		resp.Diagnostics.AddError("Error releasing bare-metal server", err.Error())
		return
	}

	if err := waitForRelease(ctx, r.client, id); err != nil {
		resp.Diagnostics.AddError("Error waiting for the bare-metal server to be released", err.Error())
	}
}

// ImportState adopts a server by its identifier. The API does not return the
// plan, OS, password, SSH keys, user data, scripts or DNS zone a server was
// ordered with; the first apply after the import takes them from the
// configuration without replacing the server (see installInputNeedsReplace).
func (r *bareMetalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("identifier"), req, resp)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
}
