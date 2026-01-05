package provider

import (
	"context"
	"fmt"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CephPoolDataSource{}
var _ datasource.DataSourceWithConfigure = &CephPoolDataSource{}

type CephPoolDataSource struct {
	client *client.Client
}

type CephPoolDataSourceModel struct {
	Name                types.String `tfsdk:"name"`
	PgNum               types.Int64  `tfsdk:"pg_num"`
	Type                types.String `tfsdk:"type"`
	PgAutoscaleMode     types.String `tfsdk:"pg_autoscale_mode"`
	Size                types.Int64  `tfsdk:"size"`
	RuleName            types.String `tfsdk:"rule_name"`
	QuotaMaxBytes       types.Int64  `tfsdk:"quota_max_bytes"`
	ApplicationMetadata types.List   `tfsdk:"application_metadata"`
	RbdMirroring        types.Bool   `tfsdk:"rbd_mirroring"`
}

func NewCephPoolDataSource() datasource.DataSource {
	return &CephPoolDataSource{}
}

func (d *CephPoolDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (d *CephPoolDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"pg_num": schema.Int64Attribute{
				Computed: true,
			},
			"type": schema.StringAttribute{
				Computed: true,
			},
			"pg_autoscale_mode": schema.StringAttribute{
				Computed: true,
			},
			"size": schema.Int64Attribute{
				Computed: true,
			},
			"rule_name": schema.StringAttribute{
				Computed: true,
			},
			"quota_max_bytes": schema.Int64Attribute{
				Computed: true,
			},
			"application_metadata": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
			"rbd_mirroring": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *CephPoolDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CephPoolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CephPoolDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := d.client.GetPool(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read pool, got error: %s", err))
		return
	}

	data.PgNum = types.Int64Value(int64(pool.PgNum))
	data.Type = types.StringValue(pool.Type)
	data.PgAutoscaleMode = types.StringValue(pool.PgAutoscaleMode)
	data.Size = types.Int64Value(int64(pool.Size))
	data.QuotaMaxBytes = types.Int64Value(pool.QuotaMaxBytes)
	// Note: RuleName, RbdMirroring, ApplicationMetadata mapping logic should be consistent with Resource Read.
	// For now, we leave them null/unknown if not returned by GetPool or if mapping is complex.
	// Assuming GetPool populates what it can.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
