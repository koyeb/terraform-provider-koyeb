variable "app_name" {
  description = "Koyeb app name"
  default     = "demo-app-01"
}

variable "domain_name" {
  description = "Koyeb domain name"
  default     = "www.koyeb.com"
}

variable "service_name" {
  description = "Koyeb service name"
  default     = "my-service"
}

variable "secret_dockerhub_registry_configuration_name" {
  description = "Koyeb secret name to store the DockerHub registry configuration"
  default     = "my-dockerhub-registry-configuration"
}

variable "secret_dockerhub_registry_configuration_username" {
  description = "The usage to connect the DockerHub registry"
  default     = "my-dockerhub-registry-configuration-token"
}

variable "secret_dockerhub_registry_configuration_token" {
  description = "The token to connect DockerHub registry"
  default     = "my-dockerhub-registry-configuration-token"
}

variable "secret_simple_name" {
  description = "Koyeb secret name"
  default     = "my-secret-name"
}

variable "secret_simple_value" {
  description = "Koyeb secret value"
  default     = "my-value"
}


