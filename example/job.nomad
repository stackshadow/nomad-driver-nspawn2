job "nspawn-example" {

  group "nspawn-group" {
    count = 1

    task "nspawn-task" {
      driver = "nspawn2"

      config {
        name                            = "development-nspawn"
        image                           = "$IMAGE"
        boot                            = true
        ephemeral                       = true
      }

      resources {
        cores  = 4
        memory = 4000
      }
    }
  }
}