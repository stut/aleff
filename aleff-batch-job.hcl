job "aleff" {
  datacenters = ["dc1"]

  periodic {
    # Run hourly.
    cron             = "0 * * * * *"
    # Only one instance of Aleff can run at once.
    prohibit_overlap = true
  }

  group "processor" {
    count = 1

    ephemeral_disk {
      size = 10
    }

    task "processor" {
      driver = "docker"

      config {
        image = "stut/aleff:latest"
        force_pull = true
      }

      env {
        # Use an external timer.
        RUN_INTERVAL = "0"

        # Location of the challenge responder job definition file (see template below).
        CHALLENGE_RESPONDER_JOB_FILENAME = "local/challenge-responder.hcl"

        # Requires access to both Nomad and Consul so set up any URLs, tokens, etc in the environment.
        NOMAD_ADDR = "http://127.0.0.1:4646"
        CONSUL_HTTP_ADDR = "http://127.0.0.1:8500"
      }

      resources {
        cpu    = 8
        memory = 16
      }

      logs {
        max_files     = 1
        max_file_size = 5
      }

      template {
        destination = "local/challenge-responder.hcl"
        data = <<EOH
job "aleff-challenge-responder" {
  datacenters = ["dc1"]

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
        # Requires access to Consul so set up any URLs, tokens, etc in the environment.
        CONSUL_HTTP_ADDR = "http://127.0.0.1:8500"
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
        # The necessary urlprefix- tag will be added by aleff before deploying this service.
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
EOH
      }
    }
  }
}

