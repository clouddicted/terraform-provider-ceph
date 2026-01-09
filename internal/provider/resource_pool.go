package provider

import (
	"context"
	"fmt"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CephPoolResource{}
var _ resource.ResourceWithConfigure = &CephPoolResource{}

type CephPoolResource struct {
	client *client.Client
}

type CephPoolResourceModel struct {
	Name                types.String `tfsdk:"name"`
	PgNum               types.Int64  `tfsdk:"pg_num"`
	Type                types.String `tfsdk:"type"`
	PgAutoscaleMode     types.Bool   `tfsdk:"pg_autoscale_mode"`
	Size                types.Int64  `tfsdk:"size"`
	RuleName            types.String `tfsdk:"rule_name"`
	QuotaMaxBytes       types.Int64  `tfsdk:"quota_max_bytes"`
	ApplicationMetadata types.List   `tfsdk:"application_metadata"`
	RbdMirroring        types.Bool   `tfsdk:"rbd_mirroring"`
	// Configuration       types.Map    `tfsdk:"configuration"` // Simplified for now
}

func NewCephPoolResource() resource.Resource {
	return &CephPoolResource{}
}

func (r *CephPoolResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

func (r *CephPoolResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "The name of the pool.",
			},
			"pg_num": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(16),
				Description: "The number of placement groups. Default: 16.",
			},
			"type": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("replicated"),
				Description: "The pool type. Default: replicated.",
			},
			"pg_autoscale_mode": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(true),
				Description: "Enable PG autoscale mode. Default: true (on).",
			},
			"size": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(3),
				Description: "The replication size. Default: 3.",
			},
			"rule_name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("replicated_rule"),
				Description: "The CRUSH rule name. Default: replicated_rule.",
			},
			"quota_max_bytes": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Maximum bytes quota. Default: 0 (no limit).",
			},
			"application_metadata": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Computed:    true,
				Default:     listdefault.StaticValue(types.ListValueMust(types.StringType, []attr.Value{types.StringValue("rbd")})),
				Description: "List of application metadata tags (rbd, cephfs, rgw). Default: [rbd].",
			},
			"rbd_mirroring": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enable RBD mirroring. Default: false.",
			},
		},
	}
}

func (r *CephPoolResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CephPoolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CephPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var appMetadata []string
	if !data.ApplicationMetadata.IsNull() {
		resp.Diagnostics.Append(data.ApplicationMetadata.ElementsAs(ctx, &appMetadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	pgAutoscaleMode := "off"
	if data.PgAutoscaleMode.ValueBool() {
		pgAutoscaleMode = "on"
	}

	pool := client.Pool{
		PoolName:            data.Name.ValueString(),
		PgNum:               int(data.PgNum.ValueInt64()),
		Type:                data.Type.ValueString(),
		PgAutoscaleMode:     pgAutoscaleMode,
		Size:                int(data.Size.ValueInt64()),
		RuleName:            data.RuleName.ValueString(),
		QuotaMaxBytes:       data.QuotaMaxBytes.ValueInt64(),
		ApplicationMetadata: appMetadata,
		RbdMirroring:        data.RbdMirroring.ValueBool(),
	}

	err := r.client.CreatePool(pool)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create pool, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephPoolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CephPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	pool, err := r.client.GetPool(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read pool, got error: %s", err))
		return
	}

	data.PgNum = types.Int64Value(int64(pool.PgNum))
	data.Type = types.StringValue(pool.Type)
	data.PgAutoscaleMode = types.BoolValue(pool.PgAutoscaleMode == "on")
	data.Size = types.Int64Value(int64(pool.Size))
	data.QuotaMaxBytes = types.Int64Value(pool.QuotaMaxBytes)

	// Note: RuleName and RbdMirroring might not be returned in the simple GET response or might be named differently.
	// We'll map what we can.

	// ApplicationMetadata is tricky because GET returns a map, but we model it as a list of keys (strings) in Create.
	// If we want to support reading it back, we need to extract keys from the map.
	// For now, let's leave it null in Read to avoid state drift issues if we can't map it perfectly yet.
	// Or better, just don't update it if it's not in the response.

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephPoolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data CephPoolResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var appMetadata []string
	if !data.ApplicationMetadata.IsNull() {
		resp.Diagnostics.Append(data.ApplicationMetadata.ElementsAs(ctx, &appMetadata, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
	}

	pgAutoscaleMode := "off"
	if data.PgAutoscaleMode.ValueBool() {
		pgAutoscaleMode = "on"
	}

	pool := client.Pool{
		PoolName:            data.Name.ValueString(),
		PgNum:               int(data.PgNum.ValueInt64()),
		Type:                data.Type.ValueString(),
		PgAutoscaleMode:     pgAutoscaleMode,
		Size:                int(data.Size.ValueInt64()),
		RuleName:            data.RuleName.ValueString(),
		QuotaMaxBytes:       data.QuotaMaxBytes.ValueInt64(),
		ApplicationMetadata: appMetadata,
		RbdMirroring:        data.RbdMirroring.ValueBool(),
	}

	// We use the name from the plan, assuming name changes force replacement (handled by Terraform)
	// or if we support rename, we need the old name.
	// For now, let's assume name is the ID and doesn't change in Update.
	err := r.client.UpdatePool(data.Name.ValueString(), pool)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to update pool, got error: %s", err))
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephPoolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CephPoolResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeletePool(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete pool, got error: %s", err))
		return
	}
}

func (r *CephPoolResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
