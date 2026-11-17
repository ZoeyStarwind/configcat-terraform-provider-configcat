package configcat

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/config"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/tfversion"
)

// These tests run the ephemeral resources against a fake ConfigCat Public
// Management API, so they need no ConfigCat credentials or TF_ACC.

const (
	fakeApiUsername = "fake-user"
	fakeApiPassword = "fake-password"

	fakeConfigId          = "11111111-1111-1111-1111-111111111111"
	fakeEnvironment1Id    = "22222222-2222-2222-2222-222222222222" // primary key only
	fakeEnvironment2Id    = "33333333-3333-3333-3333-333333333333" // primary and secondary keys
	fakeEnvironmentNotFnd = "00000000-0000-0000-0000-000000000000"

	fakePrimarySdkKey1   = "fake-sdk-key-1/primary"
	fakePrimarySdkKey2   = "fake-sdk-key-2/primary"
	fakeSecondarySdkKey2 = "fake-sdk-key-2/secondary"

	fakeWebhook1Id = "101" // key1 only
	fakeWebhook2Id = "102" // key1 and key2

	fakeWebhook1Key1 = "configcat_whsk_fake1"
	fakeWebhook2Key1 = "configcat_whsk_fake2a"
	fakeWebhook2Key2 = "configcat_whsk_fake2b"
)

func newFakeConfigCatApi(t *testing.T) *httptest.Server {
	t.Helper()

	writeJson := func(w http.ResponseWriter, body any) {
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(body); err != nil {
			t.Errorf("fake API: unable to encode response: %s", err)
		}
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/me", func(w http.ResponseWriter, r *http.Request) {
		writeJson(w, map[string]any{"email": "fake@example.com", "fullName": "Fake User"})
	})
	mux.HandleFunc("GET /v1/configs/{configId}/environments/{environmentId}", func(w http.ResponseWriter, r *http.Request) {
		if r.PathValue("configId") != fakeConfigId {
			http.NotFound(w, r)
			return
		}
		switch r.PathValue("environmentId") {
		case fakeEnvironment1Id:
			writeJson(w, map[string]any{"primary": fakePrimarySdkKey1, "secondary": nil})
		case fakeEnvironment2Id:
			writeJson(w, map[string]any{"primary": fakePrimarySdkKey2, "secondary": fakeSecondarySdkKey2})
		default:
			http.NotFound(w, r)
		}
	})
	mux.HandleFunc("GET /v1/webhooks/{webhookId}/keys", func(w http.ResponseWriter, r *http.Request) {
		switch r.PathValue("webhookId") {
		case fakeWebhook1Id:
			writeJson(w, map[string]any{"key1": fakeWebhook1Key1, "key2": nil})
		case fakeWebhook2Id:
			writeJson(w, map[string]any{"key1": fakeWebhook2Key1, "key2": fakeWebhook2Key2})
		default:
			http.NotFound(w, r)
		}
	})

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if username, password, ok := r.BasicAuth(); !ok || username != fakeApiUsername || password != fakeApiPassword {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
		mux.ServeHTTP(w, r)
	}))
	t.Cleanup(server.Close)

	t.Setenv(ENV_BASE_PATH, server.URL)
	t.Setenv(ENV_BASIC_AUTH_USERNAME, fakeApiUsername)
	t.Setenv(ENV_BASIC_AUTH_PASSWORD, fakeApiPassword)

	return server
}

func TestSdkKeysEphemeralResourceFakeApi(t *testing.T) {
	const testResourceName = "echo.test"
	newFakeConfigCatApi(t)

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		// The not-found step runs first: the post-test destroy uses the last step's config, which must be readable.
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				ConfigFile:               config.StaticFile("testdata/TestAccSdkKeysEphemeralResource/main.tf"),
				ConfigVariables: config.Variables{
					"config_id":      config.StringVariable(fakeConfigId),
					"environment_id": config.StringVariable(fakeEnvironmentNotFnd),
				},
				ExpectError: regexp.MustCompile(`Unable to read SDK Keys, got error: 404 Not Found`),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				ConfigFile:               config.StaticFile("testdata/TestAccSdkKeysEphemeralResource/main.tf"),
				ConfigVariables: config.Variables{
					"config_id":      config.StringVariable(fakeConfigId),
					"environment_id": config.StringVariable(fakeEnvironment1Id),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "data."+PrimarySdkKey, fakePrimarySdkKey1),
					resource.TestCheckNoResourceAttr(testResourceName, "data."+SecondarySdkKey),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				ConfigFile:               config.StaticFile("testdata/TestAccSdkKeysEphemeralResource/main.tf"),
				ConfigVariables: config.Variables{
					"config_id":      config.StringVariable(fakeConfigId),
					"environment_id": config.StringVariable(fakeEnvironment2Id),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "data."+PrimarySdkKey, fakePrimarySdkKey2),
					resource.TestCheckResourceAttr(testResourceName, "data."+SecondarySdkKey, fakeSecondarySdkKey2),
				),
			},
		},
	})
}

func TestWebhookSigningKeysEphemeralResourceFakeApi(t *testing.T) {
	const testResourceName = "echo.test"
	newFakeConfigCatApi(t)

	resource.UnitTest(t, resource.TestCase{
		TerraformVersionChecks: []tfversion.TerraformVersionCheck{
			tfversion.SkipBelow(tfversion.Version1_10_0),
		},
		// The not-found step runs first: the post-test destroy uses the last step's config, which must be readable.
		Steps: []resource.TestStep{
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				ConfigFile:               config.StaticFile("testdata/TestAccWebhookSigningKeysEphemeralResource/main.tf"),
				ConfigVariables: config.Variables{
					"webhook_id": config.IntegerVariable(-1),
				},
				ExpectError: regexp.MustCompile(`Unable to read Webhook Signing Keys, got error: 404 Not Found`),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				ConfigFile:               config.StaticFile("testdata/TestAccWebhookSigningKeysEphemeralResource/main.tf"),
				ConfigVariables: config.Variables{
					"webhook_id": config.IntegerVariable(101),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "data."+WebhookSigningKeyKey1, fakeWebhook1Key1),
					resource.TestCheckNoResourceAttr(testResourceName, "data."+WebhookSigningKeyKey2),
				),
			},
			{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactoriesWithEcho,
				ConfigFile:               config.StaticFile("testdata/TestAccWebhookSigningKeysEphemeralResource/main.tf"),
				ConfigVariables: config.Variables{
					"webhook_id": config.IntegerVariable(102),
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(testResourceName, "data."+WebhookSigningKeyKey1, fakeWebhook2Key1),
					resource.TestCheckResourceAttr(testResourceName, "data."+WebhookSigningKeyKey2, fakeWebhook2Key2),
				),
			},
		},
	})
}
