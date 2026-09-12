resource "vpsie_firewall" "example" {
  group_name = "web-servers"

  rule {
    action  = "ACCEPT"
    type    = "in"
    proto   = "tcp"
    macro   = "SSH" # a macro clears dport/sport
    comment = "allow SSH"
    source  = ["10.0.0.0/8"]
  }

  rule {
    action  = "ACCEPT"
    type    = "in"
    proto   = "tcp"
    dport   = "443"
    comment = "allow HTTPS"
  }

  rule {
    action  = "DROP"
    type    = "out"
    proto   = "udp"
    dport   = "53"
    comment = "block outbound DNS"
  }
}
