package provider

import (
	"context"
	"fmt"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CephMonitorsDataSource{}
var _ datasource.DataSourceWithConfigure = &CephMonitorsDataSource{}

type CephMonitorsDataSource struct {
	client *client.Client
}

type CephMonitorsDataSourceModel struct {
	Monitors []CephMonitorModel `tfsdk:"monitors"`
}

type CephMonitorModel struct {
	Name       types.String `tfsdk:"name"`
	Rank       types.Int64  `tfsdk:"rank"`
	Addr       types.String `tfsdk:"addr"`
	PublicAddr types.String `tfsdk:"public_addr"`
}

func NewCephMonitorsDataSource() datasource.DataSource {
	return &CephMonitorsDataSource{}
}

func (d *CephMonitorsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_monitors"
}

func (d *CephMonitorsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Data source for Ceph monitors.",
		Attributes: map[string]schema.Attribute{
			"monitors": schema.ListNestedAttribute{
				MarkdownDescription: "List of monitors in the cluster.",
				Computed:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							MarkdownDescription: "The name of the monitor.",
							Computed:            true,
						},
						"rank": schema.Int64Attribute{
							MarkdownDescription: "The rank of the monitor.",
							Computed:            true,
						},
						"addr": schema.StringAttribute{
							MarkdownDescription: "The address (IP:Port) of the monitor.",
							Computed:            true,
						},
						"public_addr": schema.StringAttribute{
							MarkdownDescription: "The public address of the monitor.",
							Computed:            true,
						},
					},
				},
			},
		},
	}
}

func (d *CephMonitorsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CephMonitorsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CephMonitorsDataSourceModel

	mons, err := d.client.GetMonitors()
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read monitors, got error: %s", err))
		return
	}

	for _, mon := range mons {
		data.Monitors = append(data.Monitors, CephMonitorModel{
			Name:       types.StringValue(mon.Name),
			Rank:       types.Int64Value(int64(mon.Rank)),
			Addr:       types.StringValue(mon.Addr),
			PublicAddr: types.StringValue(mon.PublicAddr),
		})
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
