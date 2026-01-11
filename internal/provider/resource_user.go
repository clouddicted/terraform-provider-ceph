package provider

import (
	"context"
	"fmt"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CephUserResource{}
var _ resource.ResourceWithConfigure = &CephUserResource{}

type CephUserResource struct {
	client *client.Client
}

type CephUserResourceModel struct {
	Name  types.String `tfsdk:"name"`
	Pools types.List   `tfsdk:"pools"`
	Key   types.String `tfsdk:"key"`
}

func NewCephUserResource() resource.Resource {
	return &CephUserResource{}
}

func (r *CephUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *CephUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Ceph user with RBD access to specified pools",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The user entity name (e.g., client.myapp)",
				Required:            true,
			},
			"pools": schema.ListAttribute{
				MarkdownDescription: "List of pool names the user can access with RBD profile",
				ElementType:         types.StringType,
				Required:            true,
			},
			"key": schema.StringAttribute{
				MarkdownDescription: "The exported keyring/key for the user",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *CephUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}

func (r *CephUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CephUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pools []string
	resp.Diagnostics.Append(data.Pools.ElementsAs(ctx, &pools, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.CreateUser(data.Name.ValueString(), pools)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user: %s", err))
		return
	}

	key, err := r.client.ExportUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to export user key: %s", err))
		return
	}
	data.Key = types.StringValue(key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CephUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.client.GetUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user: %s", err))
		return
	}

	key, err := r.client.ExportUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to export user key: %s", err))
		return
	}
	data.Key = types.StringValue(key)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CephUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var pools []string
	resp.Diagnostics.Append(data.Pools.ElementsAs(ctx, &pools, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.UpdateUser(data.Name.ValueString(), pools)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user: %s", err))
		return
	}

	key, err := r.client.ExportUser(data.Name.ValueString())
	if err == nil {
		data.Key = types.StringValue(key)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CephUserResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user: %s", err))
		return
	}
}

func (r *CephUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
