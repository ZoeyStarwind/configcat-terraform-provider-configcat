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

# The echo resource is not updated when only ephemeral values change, so
# recreate it whenever the inputs change to read the keys of the new Config and Environment.
resource "terraform_data" "inputs" {
  input = [var.config_id, var.environment_id]
}

resource "echo" "test" {
  lifecycle {
    replace_triggered_by = [terraform_data.inputs]
  }
}
