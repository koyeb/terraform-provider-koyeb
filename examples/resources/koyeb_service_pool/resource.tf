resource "koyeb_service_pool" "my-pool" {
  name = "my-pool"
  size = 2

  definition {
    name = "my-pool"
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
    regions = ["fra"]
    docker {
      image = "koyeb/demo"
    }
  }
}
