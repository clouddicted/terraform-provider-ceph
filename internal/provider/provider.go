package provider

import (
	"context"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure CephProvider satisfies various provider interfaces.
var _ provider.Provider = &CephProvider{}

// CephProvider defines the provider implementation.
type CephProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// CephProviderModel describes the provider data model.
type CephProviderModel struct {
	URL      types.String `tfsdk:"url"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Insecure types.Bool   `tfsdk:"insecure"`
}

func (p *CephProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "ceph"
	resp.Version = p.version
}

func (p *CephProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				MarkdownDescription: "The Ceph Dashboard URL (e.g., https://ceph-dashboard.example.com:8443)",
				Required:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "The username for authentication",
				Required:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "The password for authentication",
				Required:            true,
				Sensitive:           true,
			},
			"insecure": schema.BoolAttribute{
				MarkdownDescription: "Whether to skip TLS verification. Default: false.",
				Optional:            true,
			},
		},
	}
}

func (p *CephProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var data CephProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	insecure := false
	if !data.Insecure.IsNull() {
		insecure = data.Insecure.ValueBool()
	}

	c, err := client.NewClient(data.URL.ValueString(), data.Username.ValueString(), data.Password.ValueString(), insecure)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", "Unable to create Ceph client: "+err.Error())
		return
	}

	resp.DataSourceData = c
	resp.ResourceData = c
}

func (p *CephProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCephPoolResource,
		NewCephUserResource,
	}
}

func (p *CephProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCephPoolDataSource,
		NewCephUserDataSource,
		NewCephClusterDataSource,
		NewCephMonitorsDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &CephProvider{
			version: version,
		}
	}
}
