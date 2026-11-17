variable "webhook_id" {
  type = number
}

ephemeral "configcat_webhook_signing_keys" "test" {
  webhook_id = var.webhook_id
}

provider "echo" {
  data = ephemeral.configcat_webhook_signing_keys.test
}

# The echo resource is not updated when only ephemeral values change, so
# recreate it whenever the inputs change to read the keys of the new Webhook.
resource "terraform_data" "inputs" {
  input = var.webhook_id
}

resource "echo" "test" {
  lifecycle {
    replace_triggered_by = [terraform_data.inputs]
  }
}
