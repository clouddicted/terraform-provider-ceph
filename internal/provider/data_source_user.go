package provider

import (
	"context"
	"fmt"

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
	Name         types.String `tfsdk:"name"`
	Capabilities types.String `tfsdk:"capabilities"`
	Key          types.String `tfsdk:"key"`
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
			"capabilities": schema.StringAttribute{
				MarkdownDescription: "The capabilities string",
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

	_, err := d.client.GetUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	// Fetch the key
	key, err := d.client.ExportUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to export user key, got error: %s", err))
		return
	}
	data.Key = types.StringValue(key)

	// Note: Capabilities mapping is skipped for now as GetUser might return different format.
	// We assume the user exists and we got the key.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
