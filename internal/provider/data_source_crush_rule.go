package provider

import (
	"context"
	"fmt"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CephCrushRuleDataSource{}
var _ datasource.DataSourceWithConfigure = &CephCrushRuleDataSource{}

type CephCrushRuleDataSource struct {
	client *client.Client
}

type CephCrushRuleDataSourceModel struct {
	Name   types.String `tfsdk:"name"`
	RuleID types.Int64  `tfsdk:"rule_id"`
}

func NewCephCrushRuleDataSource() datasource.DataSource {
	return &CephCrushRuleDataSource{}
}

func (d *CephCrushRuleDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_crush_rule"
}

func (d *CephCrushRuleDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Look up an existing Ceph CRUSH rule",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the CRUSH rule",
				Required:            true,
			},
			"rule_id": schema.Int64Attribute{
				MarkdownDescription: "The CRUSH rule ID",
				Computed:            true,
			},
		},
	}
}

func (d *CephCrushRuleDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected DataSource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}

func (d *CephCrushRuleDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CephCrushRuleDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := d.client.GetCrushRule(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read CRUSH rule: %s", err))
		return
	}

	data.RuleID = types.Int64Value(int64(rule.RuleID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
