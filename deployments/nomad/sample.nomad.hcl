variable "environment" {
  type = string
}

variable "image_ref" {
  type        = string
  description = "Immutable OCI reference including @sha256 digest"
}

variable "datacenters" {
  type    = list(string)
  default = ["dc1"]
}

variable "count" {
  type    = number
  default = 1
}

variable "host" {
  type    = string
  default = ""
}

locals {
  job_name     = "github-actions-sample-${var.environment}"
  service_name = "github-actions-sample-${var.environment}"
  route_tags = var.host == "" ? [] : [
    "traefik.enable=true",
    "traefik.http.routers.${local.service_name}.rule=Host(`${var.host}`)",
    "traefik.http.routers.${local.service_name}.entrypoints=websecure",
    "traefik.http.routers.${local.service_name}.tls=true",
  ]
}

job "github-actions-sample" {
  name        = local.job_name
  datacenters = var.datacenters
  type        = "service"

  group "api" {
    count = var.count

    network {
      port "http" {
        to = 8080
      }
    }

    update {
      max_parallel      = 1
      canary            = 1
      auto_promote      = true
      auto_revert       = true
      min_healthy_time  = "10s"
      healthy_deadline  = "3m"
      progress_deadline = "5m"
    }

    restart {
      attempts = 3
      interval = "5m"
      delay    = "10s"
      mode     = "fail"
    }

    task "migrate" {
      lifecycle {
        hook    = "prestart"
        sidecar = false
      }

      driver = "docker"

      config {
        image   = var.image_ref
        command = "/usr/local/bin/sample"
        args    = ["migrate", "up"]
      }

      template {
        data = <<EOH
DATABASE_URL={{ key "railzway/github-actions-sample/${var.environment}/database_url" }}
MIGRATIONS_PATH=/app/db
EOH
        destination = "secrets/migration.env"
        env         = true
      }

      resources {
        cpu    = 100
        memory = 64
      }
    }

    service {
      name = local.service_name
      port = "http"
      tags = concat(["environment=${var.environment}"], local.route_tags)

      check {
        name     = "HTTP health"
        type     = "http"
        path     = "/health"
        interval = "10s"
        timeout  = "2s"
      }

      check {
        name     = "Database readiness"
        type     = "http"
        path     = "/ready"
        interval = "10s"
        timeout  = "3s"
      }
    }

    task "server" {
      driver = "docker"

      config {
        image = var.image_ref
        ports = ["http"]
      }

      env {
        PORT        = "8080"
        APP_VERSION = var.image_ref
      }

      template {
        data = <<EOH
DATABASE_URL={{ key "railzway/github-actions-sample/${var.environment}/database_url" }}
EOH
        destination = "secrets/database.env"
        env         = true
      }

      resources {
        cpu    = 200
        memory = 128
      }
    }
  }
}
