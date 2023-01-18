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
    env {
      key   = "PORT"
      value = "3000"
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
