package mackerel

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
	"github.com/mackerelio/mackerel-client-go"
)

func TestAccMackerelServiceMetadata(t *testing.T) {
	resourceName := "mackerel_service_metadata.foo"
	rand := acctest.RandString(5)
	serviceName := fmt.Sprintf("tf-%s", rand)
	namespace := fmt.Sprintf("tf-namespace-%s", rand)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckMackerelServiceMetadataDestroy,
		Steps: []resource.TestStep{
			// Test: Create
			{
				Config: testAccMackerelServiceMetadataConfig(serviceName, namespace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMackerelServiceMetadataExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
					resource.TestCheckResourceAttr(resourceName, "namespace", namespace),
					resource.TestCheckResourceAttr(resourceName, "metadata_json", `{"id":1}`),
				),
			},
			// Test: Update
			{
				Config: testAccMackerelServiceMetadataConfigUpdated(serviceName, namespace),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckMackerelServiceMetadataExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "service", serviceName),
					resource.TestCheckResourceAttr(resourceName, "namespace", namespace),
					resource.TestCheckResourceAttr(resourceName, "metadata_json", `{"id":2}`),
				),
			},
			// Test: Import
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Test: Import (invalid format)
			{
				ResourceName:  "mackerel_service_metadata.foo",
				ExpectError:   regexp.MustCompile("the ID must be in the form '<service name>/<namespace>'"),
				ImportState:   true,
				ImportStateId: "invalid_format",
			},
		},
	})
}

func testAccCheckMackerelServiceMetadataDestroy(s *terraform.State) error {
	client := testAccProvider.Meta().(*mackerel.Client)
	for _, r := range s.RootModule().Resources {
		if r.Type != "mackerel_service_metadata" {
			continue
		}

		service := r.Primary.Attributes["service"]
		namespace := r.Primary.Attributes["namespace"]
		if _, err := client.GetServiceMetaData(service, namespace); err == nil {
			return fmt.Errorf("service metadata still exists: %s:%s", service, namespace)
		}
	}
	return nil
}

func testAccCheckMackerelServiceMetadataExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("service_metadata not found from resources: %s", n)
		}
		if rs.Primary.ID == "" {
			return fmt.Errorf("no service_metadata ID is set")
		}

		client := testAccProvider.Meta().(*mackerel.Client)
		_, err := client.GetServiceMetaData(rs.Primary.Attributes["service"], rs.Primary.Attributes["namespace"])
		if err != nil {
			return err
		}
		return nil
	}
}

func testAccMackerelServiceMetadataConfig(serviceName, namespace string) string {
	return fmt.Sprintf(`
resource "mackerel_service" "foo" {
  name = "%s"
}

resource "mackerel_service_metadata" "foo" {
  service = mackerel_service.foo.id
  namespace = "%s"
  metadata_json = jsonencode({
    id = 1
  })
}
`, serviceName, namespace)
}

func testAccMackerelServiceMetadataConfigUpdated(serviceName, namespace string) string {
	return fmt.Sprintf(`
resource "mackerel_service" "foo" {
  name = "%s"
}

resource "mackerel_service_metadata" "foo" {
  service = mackerel_service.foo.id
  namespace = "%s"
  metadata_json = jsonencode({
    id = 2
  })
}
`, serviceName, namespace)
}
