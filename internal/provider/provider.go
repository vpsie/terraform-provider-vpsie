// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/vpsie/govpsie"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/accesstoken"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/backup"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/bucket"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/datacenter"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/domain"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/fip"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/firewall"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/gateway"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/image"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/ip"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/kubernetes"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/loadbalancer"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/monitoring"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/project"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/script"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/server"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/snapshot"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/sshkey"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/storage"
	"github.com/vpsie/terraform-provider-vpsie/internal/services/vpc"
	"golang.org/x/oauth2"
)

// Ensure VpsieProvider satisfies various provider interfaces.
var _ provider.Provider = &VpsieProvider{}

const (
	userAgent = "vpsie-terraform-provider/1.0.0"
)

// VpsieProvider defines the provider implementation.
type VpsieProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// VpsieProviderModel describes the provider data model.
type VpsieProviderModel struct {
	AccessToken types.String `tfsdk:"access_token"`
}

func (p *VpsieProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "vpsie"
	resp.Version = p.version
}

func (p *VpsieProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"access_token": schema.StringAttribute{
				MarkdownDescription: "Example provider attribute",
				Optional:            true,
			},
		},
	}
}

func (p *VpsieProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data VpsieProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if data.AccessToken.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Unknown Vpsie Access Token",
			"The provider cannot create the Vpsie API client as there is an unknown configuration value for the Vpsie API access token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the VPSIE_ACCESS_TOKEN environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	accessToken := os.Getenv("VPSIE_ACCESS_TOKEN")

	if !data.AccessToken.IsNull() {
		accessToken = data.AccessToken.ValueString()
	}

	if accessToken == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("access_token"),
			"Missing Vpsie Access Token",
			"The provider cannot create the Vpsie API client as there is a missing or empty value for the Vpsie API access token. "+
				"Set the access token value in the configuration or use the VPSIE_ACCESS_TOKEN environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	tflog.Debug(ctx, "Creating Vpsie client")

	client := govpsie.NewClient(oauth2.NewClient(context.Background(), nil))

	client.SetUserAgent(userAgent)
	client.SetRequestHeaders(map[string]string{
		"Vpsie-Auth": accessToken,
	})

	// Make the HashiCups client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Vpsie client created", map[string]any{"success": true})
}

func (p *VpsieProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		storage.NewStorageResource,
		storage.NewStorageAttachmentResource,
		storage.NewStorageSnapshotResource,
		vpc.NewVpcResource,
		script.NewScriptResource,
		server.NewServerResource,
		image.NewImageResource,
		snapshot.NewServerSnapshotResource,
		sshkey.NewSshkeyResource,
		project.NewProjectResource,
		domain.NewDomainResource,
		gateway.NewGatewayResource,
		backup.NewBackupResource,
		firewall.NewFirewallResource,
		kubernetes.NewKubernetesResource,
		loadbalancer.NewLoadbalancerResource,
		kubernetes.NewKubernetesGroupResource,
		fip.NewFipResource,
		bucket.NewBucketResource,
		domain.NewDnsRecordResource,
		domain.NewReverseDnsResource,
		backup.NewBackupPolicyResource,
		snapshot.NewSnapshotPolicyResource,
		monitoring.NewMonitoringRuleResource,
		accesstoken.NewAccessTokenResource,
		firewall.NewFirewallAttachmentResource,
		vpc.NewVpcServerAssignmentResource,
	}
}

func (p *VpsieProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		storage.NewStorageDataSource,
		storage.NewStorageSnapshotDataSource,
		vpc.NewVpcDataSource,
		script.NewScriptDataSource,
		server.NewServerDataSource,
		image.NewImageDataSource,
		snapshot.NewServerSnapshotDataSource,
		sshkey.NewSshKeyDataSource,
		project.NewProjectDataSource,
		domain.NewDomainDataSource,
		gateway.NewGatewayDataSource,
		backup.NewBackupDataSource,
		firewall.NewFirewallDataSource,
		kubernetes.NewKubernetesDataSource,
		loadbalancer.NewLoadbalancerDataSource,
		kubernetes.NewKubernetesGroupDataSource,
		datacenter.NewDatacenterDataSource,
		fip.NewFipDataSource,
		bucket.NewBucketDataSource,
		backup.NewBackupPolicyDataSource,
		snapshot.NewSnapshotPolicyDataSource,
		monitoring.NewMonitoringRuleDataSource,
		accesstoken.NewAccessTokenDataSource,
		ip.NewIPDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &VpsieProvider{
			version: version,
		}
	}
}
