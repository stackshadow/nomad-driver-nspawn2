#log_level = "INFO"
log_level = "DEBUG"

ports {
  http = 4656
  rpc  = 4657
  serf = 4658
}

plugin "nspawn2" {
  config { }
}
