package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/servicediscovery"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/service/servicediscovery/finder"
	"github.com/hashicorp/terraform-provider-aws/aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAWSServiceDiscoveryInstance_private(t *testing.T) {
	resourceName := "aws_service_discovery_instance.instance"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheckAWSServiceDiscovery(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsServiceDiscoveryInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePrivateNamespaceConfig(rName, domainName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"10.0.0.1\" \n    AWS_INSTANCE_IPV6 = \"2001:0db8:85a3:0000:0000:abcd:0001:2345\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "10.0.0.1"),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV6", "2001:0db8:85a3:0000:0000:abcd:0001:2345"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePrivateNamespaceConfig(rName, domainName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"10.0.0.2\" \n    AWS_INSTANCE_IPV6 = \"2001:0db8:85a3:0000:0000:abcd:0001:2345\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "10.0.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV6", "2001:0db8:85a3:0000:0000:abcd:0001:2345"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceDiscoveryInstance_public(t *testing.T) {
	resourceName := "aws_service_discovery_instance.instance"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheckAWSServiceDiscovery(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsServiceDiscoveryInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePublicNamespaceConfig(rName, domainName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"52.18.0.2\" \n    CUSTOM_KEY = \"this is a custom value\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "52.18.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.CUSTOM_KEY", "this is a custom value"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstancePublicNamespaceConfig(rName, domainName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"52.18.0.2\" \n    CUSTOM_KEY = \"this is a custom value updated\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "52.18.0.2"),
					resource.TestCheckResourceAttr(resourceName, "attributes.CUSTOM_KEY", "this is a custom value updated"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccAWSServiceDiscoveryInstance_http(t *testing.T) {
	resourceName := "aws_service_discovery_instance.instance"
	rName := sdkacctest.RandomWithPrefix("tf-acc-test")
	domainName := acctest.RandomDomainName()

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(servicediscovery.EndpointsID, t)
			testAccPreCheckAWSServiceDiscovery(t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, servicediscovery.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAwsServiceDiscoveryInstanceDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstanceHttpNamespaceConfig(rName, domainName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_EC2_INSTANCE_ID = aws_instance.test_instance.id"),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttrSet(resourceName, "attributes.AWS_EC2_INSTANCE_ID"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccAWSServiceDiscoveryInstanceBaseConfig(rName),
					testAccAWSServiceDiscoveryInstanceHttpNamespaceConfig(rName, domainName),
					testAccAWSServiceDiscoveryInstanceConfig(rName, "AWS_INSTANCE_IPV4 = \"172.18.0.12\""),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsServiceDiscoveryInstanceExists(resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "service_id"),
					resource.TestCheckResourceAttr(resourceName, "instance_id", rName),
					resource.TestCheckResourceAttr(resourceName, "attributes.AWS_INSTANCE_IPV4", "172.18.0.12"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAWSServiceDiscoveryInstanceBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "sd_register_instance" {
  cidr_block           = "10.0.0.0/16"
  enable_dns_support   = true
  enable_dns_hostnames = true

  tags = {
    Name = %q
  }
}`, rName)
}

func testAccAWSServiceDiscoveryInstancePrivateNamespaceConfig(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_private_dns_namespace" "sd_register_instance" {
  name        = %[2]q
  description = %[1]q
  vpc         = aws_vpc.sd_register_instance.id
}

resource "aws_service_discovery_service" "sd_register_instance" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_private_dns_namespace.sd_register_instance.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}`, rName, domainName)
}

func testAccAWSServiceDiscoveryInstancePublicNamespaceConfig(rName, domainName string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_public_dns_namespace" "sd_register_instance" {
  name = %[2]q
}

resource "aws_service_discovery_service" "sd_register_instance" {
  name = %[1]q

  dns_config {
    namespace_id = aws_service_discovery_public_dns_namespace.sd_register_instance.id

    dns_records {
      ttl  = 10
      type = "A"
    }

    routing_policy = "MULTIVALUE"
  }

  health_check_custom_config {
    failure_threshold = 1
  }
}`, rName, domainName)
}

func testAccAWSServiceDiscoveryInstanceHttpNamespaceConfig(rName, domainName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_instance" "test_instance" {
  instance_type = "t2.micro"
  ami           = data.aws_ami.amzn-ami-minimal-hvm-ebs.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_service_discovery_http_namespace" "sd_register_instance" {
  name = %[2]q
}

resource "aws_service_discovery_service" "sd_register_instance" {
  name         = %[1]q
  namespace_id = aws_service_discovery_http_namespace.sd_register_instance.id
}`, rName, domainName))
}

func testAccAWSServiceDiscoveryInstanceConfig(instanceID string, attributes string) string {
	return fmt.Sprintf(`
resource "aws_service_discovery_instance" "instance" {
  service_id  = aws_service_discovery_service.sd_register_instance.id
  instance_id = %[1]q

  attributes = {
    %[2]s
  }
}`, instanceID, attributes)
}

func testAccCheckAwsServiceDiscoveryInstanceExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Service Discovery Instance ID is set")
		}

		conn := acctest.Provider.Meta().(*AWSClient).sdconn

		_, err := finder.InstanceByServiceIDAndInstanceID(conn, rs.Primary.Attributes["service_id"], rs.Primary.Attributes["instance_id"])

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccAWSServiceDiscoveryInstanceImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["service_id"], rs.Primary.Attributes["instance_id"]), nil
	}
}

func testAccCheckAwsServiceDiscoveryInstanceDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*AWSClient).sdconn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_service_discovery_instance" {
			continue
		}

		_, err := finder.InstanceByServiceIDAndInstanceID(conn, rs.Primary.Attributes["service_id"], rs.Primary.Attributes["instance_id"])

		if tfresource.NotFound(err) {
			continue
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Service Discovery Instance %s still exists", rs.Primary.ID)
	}

	return nil
}
