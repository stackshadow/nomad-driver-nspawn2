#log_level = "INFO"
log_level = "DEBUG"

ports {
  http = 4656
  rpc  = 4657
  serf = 4658
}

plugin "nspawn2" {
  config {
    sudo = "true"
    nspawn_path = "/run/current-system/systemd/bin/systemd-nspawn"
    ip_path = "/nix/store/49av73h3l9rabx0jrac5hcsf1h3x5y6s-iproute2-6.17.0/bin/ip"
    machinectl_path = "/run/current-system/systemd/bin/machinectl"
  }
}
