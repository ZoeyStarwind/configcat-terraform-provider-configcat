variable "config_id" {
  type = string
}

variable "environment_id" {
  type = string
}

# The SDK Keys are never persisted to the Terraform plan or state.
ephemeral "configcat_sdkkeys" "my_sdkkey" {
  config_id      = var.config_id
  environment_id = var.environment_id
}

# Pass the SDK Key to a write-only argument of another resource, e.g. a secret store.
resource "aws_secretsmanager_secret_version" "configcat_sdkkey" {
  secret_id                = "configcat-sdkkey"
  secret_string_wo         = ephemeral.configcat_sdkkeys.my_sdkkey.primary
  secret_string_wo_version = 1
}
