package configcat

import (
	"context"
	"fmt"

	"github.com/configcat/terraform-provider-configcat/v5/internal/configcat/client"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral"
	"github.com/hashicorp/terraform-plugin-framework/ephemeral/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ ephemeral.EphemeralResource              = &webhookSigningKeysEphemeralResource{}
	_ ephemeral.EphemeralResourceWithConfigure = &webhookSigningKeysEphemeralResource{}
)

func NewWebhookSigningKeysEphemeralResource() ephemeral.EphemeralResource {
	return &webhookSigningKeysEphemeralResource{}
}

type webhookSigningKeysEphemeralResource struct {
	client *client.Client
}

type webhookSigningKeysEphemeralResourceModel struct {
	WebhookId types.Int64  `tfsdk:"webhook_id"`
	Key1      types.String `tfsdk:"key1"`
	Key2      types.String `tfsdk:"key2"`
}

func (r *webhookSigningKeysEphemeralResource) Metadata(ctx context.Context, req ephemeral.MetadataRequest, resp *ephemeral.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_webhook_signing_keys"
}

func (r *webhookSigningKeysEphemeralResource) Schema(ctx context.Context, req ephemeral.SchemaRequest, resp *ephemeral.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Use this ephemeral resource to access **" + WebhookSigningKeysResourceName + "** without persisting them to the Terraform plan or state. Requires Terraform 1.10 or later. [What is a " + WebhookResourceName + " in ConfigCat?](https://configcat.com/docs/advanced/notifications-webhooks/)",

		Attributes: map[string]schema.Attribute{
			WebhookId: schema.Int64Attribute{
				Description: "The ID of the " + WebhookResourceName + ".",
				Required:    true,
			},
			WebhookSigningKeyKey1: schema.StringAttribute{
				MarkdownDescription: "The first signing key.",
				Computed:            true,
				Sensitive:           true,
			},
			WebhookSigningKeyKey2: schema.StringAttribute{
				MarkdownDescription: "The second signing key.",
				Computed:            true,
				Sensitive:           true,
			},
		},
	}
}

func (r *webhookSigningKeysEphemeralResource) Configure(ctx context.Context, req ephemeral.ConfigureRequest, resp *ephemeral.ConfigureResponse) {
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

func (r *webhookSigningKeysEphemeralResource) Open(ctx context.Context, req ephemeral.OpenRequest, resp *ephemeral.OpenResponse) {
	var data webhookSigningKeysEphemeralResourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	webhookSigningKeys, err := r.client.GetWebhookSigningKeys(int32(data.WebhookId.ValueInt64()))
	if err != nil {
		resp.Diagnostics.AddError("Client Error", fmt.Sprintf("Unable to read "+WebhookSigningKeysResourceName+", got error: %s", err))
		return
	}

	data.Key1 = types.StringPointerValue(webhookSigningKeys.Key1.Get())
	data.Key2 = types.StringPointerValue(webhookSigningKeys.Key2.Get())

	resp.Diagnostics.Append(resp.Result.Set(ctx, &data)...)
}
