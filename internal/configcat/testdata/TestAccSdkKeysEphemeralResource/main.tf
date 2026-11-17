variable "config_id" {
  type = string
}

variable "environment_id" {
  type = string
}

ephemeral "configcat_sdkkeys" "test" {
  config_id      = var.config_id
  environment_id = var.environment_id
}

provider "echo" {
  data = ephemeral.configcat_sdkkeys.test
}

resource "echo" "test" {}
