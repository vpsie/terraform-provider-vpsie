# An SMB user can be imported with its identifier; set password and vpc_id (or share_identifier) in the configuration,
# the platform reports neither. The next apply sets the configured password.
terraform import vpsie_file_smb_user.alice <smb-user-identifier>
