variable "webhook_id" {
  type = number
}

# The Webhook Signing Keys are never persisted to the Terraform plan or state.
ephemeral "configcat_webhook_signing_keys" "test" {
  webhook_id = var.webhook_id
}

# Pass the signing key to a write-only argument of another resource, e.g. a secret store.
resource "aws_secretsmanager_secret_version" "configcat_webhook_signing_key" {
  secret_id                = "configcat-webhook-signing-key"
  secret_string_wo         = ephemeral.configcat_webhook_signing_keys.test.key1
  secret_string_wo_version = 1
}
