package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ datasource.DataSource = &CephUserDataSource{}
var _ datasource.DataSourceWithConfigure = &CephUserDataSource{}

type CephUserDataSource struct {
	client *client.Client
}

type CephUserDataSourceModel struct {
	Name  types.String `tfsdk:"name"`
	Pools types.List   `tfsdk:"pools"`
	Key   types.String `tfsdk:"key"`
}

func NewCephUserDataSource() datasource.DataSource {
	return &CephUserDataSource{}
}

func (d *CephUserDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (d *CephUserDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The user entity name (e.g., client.app)",
				Required:            true,
			},
			"pools": schema.ListAttribute{
				MarkdownDescription: "List of pool names the user can access",
				ElementType:         types.StringType,
				Computed:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The exported keyring/key for the user",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (d *CephUserDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CephUserDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data CephUserDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user, err := d.client.GetUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user: %s", err))
		return
	}

	// Extract pools from capabilities (caps is now a map[string]string)
	var pools []string
	for service, cap := range user.Caps {
		if service == "osd" && strings.HasPrefix(cap, "profile rbd pool=") {
			pool := strings.TrimPrefix(cap, "profile rbd pool=")
			pools = append(pools, pool)
		}
	}
	data.Pools, _ = types.ListValueFrom(ctx, types.StringType, pools)

	key, err := d.client.ExportUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to export user key: %s", err))
		return
	}
	data.Key = types.StringValue(key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
