package azurerm

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAzureRMCdnEndpoint_basic(t *testing.T) {
	resourceName := "azurerm_cdn_endpoint.test"
	ri := acctest.RandInt()
	config := testAccAzureRMCdnEndpoint_basic(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCdnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCdnEndpointExists(resourceName),
				),
			},
		},
	})
}

func TestAccAzureRMCdnEndpoint_disappears(t *testing.T) {
	resourceName := "azurerm_cdn_endpoint.test"
	ri := acctest.RandInt()
	config := testAccAzureRMCdnEndpoint_basic(ri, testLocation())

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCdnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCdnEndpointExists(resourceName),
					testCheckAzureRMCdnEndpointDisappears(resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccAzureRMCdnEndpoint_updateHostHeader(t *testing.T) {
	resourceName := "azurerm_cdn_endpoint.test"
	ri := acctest.RandInt()
	location := testLocation()
	config := testAccAzureRMCdnEndpoint_hostHeader(ri, "www.example.com", location)
	updatedConfig := testAccAzureRMCdnEndpoint_hostHeader(ri, "www.example2.com", location)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCdnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: config,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCdnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "origin_host_header", "www.example.com"),
				),
			},
			{
				Config: updatedConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCdnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "origin_host_header", "www.example2.com"),
				),
			},
		},
	})
}

func TestAccAzureRMCdnEndpoint_withTags(t *testing.T) {
	resourceName := "azurerm_cdn_endpoint.test"
	ri := acctest.RandInt()
	location := testLocation()
	preConfig := testAccAzureRMCdnEndpoint_withTags(ri, location)
	postConfig := testAccAzureRMCdnEndpoint_withTagsUpdate(ri, location)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testCheckAzureRMCdnEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: preConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCdnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "Production"),
					resource.TestCheckResourceAttr(resourceName, "tags.cost_center", "MSFT"),
				),
			},

			{
				Config: postConfig,
				Check: resource.ComposeTestCheckFunc(
					testCheckAzureRMCdnEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.environment", "staging"),
				),
			},
		},
	})
}

func testCheckAzureRMCdnEndpointExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		profileName := rs.Primary.Attributes["profile_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for cdn endpoint: %s", name)
		}

		conn := testAccProvider.Meta().(*ArmClient).cdnEndpointsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		resp, err := conn.Get(ctx, resourceGroup, profileName, name)
		if err != nil {
			return fmt.Errorf("Bad: Get on cdnEndpointsClient: %+v", err)
		}

		if resp.StatusCode == http.StatusNotFound {
			return fmt.Errorf("Bad: CDN Endpoint %q (resource group: %q) does not exist", name, resourceGroup)
		}

		return nil
	}
}

func testCheckAzureRMCdnEndpointDisappears(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		// Ensure we have enough information in state to look up in API
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		name := rs.Primary.Attributes["name"]
		profileName := rs.Primary.Attributes["profile_name"]
		resourceGroup, hasResourceGroup := rs.Primary.Attributes["resource_group_name"]
		if !hasResourceGroup {
			return fmt.Errorf("Bad: no resource group found in state for cdn endpoint: %s", name)
		}

		conn := testAccProvider.Meta().(*ArmClient).cdnEndpointsClient
		ctx := testAccProvider.Meta().(*ArmClient).StopContext

		future, err := conn.Delete(ctx, resourceGroup, profileName, name)
		if err != nil {
			return fmt.Errorf("Bad: Delete on cdnEndpointsClient: %+v", err)
		}

		err = future.WaitForCompletion(ctx, conn.Client)
		if err != nil {
			return fmt.Errorf("Bad: Delete on cdnEndpointsClient: %+v", err)
		}

		return nil
	}
}

func testCheckAzureRMCdnEndpointDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*ArmClient).cdnEndpointsClient
	ctx := testAccProvider.Meta().(*ArmClient).StopContext

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "azurerm_cdn_endpoint" {
			continue
		}

		name := rs.Primary.Attributes["name"]
		resourceGroup := rs.Primary.Attributes["resource_group_name"]
		profileName := rs.Primary.Attributes["profile_name"]

		resp, err := conn.Get(ctx, resourceGroup, profileName, name)
		if err != nil {
			return nil
		}

		if resp.StatusCode != http.StatusNotFound {
			return fmt.Errorf("CDN Endpoint still exists:\n%#v", resp.EndpointProperties)
		}
	}

	return nil
}

func testAccAzureRMCdnEndpoint_basic(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_cdn_profile" "test" {
  name                = "acctestcdnprof%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard_Verizon"
}

resource "azurerm_cdn_endpoint" "test" {
  name                = "acctestcdnend%d"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  origin {
    name       = "acceptanceTestCdnOrigin1"
    host_name  = "www.example.com"
    https_port = 443
    http_port  = 80
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureRMCdnEndpoint_hostHeader(rInt int, domain string, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_cdn_profile" "test" {
  name                = "acctestcdnprof%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard_Verizon"
}

resource "azurerm_cdn_endpoint" "test" {
  name                = "acctestcdnend%d"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  origin_host_header  = "%s"

  origin {
    name       = "acceptanceTestCdnOrigin2"
    host_name  = "www.example.com"
    https_port = 443
    http_port  = 80
  }

  tags {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt, rInt, domain)
}

func testAccAzureRMCdnEndpoint_withTags(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_cdn_profile" "test" {
  name                = "acctestcdnprof%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard_Verizon"
}

resource "azurerm_cdn_endpoint" "test" {
  name                = "acctestcdnend%d"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  origin {
    name       = "acceptanceTestCdnOrigin2"
    host_name  = "www.example.com"
    https_port = 443
    http_port  = 80
  }

  tags {
    environment = "Production"
    cost_center = "MSFT"
  }
}
`, rInt, location, rInt, rInt)
}

func testAccAzureRMCdnEndpoint_withTagsUpdate(rInt int, location string) string {
	return fmt.Sprintf(`
resource "azurerm_resource_group" "test" {
  name     = "acctestRG-%d"
  location = "%s"
}

resource "azurerm_cdn_profile" "test" {
  name                = "acctestcdnprof%d"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"
  sku                 = "Standard_Verizon"
}

resource "azurerm_cdn_endpoint" "test" {
  name                = "acctestcdnend%d"
  profile_name        = "${azurerm_cdn_profile.test.name}"
  location            = "${azurerm_resource_group.test.location}"
  resource_group_name = "${azurerm_resource_group.test.name}"

  origin {
    name       = "acceptanceTestCdnOrigin2"
    host_name  = "www.example.com"
    https_port = 443
    http_port  = 80
  }

  tags {
    environment = "staging"
  }
}
`, rInt, location, rInt, rInt)
}
