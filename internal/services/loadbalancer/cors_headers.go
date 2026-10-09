package loadbalancer

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vpsie/govpsie"
)

// boolPointer sends a configured flag and leaves an unset (null or unknown) one out of the request.
func boolPointer(value types.Bool) *bool {
	if value.IsNull() || value.IsUnknown() {
		return nil
	}
	v := value.ValueBool()
	return &v
}

// corsHeadersFromAPI is the cors_headers value after a read. Not set in the configuration (null), it
// stays null: the attribute is Optional only, so an API that does not return the setting (an older
// release) can never leave it unknown and force the rules to be replaced. Set, it is read back from
// the API so drift shows in the plan; an API without the field keeps the configured value.
func corsHeadersFromAPI(current types.Bool, api *govpsie.FlexBool) types.Bool {
	if current.IsNull() || current.IsUnknown() {
		return types.BoolNull()
	}
	if v := api.Bool(); v != nil {
		return types.BoolValue(*v)
	}
	return current
}
