resource "vpsie_dns_record" "example" {
  domain_identifier = "domain-identifier"
  name              = "www"
  content           = "192.168.1.1"
  type              = "A"
  ttl               = 3600
}
