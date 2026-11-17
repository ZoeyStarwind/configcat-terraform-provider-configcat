package configcat

import (
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

func TestAccWebhookSigningKeysEphemeralResource(t *testing.T) {
	const testResourceName = "echo.test"

	const webhook1Id = 6813
	const webhook2Id = 6814

	resource.Test(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
		Steps: []resource.TestStep{
			{
				ConfigFile: config.TestNameFile("main.tf"),
				ConfigVariables: config.Variables{
					"webhook_id": config.IntegerVariable(webhook1Id),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "data."+WebhookSigningKeyKey1, "configcat_whsk_Tfw6XjKPN4eE5aPAl6cTcg0fDvZyMp91cIsMFTyCFZA="),
					resource.TestCheckNoResourceAttr(testResourceName, "data."+WebhookSigningKeyKey2),
				),
			},
			{
				ConfigFile: config.TestNameFile("main.tf"),
				ConfigVariables: config.Variables{
					"webhook_id": config.IntegerVariable(webhook2Id),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "data."+WebhookSigningKeyKey1, "configcat_whsk_03svAq73DMZmYAHH5wErhNdOTXkKMnu5w6HQdqRetUU="),
					resource.TestCheckResourceAttr(testResourceName, "data."+WebhookSigningKeyKey2, "configcat_whsk_Pfhcf/IiTTG8CraCDw7RmyxdUiZXyvKfvFQmh69Pe/4="),
				),
			},
			{
				ConfigFile: config.TestNameFile("main.tf"),
				ConfigVariables: config.Variables{
					"webhook_id": config.IntegerVariable(-1),
				},
				ExpectError: regexp.MustCompile(`Unable to read Webhook Signing Keys, got error: 404 Not Found`),
			},
		},
	})
}
