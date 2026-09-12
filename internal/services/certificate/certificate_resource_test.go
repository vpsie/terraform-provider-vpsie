package certificate_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccCertificateResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.TestAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCertificateConfig,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_certificate.test", "cert_name", "terraform-acc-cert"),
					resource.TestCheckResourceAttrSet("vpsie_certificate.test", "identifier"),
				),
			},
		},
	})
}

const testAccCertificateConfig = `
resource "vpsie_certificate" "test" {
  cert_name = "terraform-acc-cert"
  domain_id = "replace-with-domain-identifier"
}
`
