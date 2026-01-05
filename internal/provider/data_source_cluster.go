package provider

import (
	"context"
	"fmt"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CephClusterDataSource{}
var _ datasource.DataSourceWithConfigure = &CephClusterDataSource{}

type CephClusterDataSource struct {
	client *client.Client
}

type CephClusterDataSourceModel struct {
	Fsid types.String `tfsdk:"fsid"`
}

func NewCephClusterDataSource() datasource.DataSource {
	return &CephClusterDataSource{}
}

func (d *CephClusterDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster"
}

func (d *CephClusterDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for Ceph cluster information.",
		Attributes: map[string]schema.Attribute{
			"fsid": schema.StringAttribute{
				MarkdownDescription: "The unique identifier (FSID) of the Ceph cluster.",
				Computed:            true,
			},
		},
	}
}

func (d *CephClusterDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *CephClusterDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CephClusterDataSourceModel

	// No config to read, just fetch data
	fsid, err := d.client.GetClusterFSID()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read cluster FSID, got error: %s", err))
		return
	}

	data.Fsid = types.StringValue(fsid)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
