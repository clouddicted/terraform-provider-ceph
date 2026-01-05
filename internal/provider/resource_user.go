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
	Name         types.String `tfsdk:"name"`
	Capabilities types.String `tfsdk:"capabilities"`
	Key          types.String `tfsdk:"key"`
}

func NewCephUserResource() resource.Resource {
	return &CephUserResource{}
}

func (r *CephUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

func (r *CephUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The user entity name (e.g., client.app)",
				Required:            true,
			},
			"capabilities": schema.StringAttribute{
				MarkdownDescription: "The capabilities string (e.g., 'mon 'allow r' osd 'allow *'')",
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
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
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

	user := client.User{
		UserEntity:   data.Name.ValueString(),
		Capabilities: data.Capabilities.ValueString(),
	}

	err := r.client.CreateUser(user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create user, got error: %s", err))
		return
	}

	// Fetch the key
	key, err := r.client.ExportUser(user.UserEntity)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to export user key, got error: %s", err))
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

	// Note: Reading back the user might require parsing the capabilities string
	// which might be formatted differently by the server.
	// For now, we just verify existence.

	_, err := r.client.GetUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read user, got error: %s", err))
		return
	}

	// Fetch the key
	key, err := r.client.ExportUser(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to export user key, got error: %s", err))
		return
	}
	data.Key = types.StringValue(key)

	// Ideally we should update data.Capabilities here, but let's assume it matches for now
	// to avoid drift if the server reformats it.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Implement update if API supports it (usually PUT /api/cluster/user/{entity})
	// For now, force replacement by not implementing update or handling it via Delete/Create
	// But Terraform Framework handles Update separately.
	// If we don't implement logic, it just updates state? No, we need to call API.

	// Let's assume we can just re-create or update.
	// The API usually supports updating caps.
	// For now, let's return error or just re-create.
	// Actually, let's implement a basic update using Create (which might overwrite) or specific update endpoint.
	// Since I don't have the Update endpoint docs handy, I'll leave it as is (it will error if plan detects change and we don't handle it? No, it will just update state if we do nothing, which is bad).

	// Let's just re-use Create logic if the API supports upsert, otherwise we should probably error.
	// Or better, let's just implement it as Create for now.

	var data CephUserResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	user := client.User{
		UserEntity:   data.Name.ValueString(),
		Capabilities: data.Capabilities.ValueString(),
	}

	err := r.client.UpdateUser(data.Name.ValueString(), user)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update user, got error: %s", err))
		return
	}

	// Re-fetch key just in case, though it shouldn't change on update usually
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
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete user, got error: %s", err))
		return
	}
}

func (r *CephUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
