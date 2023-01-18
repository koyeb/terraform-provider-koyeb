
terraform {
  required_providers {
    koyeb = {
      source = "koyeb/koyeb"
    }
  }
}
provider "koyeb" {
  #
  # Use the KOYEB_TOKEN env variable to set your Koyeb API token.
  #
}

resource "koyeb_app" "my-app" {
  name = var.app_name
}

resource "koyeb_domain" "my-domain" {
  name     = var.domain_name
  app_name = var.app_name

  depends_on = [
    koyeb_app.my-app
  ]
}

resource "koyeb_secret" "simple" {
  name  = var.secret_simple_name
  value = var.secret_simple_value
}

resource "koyeb_secret" "secret_dockerhub_registry_configuration" {
  name = var.secret_dockerhub_registry_configuration_name
  type = "REGISTRY"
  docker_hub_registry {
    username = var.secret_dockerhub_registry_configuration_username
    password = var.secret_dockerhub_registry_configuration_token
  }
}

resource "koyeb_service" "my-service" {
  app_name = var.app_name
  definition {
    name = var.service_name
    instance_types {
      type = "micro"
    }
    ports {
      port     = 3000
      protocol = "http"
    }
    scalings {
      min = 1
      max = 1
    }
    env {
      key   = "FOO"
      value = "BAR"
    }
    routes {
      path = "/"
      port = 3000
    }
    regions = ["fra"]
    docker {
      image = "koyeb/demo"
    }
  }

  depends_on = [
    koyeb_app.my-app
  ]
}