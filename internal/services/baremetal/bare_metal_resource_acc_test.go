package baremetal_test

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

// Ordering a bare-metal server bills the account and takes a server out of
// stock, so the acceptance test also needs the datacenter, plan and OS to
// use; it is skipped without them. Destroy releases the server, which needs
// VPSIE_ACCOUNT_PASSWORD.
const (
	envBareMetalDC   = "VPSIE_ACC_BARE_METAL_DC"
	envBareMetalPlan = "VPSIE_ACC_BARE_METAL_PLAN"
	envBareMetalOS   = "VPSIE_ACC_BARE_METAL_OS"
)

func bareMetalAccEnv(t *testing.T) (dc, plan, osIdentifier string) {
	t.Helper()

	dc, plan, osIdentifier = os.Getenv(envBareMetalDC), os.Getenv(envBareMetalPlan), os.Getenv(envBareMetalOS)
	if dc == "" || plan == "" || osIdentifier == "" || os.Getenv("VPSIE_ACCOUNT_PASSWORD") == "" {
		t.Skipf("set %s, %s, %s and VPSIE_ACCOUNT_PASSWORD to order a bare-metal server", envBareMetalDC, envBareMetalPlan, envBareMetalOS)
	}

	return dc, plan, osIdentifier
}

func TestAccBareMetalResource(t *testing.T) {
	if os.Getenv(resource.EnvTfAcc) == "" {
		t.Skipf("acceptance test; set %s=1 to run it", resource.EnvTfAcc)
	}
	dc, plan, osIdentifier := bareMetalAccEnv(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.TestAccPreCheck(t) },
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBareMetalConfig(dc, plan, osIdentifier, "tf-acc-bm.example.com", "Acc-test-pass1"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("vpsie_bare_metal.test", "identifier"),
					resource.TestCheckResourceAttr("vpsie_bare_metal.test", "order_status", "sold"),
					resource.TestCheckResourceAttr("vpsie_bare_metal.test", "os_phase", "running"),
					resource.TestCheckResourceAttrSet("vpsie_bare_metal.test", "ip_address"),
					resource.TestCheckResourceAttrSet("data.vpsie_bare_metal_plans.all", "plans.#"),
					resource.TestCheckResourceAttrSet("data.vpsie_bare_metal_os.all", "operating_systems.#"),
				),
			},
			{
				ResourceName:            "vpsie_bare_metal.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"plan_identifier", "os_identifier", "password", "release_reason", "delivery_hours", "timeouts"},
			},
			{
				Config: testAccBareMetalConfig(dc, plan, osIdentifier, "tf-acc-bm-renamed.example.com", "Acc-test-pass2"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_bare_metal.test", "hostname", "tf-acc-bm-renamed.example.com"),
					resource.TestCheckResourceAttr("vpsie_bare_metal.test", "os_phase", "running"),
				),
			},
		},
	})
}

func testAccBareMetalConfig(dc, plan, osIdentifier, hostname, password string) string {
	return fmt.Sprintf(`
data "vpsie_bare_metal_plans" "all" {
  dc_identifier = %[1]q
}

data "vpsie_bare_metal_os" "all" {
  dc_identifier   = %[1]q
  plan_identifier = %[2]q
}

resource "vpsie_bare_metal" "test" {
  dc_identifier   = %[1]q
  plan_identifier = %[2]q
  os_identifier   = %[3]q
  hostname        = %[4]q
  password        = %[5]q
  release_reason  = "Terraform acceptance test"

  timeouts = {
    create = "90m"
  }
}
`, dc, plan, osIdentifier, hostname, password)
}
