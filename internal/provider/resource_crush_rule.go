package provider

import (
	"context"
	"fmt"

	"github.com/clouddicted/terraform-provider-ceph/internal/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var _ resource.Resource = &CephCrushRuleResource{}
var _ resource.ResourceWithConfigure = &CephCrushRuleResource{}
var _ resource.ResourceWithImportState = &CephCrushRuleResource{}

type CephCrushRuleResource struct {
	client *client.Client
}

type CephCrushRuleResourceModel struct {
	Name          types.String `tfsdk:"name"`
	Root          types.String `tfsdk:"root"`
	FailureDomain types.String `tfsdk:"failure_domain"`
	DeviceClass   types.String `tfsdk:"device_class"`
	RuleID        types.Int64  `tfsdk:"rule_id"`
}

func NewCephCrushRuleResource() resource.Resource {
	return &CephCrushRuleResource{}
}

func (r *CephCrushRuleResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_crush_rule"
}

func (r *CephCrushRuleResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a Ceph CRUSH rule for data placement",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				MarkdownDescription: "The name of the CRUSH rule",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"root": schema.StringAttribute{
				MarkdownDescription: "The root bucket (e.g., default, host name)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"failure_domain": schema.StringAttribute{
				MarkdownDescription: "The failure domain type (e.g., host, osd)",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"device_class": schema.StringAttribute{
				MarkdownDescription: "The device class (hdd, ssd, or empty for all)",
				Optional:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"rule_id": schema.Int64Attribute{
				MarkdownDescription: "The CRUSH rule ID assigned by Ceph",
				Computed:            true,
			},
		},
	}
}

func (r *CephCrushRuleResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *CephCrushRuleResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data CephCrushRuleResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule := client.CrushRule{
		Name:          data.Name.ValueString(),
		Root:          data.Root.ValueString(),
		FailureDomain: data.FailureDomain.ValueString(),
		DeviceClass:   data.DeviceClass.ValueString(),
	}

	err := r.client.CreateCrushRule(rule)
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to create CRUSH rule: %s", err))
		return
	}

	// Read back to get rule_id
	created, err := r.client.GetCrushRule(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read created CRUSH rule: %s", err))
		return
	}
	data.RuleID = types.Int64Value(int64(created.RuleID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephCrushRuleResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data CephCrushRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	rule, err := r.client.GetCrushRule(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read CRUSH rule: %s", err))
		return
	}

	data.RuleID = types.Int64Value(int64(rule.RuleID))

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *CephCrushRuleResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// CRUSH rules cannot be updated, only replaced (handled by RequiresReplace)
	resp.Diagnostics.AddError("Update Not Supported", "CRUSH rules cannot be updated. Changes require replacement.")
}

func (r *CephCrushRuleResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data CephCrushRuleResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteCrushRule(data.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to delete CRUSH rule: %s", err))
		return
	}
}

func (r *CephCrushRuleResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("name"), req, resp)
}
