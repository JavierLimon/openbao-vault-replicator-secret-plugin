listener "tcp" {
  address = "[::]:8200"
  tls_disable = true
}

storage "file" {
  path = "/vault/data"
}

plugin_directory = "/vault/plugins"

disable_mlock = true