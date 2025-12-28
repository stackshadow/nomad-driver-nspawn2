job "nspawn-example" {

  group "nspawn-group" {
    count = 1

    task "nspawn-task" {
      driver = "nspawn2"

      identity {
        env  = true
        file = true
      }
  
      config {
        name                            = "development-nspawn"
        image                           = "$IMAGE"
        boot                            = true
        ephemeral                       = true
        commands                        = [ "/bin/bash", "-c", "echo 'testdata' > /tmp/test" ]
      }

      resources {
        cores  = 1
        memory = 300
      }
    }
  }
}