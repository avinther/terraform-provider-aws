package ec2

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/verify"
)

func ResourceFleet() *schema.Resource {
	return &schema.Resource{
		Create: resourceFleetCreate,
		Read:   resourceFleetRead,
		Update: resourceFleetUpdate,
		Delete: resourceFleetDelete,
		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		CustomizeDiff: verify.SetTagsDiff,
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(10 * time.Minute),
			Delete: schema.DefaultTimeout(10 * time.Minute),
			Update: schema.DefaultTimeout(10 * time.Minute),
		},

		Schema: map[string]*schema.Schema{
			"context": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"excess_capacity_termination_policy": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  ec2.FleetExcessCapacityTerminationPolicyTermination,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.FleetExcessCapacityTerminationPolicyNoTermination,
					ec2.FleetExcessCapacityTerminationPolicyTermination,
				}, false),
			},
			"launch_template_config": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"launch_template_specification": {
							Type:     schema.TypeList,
							Required: true,
							MinItems: 1,
							MaxItems: 1,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"launch_template_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"launch_template_name": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"version": {
										Type:     schema.TypeString,
										Required: true,
									},
								},
							},
						},
						"override": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 50,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"availability_zone": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"instance_type": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"max_price": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"priority": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
									"subnet_id": {
										Type:     schema.TypeString,
										Optional: true,
									},
									"weighted_capacity": {
										Type:     schema.TypeFloat,
										Optional: true,
									},
								},
							},
						},
					},
				},
			},
			"on_demand_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "lowestPrice",
							ValidateFunc: validation.StringInSlice([]string{
								// AWS SDK constant incorrect
								// ec2.FleetOnDemandAllocationStrategyLowestPrice,
								"lowestPrice",
								ec2.FleetOnDemandAllocationStrategyPrioritized,
							}, false),
						},
					},
				},
			},
			"replace_unhealthy_instances": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"spot_options": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
					if old == "1" && new == "0" {
						return true
					}
					return false
				},
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"allocation_strategy": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  "lowestPrice",
							ValidateFunc: validation.StringInSlice([]string{
								ec2.SpotAllocationStrategyDiversified,
								// AWS SDK constant incorrect
								// ec2.SpotAllocationStrategyLowestPrice,
								"lowestPrice",
							}, false),
						},
						"instance_interruption_behavior": {
							Type:     schema.TypeString,
							Optional: true,
							ForceNew: true,
							Default:  ec2.SpotInstanceInterruptionBehaviorTerminate,
							ValidateFunc: validation.StringInSlice([]string{
								ec2.SpotInstanceInterruptionBehaviorHibernate,
								ec2.SpotInstanceInterruptionBehaviorStop,
								ec2.SpotInstanceInterruptionBehaviorTerminate,
							}, false),
						},
						"instance_pools_to_use_count": {
							Type:         schema.TypeInt,
							Optional:     true,
							ForceNew:     true,
							Default:      1,
							ValidateFunc: validation.IntAtLeast(1),
						},
						"maintenance_strategies": {
							Type:     schema.TypeList,
							Optional: true,
							MaxItems: 1,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								if old == "1" && new == "0" {
									return true
								}
								return false
							},
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"capacity_rebalance": {
										Type:     schema.TypeList,
										Optional: true,
										MaxItems: 1,
										DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
											if old == "1" && new == "0" {
												return true
											}
											return false
										},
										Elem: &schema.Resource{
											Schema: map[string]*schema.Schema{
												"replacement_strategy": {
													Type:     schema.TypeString,
													Optional: true,
													ForceNew: true,
													ValidateFunc: validation.StringInSlice([]string{
														"launch",
													}, false),
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			"tags":     tftags.TagsSchema(),
			"tags_all": tftags.TagsSchemaComputed(),
			"target_capacity_specification": {
				Type:     schema.TypeList,
				Required: true,
				MinItems: 1,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"default_target_capacity_type": {
							Type:     schema.TypeString,
							Required: true,
							ForceNew: true,
							ValidateFunc: validation.StringInSlice([]string{
								ec2.DefaultTargetCapacityTypeOnDemand,
								ec2.DefaultTargetCapacityTypeSpot,
							}, false),
						},
						"on_demand_target_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Show difference for new resources
								if d.Id() == "" {
									return false
								}
								// Show difference if value is configured
								if new != "0" {
									return false
								}
								// Show difference if existing state reflects different default type
								defaultTargetCapacityTypeO, _ := d.GetChange("target_capacity_specification.0.default_target_capacity_type")
								if defaultTargetCapacityTypeO.(string) != ec2.DefaultTargetCapacityTypeOnDemand {
									return false
								}
								// Show difference if existing state reflects different total capacity
								oldInt, err := strconv.Atoi(old)
								if err != nil {
									log.Printf("[WARN] %s DiffSuppressFunc error converting %s to integer: %s", k, old, err)
									return false
								}
								totalTargetCapacityO, _ := d.GetChange("target_capacity_specification.0.total_target_capacity")
								return oldInt == totalTargetCapacityO.(int)
							},
						},
						"spot_target_capacity": {
							Type:     schema.TypeInt,
							Optional: true,
							ForceNew: true,
							DiffSuppressFunc: func(k, old, new string, d *schema.ResourceData) bool {
								// Show difference for new resources
								if d.Id() == "" {
									return false
								}
								// Show difference if value is configured
								if new != "0" {
									return false
								}
								// Show difference if existing state reflects different default type
								defaultTargetCapacityTypeO, _ := d.GetChange("target_capacity_specification.0.default_target_capacity_type")
								if defaultTargetCapacityTypeO.(string) != ec2.DefaultTargetCapacityTypeSpot {
									return false
								}
								// Show difference if existing state reflects different total capacity
								oldInt, err := strconv.Atoi(old)
								if err != nil {
									log.Printf("[WARN] %s DiffSuppressFunc error converting %s to integer: %s", k, old, err)
									return false
								}
								totalTargetCapacityO, _ := d.GetChange("target_capacity_specification.0.total_target_capacity")
								return oldInt == totalTargetCapacityO.(int)
							},
						},
						"total_target_capacity": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"terminate_instances": {
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
			"terminate_instances_with_expiration": {
				Type:     schema.TypeBool,
				Optional: true,
				ForceNew: true,
			},
			"type": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Default:  ec2.FleetTypeMaintain,
				ValidateFunc: validation.StringInSlice([]string{
					ec2.FleetTypeMaintain,
					ec2.FleetTypeRequest,
				}, false),
			},
		},
	}
}

func resourceFleetCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	tags := defaultTagsConfig.MergeTags(tftags.New(d.Get("tags").(map[string]interface{})))

	input := &ec2.CreateFleetInput{
		ExcessCapacityTerminationPolicy:  aws.String(d.Get("excess_capacity_termination_policy").(string)),
		LaunchTemplateConfigs:            expandEc2FleetLaunchTemplateConfigRequests(d.Get("launch_template_config").([]interface{})),
		OnDemandOptions:                  expandEc2OnDemandOptionsRequest(d.Get("on_demand_options").([]interface{})),
		ReplaceUnhealthyInstances:        aws.Bool(d.Get("replace_unhealthy_instances").(bool)),
		SpotOptions:                      expandEc2SpotOptionsRequest(d.Get("spot_options").([]interface{})),
		TagSpecifications:                ec2TagSpecificationsFromKeyValueTags(tags, ec2.ResourceTypeFleet),
		TargetCapacitySpecification:      expandEc2TargetCapacitySpecificationRequest(d.Get("target_capacity_specification").([]interface{})),
		TerminateInstancesWithExpiration: aws.Bool(d.Get("terminate_instances_with_expiration").(bool)),
		Type:                             aws.String(d.Get("type").(string)),
	}

	if d.Get("type").(string) != ec2.FleetTypeMaintain {
		if input.SpotOptions.MaintenanceStrategies != nil {
			log.Printf("[WARN] EC2 Fleet (%s) has an invalid configuration and can not be created. Capacity Rebalance maintenance strategies can only be specified for fleets of type maintain.", input)
			return nil
		}
	}

	if v, ok := d.GetOk("context"); ok {
		input.Context = aws.String(v.(string))
	}

	log.Printf("[DEBUG] Creating EC2 Fleet: %s", input)
	output, err := conn.CreateFleet(input)

	if err != nil {
		return fmt.Errorf("error creating EC2 Fleet: %w", err)
	}

	d.SetId(aws.StringValue(output.FleetId))

	// If a request type is fulfilled immediately, we can miss the transition from active to deleted
	// Instead of an error here, allow the Read function to trigger recreation
	target := []string{ec2.FleetStateCodeActive}
	if d.Get("type").(string) == ec2.FleetTypeRequest {
		target = append(target, ec2.FleetStateCodeDeleted)
		target = append(target, ec2.FleetStateCodeDeletedRunning)
		target = append(target, ec2.FleetStateCodeDeletedTerminating)
		// AWS SDK constants incorrect
		target = append(target, "deleted_running")
		target = append(target, "deleted_terminating")
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.FleetStateCodeSubmitted},
		Target:  target,
		Refresh: ec2FleetRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutCreate),
	}

	log.Printf("[DEBUG] Waiting for EC2 Fleet (%s) activation", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for EC2 Fleet (%s) activation: %s", d.Id(), err)
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn
	defaultTagsConfig := meta.(*conns.AWSClient).DefaultTagsConfig
	ignoreTagsConfig := meta.(*conns.AWSClient).IgnoreTagsConfig

	input := &ec2.DescribeFleetsInput{
		FleetIds: []*string{aws.String(d.Id())},
	}

	log.Printf("[DEBUG] Reading EC2 Fleet (%s): %s", d.Id(), input)
	output, err := conn.DescribeFleets(input)

	if tfawserr.ErrMessageContains(err, "InvalidFleetId.NotFound", "") {
		log.Printf("[WARN] EC2 Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	if err != nil {
		return fmt.Errorf("error reading EC2 Fleet: %s", err)
	}

	if output == nil || len(output.Fleets) == 0 {
		log.Printf("[WARN] EC2 Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	var fleet *ec2.FleetData
	for _, fleetData := range output.Fleets {
		if fleetData == nil {
			continue
		}
		if aws.StringValue(fleetData.FleetId) != d.Id() {
			continue
		}
		fleet = fleetData
		break
	}

	if fleet == nil {
		log.Printf("[WARN] EC2 Fleet (%s) not found, removing from state", d.Id())
		d.SetId("")
		return nil
	}

	deletedStates := []string{
		ec2.FleetStateCodeDeleted,
		ec2.FleetStateCodeDeletedRunning,
		ec2.FleetStateCodeDeletedTerminating,
		// AWS SDK constants are incorrect
		"deleted_running",
		"deleted_terminating",
	}
	for _, deletedState := range deletedStates {
		if aws.StringValue(fleet.FleetState) == deletedState {
			log.Printf("[WARN] EC2 Fleet (%s) in deleted state (%s), removing from state", d.Id(), aws.StringValue(fleet.FleetState))
			d.SetId("")
			return nil
		}
	}

	d.Set("context", fleet.Context)
	d.Set("excess_capacity_termination_policy", fleet.ExcessCapacityTerminationPolicy)

	if err := d.Set("launch_template_config", flattenEc2FleetLaunchTemplateConfigs(fleet.LaunchTemplateConfigs)); err != nil {
		return fmt.Errorf("error setting launch_template_config: %s", err)
	}

	if err := d.Set("on_demand_options", flattenEc2OnDemandOptions(fleet.OnDemandOptions)); err != nil {
		return fmt.Errorf("error setting on_demand_options: %s", err)
	}

	d.Set("replace_unhealthy_instances", fleet.ReplaceUnhealthyInstances)

	if err := d.Set("spot_options", flattenEc2SpotOptions(fleet.SpotOptions)); err != nil {
		return fmt.Errorf("error setting spot_options: %s", err)
	}

	if err := d.Set("target_capacity_specification", flattenEc2TargetCapacitySpecification(fleet.TargetCapacitySpecification)); err != nil {
		return fmt.Errorf("error setting target_capacity_specification: %s", err)
	}

	d.Set("terminate_instances_with_expiration", fleet.TerminateInstancesWithExpiration)
	d.Set("type", fleet.Type)

	tags := KeyValueTags(fleet.Tags).IgnoreAWS().IgnoreConfig(ignoreTagsConfig)

	//lintignore:AWSR002
	if err := d.Set("tags", tags.RemoveDefaultConfig(defaultTagsConfig).Map()); err != nil {
		return fmt.Errorf("error setting tags: %w", err)
	}

	if err := d.Set("tags_all", tags.Map()); err != nil {
		return fmt.Errorf("error setting tags_all: %w", err)
	}

	return nil
}

func resourceFleetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.ModifyFleetInput{
		Context:                         aws.String(d.Get("context").(string)),
		ExcessCapacityTerminationPolicy: aws.String(d.Get("excess_capacity_termination_policy").(string)),
		LaunchTemplateConfigs:           expandEc2FleetLaunchTemplateConfigRequests(d.Get("launch_template_config").([]interface{})),
		FleetId:                         aws.String(d.Id()),
		// InvalidTargetCapacitySpecification: Currently we only support total target capacity modification.
		// TargetCapacitySpecification: expandEc2TargetCapacitySpecificationRequest(d.Get("target_capacity_specification").([]interface{})),
		TargetCapacitySpecification: &ec2.TargetCapacitySpecificationRequest{
			TotalTargetCapacity: aws.Int64(int64(d.Get("target_capacity_specification.0.total_target_capacity").(int))),
		},
	}

	log.Printf("[DEBUG] Modifying EC2 Fleet (%s): %s", d.Id(), input)
	_, err := conn.ModifyFleet(input)

	if err != nil {
		return fmt.Errorf("error modifying EC2 Fleet: %s", err)
	}

	stateConf := &resource.StateChangeConf{
		Pending: []string{ec2.FleetStateCodeModifying},
		Target:  []string{ec2.FleetStateCodeActive},
		Refresh: ec2FleetRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutUpdate),
	}

	log.Printf("[DEBUG] Waiting for EC2 Fleet (%s) modification", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for EC2 Fleet (%s) modification: %s", d.Id(), err)
	}

	if d.HasChange("tags_all") {
		o, n := d.GetChange("tags_all")

		if err := UpdateTags(conn, d.Id(), o, n); err != nil {
			return fmt.Errorf("error updating tags: %s", err)
		}
	}

	return resourceFleetRead(d, meta)
}

func resourceFleetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*conns.AWSClient).EC2Conn

	input := &ec2.DeleteFleetsInput{
		FleetIds:           []*string{aws.String(d.Id())},
		TerminateInstances: aws.Bool(d.Get("terminate_instances").(bool)),
	}

	log.Printf("[DEBUG] Deleting EC2 Fleet (%s): %s", d.Id(), input)
	_, err := conn.DeleteFleets(input)

	if tfawserr.ErrMessageContains(err, "InvalidFleetId.NotFound", "") {
		return nil
	}

	if err != nil {
		return fmt.Errorf("error deleting EC2 Fleet: %s", err)
	}

	pending := []string{ec2.FleetStateCodeActive}
	target := []string{ec2.FleetStateCodeDeleted}
	if d.Get("terminate_instances").(bool) {
		pending = append(pending, ec2.FleetStateCodeDeletedTerminating)
		// AWS SDK constant is incorrect: unexpected state 'deleted_terminating', wanted target 'deleted, deleted-terminating'
		pending = append(pending, "deleted_terminating")
	} else {
		target = append(target, ec2.FleetStateCodeDeletedRunning)
		// AWS SDK constant is incorrect: unexpected state 'deleted_running', wanted target 'deleted, deleted-running'
		target = append(target, "deleted_running")
	}

	stateConf := &resource.StateChangeConf{
		Pending: pending,
		Target:  target,
		Refresh: ec2FleetRefreshFunc(conn, d.Id()),
		Timeout: d.Timeout(schema.TimeoutDelete),
	}

	log.Printf("[DEBUG] Waiting for EC2 Fleet (%s) deletion", d.Id())
	_, err = stateConf.WaitForState()
	if err != nil {
		return fmt.Errorf("error waiting for EC2 Fleet (%s) deletion: %s", d.Id(), err)
	}

	return nil
}

func ec2FleetRefreshFunc(conn *ec2.EC2, fleetID string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		input := &ec2.DescribeFleetsInput{
			FleetIds: []*string{aws.String(fleetID)},
		}

		log.Printf("[DEBUG] Reading EC2 Fleet (%s): %s", fleetID, input)
		output, err := conn.DescribeFleets(input)

		if tfawserr.ErrMessageContains(err, "InvalidFleetId.NotFound", "") {
			return nil, ec2.FleetStateCodeDeleted, nil
		}

		if err != nil {
			return nil, "", fmt.Errorf("error reading EC2 Fleet: %s", err)
		}

		if output == nil || len(output.Fleets) == 0 {
			return nil, ec2.FleetStateCodeDeleted, nil
		}

		var fleet *ec2.FleetData
		for _, fleetData := range output.Fleets {
			if fleetData == nil {
				continue
			}
			if aws.StringValue(fleetData.FleetId) == fleetID {
				fleet = fleetData
				break
			}
		}

		if fleet == nil {
			return nil, ec2.FleetStateCodeDeleted, nil
		}

		return fleet, aws.StringValue(fleet.FleetState), nil
	}
}

func expandEc2FleetLaunchTemplateConfigRequests(l []interface{}) []*ec2.FleetLaunchTemplateConfigRequest {
	fleetLaunchTemplateConfigRequests := make([]*ec2.FleetLaunchTemplateConfigRequest, len(l))
	for i, m := range l {
		if m == nil {
			fleetLaunchTemplateConfigRequests[i] = &ec2.FleetLaunchTemplateConfigRequest{}
			continue
		}

		fleetLaunchTemplateConfigRequests[i] = expandEc2FleetLaunchTemplateConfigRequest(m.(map[string]interface{}))
	}
	return fleetLaunchTemplateConfigRequests
}

func expandEc2FleetLaunchTemplateConfigRequest(m map[string]interface{}) *ec2.FleetLaunchTemplateConfigRequest {
	fleetLaunchTemplateConfigRequest := &ec2.FleetLaunchTemplateConfigRequest{
		LaunchTemplateSpecification: expandEc2LaunchTemplateSpecificationRequest(m["launch_template_specification"].([]interface{})),
	}

	if v, ok := m["override"]; ok {
		fleetLaunchTemplateConfigRequest.Overrides = expandEc2FleetLaunchTemplateOverridesRequests(v.([]interface{}))
	}

	return fleetLaunchTemplateConfigRequest
}

func expandEc2FleetLaunchTemplateOverridesRequests(l []interface{}) []*ec2.FleetLaunchTemplateOverridesRequest {
	if len(l) == 0 {
		return nil
	}

	fleetLaunchTemplateOverridesRequests := make([]*ec2.FleetLaunchTemplateOverridesRequest, len(l))
	for i, m := range l {
		if m == nil {
			fleetLaunchTemplateOverridesRequests[i] = &ec2.FleetLaunchTemplateOverridesRequest{}
			continue
		}

		fleetLaunchTemplateOverridesRequests[i] = expandEc2FleetLaunchTemplateOverridesRequest(m.(map[string]interface{}))
	}
	return fleetLaunchTemplateOverridesRequests
}

func expandEc2FleetLaunchTemplateOverridesRequest(m map[string]interface{}) *ec2.FleetLaunchTemplateOverridesRequest {
	fleetLaunchTemplateOverridesRequest := &ec2.FleetLaunchTemplateOverridesRequest{}

	if v, ok := m["availability_zone"]; ok && v.(string) != "" {
		fleetLaunchTemplateOverridesRequest.AvailabilityZone = aws.String(v.(string))
	}

	if v, ok := m["instance_type"]; ok && v.(string) != "" {
		fleetLaunchTemplateOverridesRequest.InstanceType = aws.String(v.(string))
	}

	if v, ok := m["max_price"]; ok && v.(string) != "" {
		fleetLaunchTemplateOverridesRequest.MaxPrice = aws.String(v.(string))
	}

	if v, ok := m["priority"]; ok && v.(float64) != 0.0 {
		fleetLaunchTemplateOverridesRequest.Priority = aws.Float64(v.(float64))
	}

	if v, ok := m["subnet_id"]; ok && v.(string) != "" {
		fleetLaunchTemplateOverridesRequest.SubnetId = aws.String(v.(string))
	}

	if v, ok := m["weighted_capacity"]; ok && v.(float64) != 0.0 {
		fleetLaunchTemplateOverridesRequest.WeightedCapacity = aws.Float64(v.(float64))
	}

	return fleetLaunchTemplateOverridesRequest
}

func expandEc2LaunchTemplateSpecificationRequest(l []interface{}) *ec2.FleetLaunchTemplateSpecificationRequest {
	fleetLaunchTemplateSpecificationRequest := &ec2.FleetLaunchTemplateSpecificationRequest{}

	if len(l) == 0 || l[0] == nil {
		return fleetLaunchTemplateSpecificationRequest
	}

	m := l[0].(map[string]interface{})

	if v, ok := m["launch_template_id"]; ok && v.(string) != "" {
		fleetLaunchTemplateSpecificationRequest.LaunchTemplateId = aws.String(v.(string))
	}

	if v, ok := m["launch_template_name"]; ok && v.(string) != "" {
		fleetLaunchTemplateSpecificationRequest.LaunchTemplateName = aws.String(v.(string))
	}

	if v, ok := m["version"]; ok && v.(string) != "" {
		fleetLaunchTemplateSpecificationRequest.Version = aws.String(v.(string))
	}

	return fleetLaunchTemplateSpecificationRequest
}

func expandEc2OnDemandOptionsRequest(l []interface{}) *ec2.OnDemandOptionsRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	return &ec2.OnDemandOptionsRequest{
		AllocationStrategy: aws.String(m["allocation_strategy"].(string)),
	}
}

func expandEc2SpotOptionsRequest(l []interface{}) *ec2.SpotOptionsRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	spotOptionsRequest := &ec2.SpotOptionsRequest{
		AllocationStrategy:           aws.String(m["allocation_strategy"].(string)),
		InstanceInterruptionBehavior: aws.String(m["instance_interruption_behavior"].(string)),
		MaintenanceStrategies:        expandFleetSpotMaintenanceStrategiesRequest(m["maintenance_strategies"].([]interface{})),
	}

	// InvalidFleetConfig: InstancePoolsToUseCount option is only available with the lowestPrice allocation strategy.
	if aws.StringValue(spotOptionsRequest.AllocationStrategy) == "lowestPrice" {
		spotOptionsRequest.InstancePoolsToUseCount = aws.Int64(int64(m["instance_pools_to_use_count"].(int)))
	}

	return spotOptionsRequest
}

func expandFleetSpotMaintenanceStrategiesRequest(l []interface{}) *ec2.FleetSpotMaintenanceStrategiesRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	fleetSpotMaintenanceStrategiesRequest := &ec2.FleetSpotMaintenanceStrategiesRequest{
		CapacityRebalance: expandFleetSpotCapacityRebalanceRequest(m["capacity_rebalance"].([]interface{})),
	}

	return fleetSpotMaintenanceStrategiesRequest
}

func expandFleetSpotCapacityRebalanceRequest(l []interface{}) *ec2.FleetSpotCapacityRebalanceRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	capacityRebalance := &ec2.FleetSpotCapacityRebalanceRequest{}

	if v, ok := m["replacement_strategy"]; ok && v.(string) != "" {
		capacityRebalance.ReplacementStrategy = aws.String(v.(string))
	}

	return capacityRebalance
}

func expandEc2TargetCapacitySpecificationRequest(l []interface{}) *ec2.TargetCapacitySpecificationRequest {
	if len(l) == 0 || l[0] == nil {
		return nil
	}

	m := l[0].(map[string]interface{})

	targetCapacitySpecificationrequest := &ec2.TargetCapacitySpecificationRequest{
		TotalTargetCapacity: aws.Int64(int64(m["total_target_capacity"].(int))),
	}

	if v, ok := m["default_target_capacity_type"]; ok && v.(string) != "" {
		targetCapacitySpecificationrequest.DefaultTargetCapacityType = aws.String(v.(string))
	}

	if v, ok := m["on_demand_target_capacity"]; ok && v.(int) != 0 {
		targetCapacitySpecificationrequest.OnDemandTargetCapacity = aws.Int64(int64(v.(int)))
	}

	if v, ok := m["spot_target_capacity"]; ok && v.(int) != 0 {
		targetCapacitySpecificationrequest.SpotTargetCapacity = aws.Int64(int64(v.(int)))
	}

	return targetCapacitySpecificationrequest
}

func flattenEc2FleetLaunchTemplateConfigs(fleetLaunchTemplateConfigs []*ec2.FleetLaunchTemplateConfig) []interface{} {
	l := make([]interface{}, len(fleetLaunchTemplateConfigs))

	for i, fleetLaunchTemplateConfig := range fleetLaunchTemplateConfigs {
		if fleetLaunchTemplateConfig == nil {
			l[i] = map[string]interface{}{}
			continue
		}
		m := map[string]interface{}{
			"launch_template_specification": flattenEc2FleetLaunchTemplateSpecification(fleetLaunchTemplateConfig.LaunchTemplateSpecification),
			"override":                      flattenEc2FleetLaunchTemplateOverrides(fleetLaunchTemplateConfig.Overrides),
		}
		l[i] = m
	}

	return l
}

func flattenEc2FleetLaunchTemplateOverrides(fleetLaunchTemplateOverrides []*ec2.FleetLaunchTemplateOverrides) []interface{} {
	l := make([]interface{}, len(fleetLaunchTemplateOverrides))

	for i, fleetLaunchTemplateOverride := range fleetLaunchTemplateOverrides {
		if fleetLaunchTemplateOverride == nil {
			l[i] = map[string]interface{}{}
			continue
		}
		m := map[string]interface{}{
			"availability_zone": aws.StringValue(fleetLaunchTemplateOverride.AvailabilityZone),
			"instance_type":     aws.StringValue(fleetLaunchTemplateOverride.InstanceType),
			"max_price":         aws.StringValue(fleetLaunchTemplateOverride.MaxPrice),
			"priority":          aws.Float64Value(fleetLaunchTemplateOverride.Priority),
			"subnet_id":         aws.StringValue(fleetLaunchTemplateOverride.SubnetId),
			"weighted_capacity": aws.Float64Value(fleetLaunchTemplateOverride.WeightedCapacity),
		}
		l[i] = m
	}

	return l
}

func flattenEc2FleetLaunchTemplateSpecification(fleetLaunchTemplateSpecification *ec2.FleetLaunchTemplateSpecification) []interface{} {
	if fleetLaunchTemplateSpecification == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"launch_template_id":   aws.StringValue(fleetLaunchTemplateSpecification.LaunchTemplateId),
		"launch_template_name": aws.StringValue(fleetLaunchTemplateSpecification.LaunchTemplateName),
		"version":              aws.StringValue(fleetLaunchTemplateSpecification.Version),
	}

	return []interface{}{m}
}

func flattenEc2OnDemandOptions(onDemandOptions *ec2.OnDemandOptions) []interface{} {
	if onDemandOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"allocation_strategy": aws.StringValue(onDemandOptions.AllocationStrategy),
	}

	return []interface{}{m}
}

func flattenEc2SpotOptions(spotOptions *ec2.SpotOptions) []interface{} {
	if spotOptions == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"allocation_strategy":            aws.StringValue(spotOptions.AllocationStrategy),
		"instance_interruption_behavior": aws.StringValue(spotOptions.InstanceInterruptionBehavior),
		"instance_pools_to_use_count":    aws.Int64Value(spotOptions.InstancePoolsToUseCount),
		"maintenance_strategies":         flattenFleetSpotMaintenanceStrategies(spotOptions.MaintenanceStrategies),
	}

	// API will omit InstancePoolsToUseCount if AllocationStrategy is diversified, which breaks our Default: 1
	// Here we just reset it to 1 to prevent removing the Default and setting up a special DiffSuppressFunc
	if spotOptions.InstancePoolsToUseCount == nil && aws.StringValue(spotOptions.AllocationStrategy) == ec2.SpotAllocationStrategyDiversified {
		m["instance_pools_to_use_count"] = 1
	}

	return []interface{}{m}
}

func flattenFleetSpotMaintenanceStrategies(fleetSpotMaintenanceStrategies *ec2.FleetSpotMaintenanceStrategies) []interface{} {
	if fleetSpotMaintenanceStrategies == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"capacity_rebalance": flattenFleetSpotCapacityRebalance(fleetSpotMaintenanceStrategies.CapacityRebalance),
	}

	return []interface{}{m}
}

func flattenFleetSpotCapacityRebalance(fleetSpotCapacityRebalance *ec2.FleetSpotCapacityRebalance) []interface{} {
	if fleetSpotCapacityRebalance == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"replacement_strategy": aws.StringValue(fleetSpotCapacityRebalance.ReplacementStrategy),
	}

	return []interface{}{m}
}

func flattenEc2TargetCapacitySpecification(targetCapacitySpecification *ec2.TargetCapacitySpecification) []interface{} {
	if targetCapacitySpecification == nil {
		return []interface{}{}
	}

	m := map[string]interface{}{
		"default_target_capacity_type": aws.StringValue(targetCapacitySpecification.DefaultTargetCapacityType),
		"on_demand_target_capacity":    aws.Int64Value(targetCapacitySpecification.OnDemandTargetCapacity),
		"spot_target_capacity":         aws.Int64Value(targetCapacitySpecification.SpotTargetCapacity),
		"total_target_capacity":        aws.Int64Value(targetCapacitySpecification.TotalTargetCapacity),
	}

	return []interface{}{m}
}
