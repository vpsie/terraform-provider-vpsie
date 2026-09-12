# DNS records have no standalone id, so the import id is a composite of
# "domain_identifier/type/name/content".
terraform import vpsie_dns_record.example "3fa85f64-5717-4562-b3fc-2c963f66afa6/A/www/192.168.1.1"
