package storage_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/vpsie/terraform-provider-vpsie/internal/acctest"
)

func TestAccStorageResource(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.TestAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// create and read testing
			{
				Config: testAccVpsieStorageConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "region", "eu-west-1"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "size", "10"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "description", "test description"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "name", "eu-west-1"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "diskformat", "eu-west-1"),
				),
			},
			{
				ResourceName:      "vpsie_storage.test-storage",
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccVpsieStorageConfig_update("test-update", "20"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "region", "eu-west-1"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "size", "20"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "description", "test description"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "name", "eu-west-1"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "diskformat", "eu-west-1"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "size", "20"),
				),
			},
		},
	})

}

func TestAccStorageResource_Resize(t *testing.T) {
	var storageName = "test-update"

	resource.ParallelTest(t, resource.TestCase{
		ProtoV5ProviderFactories: acctest.TestAccProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// create and read testing
			{
				Config: testAccVpsieStorageConfig_update(storageName, "20"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "dcIdentifier", "eu-west-1"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "description", "test description"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "name", storageName),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "size", "20"),
				),
			},

			{
				Config: testAccVpsieStorageConfig_update(storageName, "50"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "dcIdentifier", "eu-west-1"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "description", "test description"),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "name", storageName),
					resource.TestCheckResourceAttr("vpsie_storage.test-storage", "size", "50"),
				),
			},
		},
	})

}

const testAccVpsieStorageConfig_basic = `
resource "vpsie_storage" "test-storage" {
	region = "eu-west-1"
	name = "test"
	size = "10"
	type = "standard"
	description = "test description"
}
`

func testAccVpsieStorageConfig_update(name, size string) string {
	return fmt.Sprintf(`
		resource "vpsie_storage" "test-storage" {
			dcIdentifier = ""
			name = %s
			size = %s
			disk_format = "EXT4"
			description = "test description"
		}`, name, size)
}
