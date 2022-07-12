resource "koyeb_service" "my-service" {
  app_name = koyeb_app.my_app.name
  definition {
    name = "my-service"
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
    regions = ["par"]
    docker {
      image = "koyeb/demo"
    }
  }

  depends_on = [
    koyeb_app.my-app
  ]
}