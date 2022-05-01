job "aleff-challenge-responder" {
  datacenters = ["stutnet"]

  group "responder" {
    count = 1

    network {
      port "http" {}
    }

    ephemeral_disk {
      size = 10
    }

    task "server" {
      driver = "docker"

      config {
        image      = "stut/aleff-challenge-responder:latest"
        force_pull = true
        ports      = ["http"]
      }

      env {
        CONSUL_HTTP_ADDR = "http://192.168.200.1:8500"
        NOMAD_ADDR = "http://192.168.200.1:4646"
      }

      resources {
        cpu    = 8
        memory = 16
      }

      logs {
        max_files     = 1
        max_file_size = 5
      }

      service {
        tags = []
        port = "http"
        check {
          type     = "http"
          port     = "http"
          path     = "/.well-known/acme-challenge/health"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}
