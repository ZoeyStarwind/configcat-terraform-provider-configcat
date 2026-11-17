variable "webhook_id" {
  type = number
}

ephemeral "configcat_webhook_signing_keys" "test" {
  webhook_id = var.webhook_id
}

provider "echo" {
  data = ephemeral.configcat_webhook_signing_keys.test
}

resource "echo" "test" {}
