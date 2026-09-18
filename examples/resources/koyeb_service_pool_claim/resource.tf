resource "koyeb_service_pool" "my-pool" {
  name = "my-pool"
  size = 1

  definition {
    name = "pool"
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
    regions = ["tyo"]
    docker {
      image = "koyeb/demo"
    }
  }
}

# Claim a prewarmed service from the pool; the request ID makes the
# claim idempotent (the same ID re-claims the same service).
resource "koyeb_service_pool_claim" "example" {
  pool       = koyeb_service_pool.my-pool.name
  request_id = "ci-run-42"
}
