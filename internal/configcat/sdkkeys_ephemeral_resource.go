package configcat

import (
	"context"
	"fmt"

	"github.com/configcat/terraform-provider-configcat/v5/internal/configcat/client"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &sdkKeyEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &sdkKeyEphemeralResource{}
)

func NewSdkKeyEphemeralResource() ephemeral.EphemeralResource {
	return &sdkKeyEphemeralResource{}
}

type sdkKeyEphemeralResource struct {
	client *client.Client
}

type sdkKeyEphemeralResourceModel struct {
	ConfigId        types.String `tfsdk:"config_id"`
	EnvironmentId   types.String `tfsdk:"environment_id"`
	PrimarySdkKey   types.String `tfsdk:"primary"`
	SecondarySdkKey types.String `tfsdk:"secondary"`
}

func (r *sdkKeyEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sdkkeys"
}

func (r *sdkKeyEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this ephemeral resource to access existing **" + SdkKeyResourceName + "s** without persisting them to the Terraform plan or state. Requires Terraform 1.10 or later.",

		Attributes: map[string]schema.Attribute{
			ConfigId: schema.StringAttribute{
				Description: "The ID of the " + ConfigResourceName + ".",
				Required:    true,
				Validators:  []validator.String{IsGuid()},
			},
			EnvironmentId: schema.StringAttribute{
				Description: "The ID of the " + EnvironmentResourceName + ".",
				Required:    true,
				Validators:  []validator.String{IsGuid()},
			},
			PrimarySdkKey: schema.StringAttribute{
				MarkdownDescription: "The primary SDK Key associated with your **Config** and **Environment**.",
				Computed:            true,
				Sensitive:           true,
			},
			SecondarySdkKey: schema.StringAttribute{
				MarkdownDescription: "The secondary SDK Key associated with your **Config** and **Environment**.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *sdkKeyEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Ephemeral Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.client = client
}

func (r *sdkKeyEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data sdkKeyEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sdkKeys, err := r.client.GetSdkKeys(data.ConfigId.ValueString(), data.EnvironmentId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read "+SdkKeyResourceName+"s, got error: %s", err))
		return
	}

	data.PrimarySdkKey = types.StringValue(sdkKeys.Primary)
	data.SecondarySdkKey = types.StringPointerValue(sdkKeys.Secondary.Get())

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
