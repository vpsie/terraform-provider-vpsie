# Managed databases are imported by their UUID identifier. Note: the offer,
# VPC, project, and datacenter are not returned by the API, so those arguments
# must be set in config (a plan after import will otherwise propose replacement).
terraform import vpsie_managed_database.example "3fa85f64-5717-4562-b3fc-2c963f66afa6"
