job "nspawn-example" {

  group "nspawn-group" {
    count = 1

    network {
      # mode = "bridge"

      port "http" {
        to = 80
      }
    }


    task "nspawn-task" {
      driver = "nspawn2"

      identity {
        env  = true
        file = true
      }

      template {
        data        = <<EOH
        Guest System
        EOH
        destination = "local/index.html"
      }

      service {
        provider = "nomad"
        address_mode = "driver"
        tags = ["leader", "mysql"]

        name = "avnav"
        port = "http"

        check {
          address_mode = "driver"
          type     = "tcp"
          port     = "http"
          interval = "10s"
          timeout  = "2s"
        }
      }

  
      config {
        name                            = "development-nspawn"
        image                           = "$IMAGE"
        boot                            = true
        ephemeral                       = true
        network_bridge                  = "avnet"
        # network_veth_extra              = "ve-example"
      }

      resources {
        cores  = 1
        memory = 300
      }
    }
  }
}