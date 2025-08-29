// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package arcregionswitch

import (
	"context"
	"fmt"
	"sort"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/arcregionswitch"
	awstypes "github.com/aws/aws-sdk-go-v2/service/arcregionswitch/types"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	fwdiag "github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	fwschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/listplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework"
	fwvalidators "github.com/hashicorp/terraform-provider-aws/internal/framework/validators"
	tftags "github.com/hashicorp/terraform-provider-aws/internal/tags"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func FindPlanByARN(ctx context.Context, conn *arcregionswitch.Client, arn string) (*awstypes.Plan, error) {
	input := arcregionswitch.GetPlanInput{
		Arn: aws.String(arn),
	}

	output, err := conn.GetPlan(ctx, &input)

	if err != nil {
		return nil, err
	}

	if output == nil || output.Plan == nil {
		return nil, tfresource.NewEmptyResultError(input)
	}

	return output.Plan, nil
}

func ListTags(ctx context.Context, conn *arcregionswitch.Client, identifier string) (tftags.KeyValueTags, error) {
	input := arcregionswitch.ListTagsForResourceInput{
		Arn: aws.String(identifier),
	}

	output, err := conn.ListTagsForResource(ctx, &input)

	if err != nil {
		return tftags.New(ctx, nil), err
	}

	return tftags.New(ctx, output.ResourceTags), nil
}

func UpdateTags(ctx context.Context, conn *arcregionswitch.Client, identifier string, oldTagsMap, newTagsMap any) error {
	oldTags := tftags.New(ctx, oldTagsMap)
	newTags := tftags.New(ctx, newTagsMap)

	if removedTags := oldTags.Removed(newTags); len(removedTags) > 0 {
		input := arcregionswitch.UntagResourceInput{
			Arn:             aws.String(identifier),
			ResourceTagKeys: removedTags.Keys(),
		}

		_, err := conn.UntagResource(ctx, &input)

		if err != nil {
			return fmt.Errorf("untagging resource (%s): %w", identifier, err)
		}
	}

	if updatedTags := oldTags.Updated(newTags); len(updatedTags) > 0 {
		input := arcregionswitch.TagResourceInput{
			Arn:  aws.String(identifier),
			Tags: updatedTags.Map(),
		}

		_, err := conn.TagResource(ctx, &input)

		if err != nil {
			return fmt.Errorf("tagging resource (%s): %w", identifier, err)
		}
	}

	return nil
}

// @FrameworkResource("aws_arcregionswitch_plan", name="Plan")
// @Tags(identifierAttribute="arn")
func newResourcePlan(context.Context) (resource.ResourceWithConfigure, error) {
	r := &resourcePlan{}
	return r, nil
}

type resourcePlan struct {
	framework.ResourceWithConfigure
}

func (r *resourcePlan) ValidateModel(ctx context.Context, schema *fwschema.Schema) fwdiag.Diagnostics {
	var diags fwdiag.Diagnostics
	// Basic validation is handled by the schema validators
	return diags
}

type resourcePlanModel struct {
	ARN                          types.String `tfsdk:"arn"`
	ID                           types.String `tfsdk:"id"`
	Name                         types.String `tfsdk:"name"`
	ExecutionRole                types.String `tfsdk:"execution_role"`
	RecoveryApproach             types.String `tfsdk:"recovery_approach"`
	Regions                      types.List   `tfsdk:"regions"`
	Description                  types.String `tfsdk:"description"`
	PrimaryRegion                types.String `tfsdk:"primary_region"`
	RecoveryTimeObjectiveMinutes types.Int64  `tfsdk:"recovery_time_objective_minutes"`
	AssociatedAlarms             types.Set    `tfsdk:"associated_alarms"`
	Workflow                     types.List   `tfsdk:"workflow"`
	Region                       types.String `tfsdk:"region"`
	Tags                         tftags.Map   `tfsdk:"tags"`
	TagsAll                      tftags.Map   `tfsdk:"tags_all"`
}

type associatedAlarmModel struct {
	Name               types.String `tfsdk:"name"`
	AlarmType          types.String `tfsdk:"alarm_type"`
	ResourceIdentifier types.String `tfsdk:"resource_identifier"`
	CrossAccountRole   types.String `tfsdk:"cross_account_role"`
	ExternalId         types.String `tfsdk:"external_id"`
}

type workflowModel struct {
	WorkflowTargetAction types.String `tfsdk:"workflow_target_action"`
	WorkflowTargetRegion types.String `tfsdk:"workflow_target_region"`
	WorkflowDescription  types.String `tfsdk:"workflow_description"`
	Step                 types.List   `tfsdk:"step"`
}

type route53HealthCheckConfigModel struct {
	HostedZoneId     types.String `tfsdk:"hosted_zone_id"`
	RecordName       types.String `tfsdk:"record_name"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
	TimeoutMinutes   types.Int64  `tfsdk:"timeout_minutes"`
	RecordSets       types.List   `tfsdk:"record_sets"`
}

type recordSetModel struct {
	RecordSetIdentifier types.String `tfsdk:"record_set_identifier"`
	Region              types.String `tfsdk:"region"`
}

type stepModel struct {
	Name                         types.String `tfsdk:"name"`
	ExecutionBlockType           types.String `tfsdk:"execution_block_type"`
	Description                  types.String `tfsdk:"description"`
	ExecutionApprovalConfig      types.List   `tfsdk:"execution_approval_config"`
	Route53HealthCheckConfig     types.List   `tfsdk:"route53_health_check_config"`
	CustomActionLambdaConfig     types.List   `tfsdk:"custom_action_lambda_config"`
	GlobalAuroraConfig           types.List   `tfsdk:"global_aurora_config"`
	Ec2AsgCapacityIncreaseConfig types.List   `tfsdk:"ec2_asg_capacity_increase_config"`
	EcsCapacityIncreaseConfig    types.List   `tfsdk:"ecs_capacity_increase_config"`
	EksResourceScalingConfig     types.List   `tfsdk:"eks_resource_scaling_config"`
	ArcRoutingControlConfig      types.List   `tfsdk:"arc_routing_control_config"`
	ParallelConfig               types.List   `tfsdk:"parallel_config"`
}

type executionApprovalConfigModel struct {
	ApprovalRole   types.String `tfsdk:"approval_role"`
	TimeoutMinutes types.Int64  `tfsdk:"timeout_minutes"`
}

type customActionLambdaConfigModel struct {
	RegionToRun          types.String  `tfsdk:"region_to_run"`
	RetryIntervalMinutes types.Float64 `tfsdk:"retry_interval_minutes"`
	TimeoutMinutes       types.Int64   `tfsdk:"timeout_minutes"`
	Lambda               types.List    `tfsdk:"lambda"`
	Ungraceful           types.List    `tfsdk:"ungraceful"`
}

type lambdaModel struct {
	ARN              types.String `tfsdk:"arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalID       types.String `tfsdk:"external_id"`
}

type ungracefulModel struct {
	Behavior types.String `tfsdk:"behavior"`
}

// Global Aurora Configuration Models
type globalAuroraConfigModel struct {
	Behavior                types.String `tfsdk:"behavior"`
	GlobalClusterIdentifier types.String `tfsdk:"global_cluster_identifier"`
	DatabaseClusterArns     types.List   `tfsdk:"database_cluster_arns"`
	CrossAccountRole        types.String `tfsdk:"cross_account_role"`
	ExternalId              types.String `tfsdk:"external_id"`
	TimeoutMinutes          types.Int64  `tfsdk:"timeout_minutes"`
	Ungraceful              types.List   `tfsdk:"ungraceful"`
}

type globalAuroraUngracefulModel struct {
	Ungraceful types.String `tfsdk:"ungraceful"`
}

// EC2 ASG Configuration Models
type ec2AsgCapacityIncreaseConfigModel struct {
	CapacityMonitoringApproach types.String `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64  `tfsdk:"timeout_minutes"`
	Asgs                       types.List   `tfsdk:"asgs"`
	Ungraceful                 types.List   `tfsdk:"ungraceful"`
}

type asgModel struct {
	ARN              types.String `tfsdk:"arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
}

type ec2UngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// ECS Configuration Models
type ecsCapacityIncreaseConfigModel struct {
	CapacityMonitoringApproach types.String `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64  `tfsdk:"timeout_minutes"`
	Services                   types.List   `tfsdk:"services"`
	Ungraceful                 types.List   `tfsdk:"ungraceful"`
}

type serviceModel struct {
	ClusterArn       types.String `tfsdk:"cluster_arn"`
	ServiceArn       types.String `tfsdk:"service_arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
}

type ecsUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// EKS Configuration Models
type eksResourceScalingConfigModel struct {
	CapacityMonitoringApproach types.String `tfsdk:"capacity_monitoring_approach"`
	TargetPercent              types.Int64  `tfsdk:"target_percent"`
	TimeoutMinutes             types.Int64  `tfsdk:"timeout_minutes"`
	KubernetesResourceType     types.List   `tfsdk:"kubernetes_resource_type"`
	EksClusters                types.List   `tfsdk:"eks_clusters"`
	ScalingResources           types.List   `tfsdk:"scaling_resources"`
	Ungraceful                 types.List   `tfsdk:"ungraceful"`
}

type kubernetesResourceTypeModel struct {
	ApiVersion types.String `tfsdk:"api_version"`
	Kind       types.String `tfsdk:"kind"`
}

type eksClusterModel struct {
	ClusterArn       types.String `tfsdk:"cluster_arn"`
	CrossAccountRole types.String `tfsdk:"cross_account_role"`
	ExternalId       types.String `tfsdk:"external_id"`
}

type scalingResourcesModel struct {
	Namespace types.String `tfsdk:"namespace"`
	Resources types.List   `tfsdk:"resources"`
}

type kubernetesScalingResourceModel struct {
	ResourceName types.String `tfsdk:"resource_name"`
	Name         types.String `tfsdk:"name"`
	Namespace    types.String `tfsdk:"namespace"`
	HpaName      types.String `tfsdk:"hpa_name"`
}

type eksUngracefulModel struct {
	MinimumSuccessPercentage types.Int64 `tfsdk:"minimum_success_percentage"`
}

// ARC Routing Control Configuration Models
type arcRoutingControlConfigModel struct {
	CrossAccountRole         types.String `tfsdk:"cross_account_role"`
	ExternalId               types.String `tfsdk:"external_id"`
	TimeoutMinutes           types.Int64  `tfsdk:"timeout_minutes"`
	RegionAndRoutingControls types.List   `tfsdk:"region_and_routing_controls"`
}

type regionAndRoutingControlsModel struct {
	Region             types.String `tfsdk:"region"`
	RoutingControlArns types.List   `tfsdk:"routing_control_arns"`
}

// Parallel Configuration Models
type parallelConfigModel struct {
	Step types.List `tfsdk:"step"`
}

type parallelStepModel struct {
	Name                     types.String `tfsdk:"name"`
	ExecutionBlockType       types.String `tfsdk:"execution_block_type"`
	Description              types.String `tfsdk:"description"`
	ExecutionApprovalConfig  types.List   `tfsdk:"execution_approval_config"`
	CustomActionLambdaConfig types.List   `tfsdk:"custom_action_lambda_config"`
}

func (r *resourcePlan) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan resourcePlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	var regions []string
	resp.Diagnostics.Append(plan.Regions.ElementsAs(ctx, &regions, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := arcregionswitch.CreatePlanInput{
		Name:             aws.String(plan.Name.ValueString()),
		ExecutionRole:    aws.String(plan.ExecutionRole.ValueString()),
		RecoveryApproach: awstypes.RecoveryApproach(plan.RecoveryApproach.ValueString()),
		Regions:          regions,
	}

	if !plan.Description.IsNull() {
		input.Description = aws.String(plan.Description.ValueString())
	}

	if !plan.PrimaryRegion.IsNull() {
		input.PrimaryRegion = aws.String(plan.PrimaryRegion.ValueString())
	}

	if !plan.RecoveryTimeObjectiveMinutes.IsNull() {
		input.RecoveryTimeObjectiveMinutes = aws.Int32(int32(plan.RecoveryTimeObjectiveMinutes.ValueInt64()))
	}

	// Handle workflows - API requires this field
	if !plan.Workflow.IsNull() {
		var workflows []workflowModel
		resp.Diagnostics.Append(plan.Workflow.ElementsAs(ctx, &workflows, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.Workflows = expandWorkflowsFromFramework(workflows)
	}

	// Handle associated alarms
	if !plan.AssociatedAlarms.IsNull() && !plan.AssociatedAlarms.IsUnknown() {
		var alarms []associatedAlarmModel
		resp.Diagnostics.Append(plan.AssociatedAlarms.ElementsAs(ctx, &alarms, false)...)
		if resp.Diagnostics.HasError() {
			return
		}
		input.AssociatedAlarms = expandAssociatedAlarmsFromFramework(alarms)
	}

	// Handle tags - use getTagsIn to get all tags including provider defaults
	input.Tags = getTagsIn(ctx)

	output, err := conn.CreatePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("creating ARC Region Switch Plan", err.Error())
		return
	}

	plan.ARN = types.StringValue(aws.ToString(output.Plan.Arn))
	plan.ID = types.StringValue(aws.ToString(output.Plan.Arn))

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func expandAssociatedAlarmsFromFramework(alarms []associatedAlarmModel) map[string]awstypes.AssociatedAlarm {
	if len(alarms) == 0 {
		return nil
	}

	result := make(map[string]awstypes.AssociatedAlarm)
	for _, alarm := range alarms {
		awsAlarm := awstypes.AssociatedAlarm{
			AlarmType:          awstypes.AlarmType(alarm.AlarmType.ValueString()),
			ResourceIdentifier: aws.String(alarm.ResourceIdentifier.ValueString()),
		}

		if !alarm.CrossAccountRole.IsNull() && !alarm.CrossAccountRole.IsUnknown() {
			awsAlarm.CrossAccountRole = aws.String(alarm.CrossAccountRole.ValueString())
		}

		if !alarm.ExternalId.IsNull() && !alarm.ExternalId.IsUnknown() {
			awsAlarm.ExternalId = aws.String(alarm.ExternalId.ValueString())
		}

		result[alarm.Name.ValueString()] = awsAlarm
	}
	return result
}

func expandWorkflowsFromFramework(workflows []workflowModel) []awstypes.Workflow {
	if len(workflows) == 0 {
		return nil
	}

	result := make([]awstypes.Workflow, len(workflows))
	for i, workflow := range workflows {
		result[i] = awstypes.Workflow{
			WorkflowTargetAction: awstypes.WorkflowTargetAction(workflow.WorkflowTargetAction.ValueString()),
		}

		if !workflow.WorkflowTargetRegion.IsNull() {
			result[i].WorkflowTargetRegion = aws.String(workflow.WorkflowTargetRegion.ValueString())
		}

		if !workflow.WorkflowDescription.IsNull() {
			result[i].WorkflowDescription = aws.String(workflow.WorkflowDescription.ValueString())
		}

		// Handle steps
		if !workflow.Step.IsNull() {
			var steps []stepModel
			workflow.Step.ElementsAs(context.Background(), &steps, false)

			result[i].Steps = make([]awstypes.Step, len(steps))
			for j, step := range steps {
				result[i].Steps[j] = awstypes.Step{
					Name:               aws.String(step.Name.ValueString()),
					ExecutionBlockType: awstypes.ExecutionBlockType(step.ExecutionBlockType.ValueString()),
				}

				if !step.Description.IsNull() {
					result[i].Steps[j].Description = aws.String(step.Description.ValueString())
				}

				// Handle execution approval config
				if !step.ExecutionApprovalConfig.IsNull() {
					var approvalConfigs []executionApprovalConfigModel
					step.ExecutionApprovalConfig.ElementsAs(context.Background(), &approvalConfigs, false)

					if len(approvalConfigs) > 0 {
						approvalConfig := approvalConfigs[0]
						config := awstypes.ExecutionApprovalConfiguration{
							ApprovalRole: aws.String(approvalConfig.ApprovalRole.ValueString()),
						}
						if !approvalConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(approvalConfig.TimeoutMinutes.ValueInt64()))
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
							Value: config,
						}
					}
				}

				// Handle route53 health check config
				if !step.Route53HealthCheckConfig.IsNull() {
					var healthCheckConfigs []route53HealthCheckConfigModel
					step.Route53HealthCheckConfig.ElementsAs(context.Background(), &healthCheckConfigs, false)

					if len(healthCheckConfigs) > 0 {
						healthCheckConfig := healthCheckConfigs[0]
						config := awstypes.Route53HealthCheckConfiguration{
							HostedZoneId: aws.String(healthCheckConfig.HostedZoneId.ValueString()),
							RecordName:   aws.String(healthCheckConfig.RecordName.ValueString()),
						}

						if !healthCheckConfig.CrossAccountRole.IsNull() {
							config.CrossAccountRole = aws.String(healthCheckConfig.CrossAccountRole.ValueString())
						}
						if !healthCheckConfig.ExternalId.IsNull() {
							config.ExternalId = aws.String(healthCheckConfig.ExternalId.ValueString())
						}
						if !healthCheckConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(healthCheckConfig.TimeoutMinutes.ValueInt64()))
						}

						// Handle record sets
						if !healthCheckConfig.RecordSets.IsNull() {
							var recordSets []recordSetModel
							healthCheckConfig.RecordSets.ElementsAs(context.Background(), &recordSets, false)

							config.RecordSets = make([]awstypes.Route53ResourceRecordSet, len(recordSets))
							for k, recordSet := range recordSets {
								config.RecordSets[k] = awstypes.Route53ResourceRecordSet{
									RecordSetIdentifier: aws.String(recordSet.RecordSetIdentifier.ValueString()),
									Region:              aws.String(recordSet.Region.ValueString()),
								}
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig{
							Value: config,
						}
					}
				}

				// Handle custom action lambda config
				if !step.CustomActionLambdaConfig.IsNull() {
					var lambdaConfigs []customActionLambdaConfigModel
					step.CustomActionLambdaConfig.ElementsAs(context.Background(), &lambdaConfigs, false)

					if len(lambdaConfigs) > 0 {
						lambdaConfig := lambdaConfigs[0]
						config := awstypes.CustomActionLambdaConfiguration{
							RegionToRun:          awstypes.RegionToRunIn(lambdaConfig.RegionToRun.ValueString()),
							RetryIntervalMinutes: aws.Float32(float32(lambdaConfig.RetryIntervalMinutes.ValueFloat64())),
						}

						if !lambdaConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(lambdaConfig.TimeoutMinutes.ValueInt64()))
						}

						// Handle lambdas
						if !lambdaConfig.Lambda.IsNull() {
							var lambdas []lambdaModel
							lambdaConfig.Lambda.ElementsAs(context.Background(), &lambdas, false)

							config.Lambdas = make([]awstypes.Lambdas, len(lambdas))
							for k, lambda := range lambdas {
								config.Lambdas[k] = awstypes.Lambdas{
									Arn: aws.String(lambda.ARN.ValueString()),
								}
								if !lambda.CrossAccountRole.IsNull() {
									config.Lambdas[k].CrossAccountRole = aws.String(lambda.CrossAccountRole.ValueString())
								}
								if !lambda.ExternalID.IsNull() {
									config.Lambdas[k].ExternalId = aws.String(lambda.ExternalID.ValueString())
								}
							}
						}

						// Handle ungraceful
						if !lambdaConfig.Ungraceful.IsNull() {
							var ungracefuls []ungracefulModel
							lambdaConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

							if len(ungracefuls) > 0 {
								config.Ungraceful = &awstypes.LambdaUngraceful{
									Behavior: awstypes.LambdaUngracefulBehavior(ungracefuls[0].Behavior.ValueString()),
								}
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
							Value: config,
						}
					}
				}

				// Handle global aurora config
				if !step.GlobalAuroraConfig.IsNull() {
					var auroraConfigs []globalAuroraConfigModel
					step.GlobalAuroraConfig.ElementsAs(context.Background(), &auroraConfigs, false)

					if len(auroraConfigs) > 0 {
						auroraConfig := auroraConfigs[0]
						config := awstypes.GlobalAuroraConfiguration{
							Behavior:                awstypes.GlobalAuroraDefaultBehavior(auroraConfig.Behavior.ValueString()),
							GlobalClusterIdentifier: aws.String(auroraConfig.GlobalClusterIdentifier.ValueString()),
						}

						// Handle database cluster ARNs
						if !auroraConfig.DatabaseClusterArns.IsNull() {
							var arns []string
							auroraConfig.DatabaseClusterArns.ElementsAs(context.Background(), &arns, false)
							config.DatabaseClusterArns = arns
						}

						if !auroraConfig.CrossAccountRole.IsNull() {
							config.CrossAccountRole = aws.String(auroraConfig.CrossAccountRole.ValueString())
						}
						if !auroraConfig.ExternalId.IsNull() {
							config.ExternalId = aws.String(auroraConfig.ExternalId.ValueString())
						}
						if !auroraConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(auroraConfig.TimeoutMinutes.ValueInt64()))
						}

						// Handle ungraceful
						if !auroraConfig.Ungraceful.IsNull() {
							var ungracefuls []globalAuroraUngracefulModel
							auroraConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

							if len(ungracefuls) > 0 {
								config.Ungraceful = &awstypes.GlobalAuroraUngraceful{
									Ungraceful: awstypes.GlobalAuroraUngracefulBehavior(ungracefuls[0].Ungraceful.ValueString()),
								}
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig{
							Value: config,
						}
					}
				}

				// Handle EC2 ASG capacity increase config
				if !step.Ec2AsgCapacityIncreaseConfig.IsNull() {
					var asgConfigs []ec2AsgCapacityIncreaseConfigModel
					step.Ec2AsgCapacityIncreaseConfig.ElementsAs(context.Background(), &asgConfigs, false)

					if len(asgConfigs) > 0 {
						asgConfig := asgConfigs[0]
						config := awstypes.Ec2AsgCapacityIncreaseConfiguration{
							CapacityMonitoringApproach: awstypes.Ec2AsgCapacityMonitoringApproach(asgConfig.CapacityMonitoringApproach.ValueString()),
							TargetPercent:              aws.Int32(int32(asgConfig.TargetPercent.ValueInt64())),
						}

						if !asgConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(asgConfig.TimeoutMinutes.ValueInt64()))
						}

						// Handle ASGs
						if !asgConfig.Asgs.IsNull() {
							var asgs []asgModel
							asgConfig.Asgs.ElementsAs(context.Background(), &asgs, false)

							config.Asgs = make([]awstypes.Asg, len(asgs))
							for k, asg := range asgs {
								config.Asgs[k] = awstypes.Asg{
									Arn: aws.String(asg.ARN.ValueString()),
								}
								if !asg.CrossAccountRole.IsNull() {
									config.Asgs[k].CrossAccountRole = aws.String(asg.CrossAccountRole.ValueString())
								}
								if !asg.ExternalId.IsNull() {
									config.Asgs[k].ExternalId = aws.String(asg.ExternalId.ValueString())
								}
							}
						}

						// Handle ungraceful
						if !asgConfig.Ungraceful.IsNull() {
							var ungracefuls []ec2UngracefulModel
							asgConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

							if len(ungracefuls) > 0 {
								config.Ungraceful = &awstypes.Ec2Ungraceful{
									MinimumSuccessPercentage: aws.Int32(int32(ungracefuls[0].MinimumSuccessPercentage.ValueInt64())),
								}
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig{
							Value: config,
						}
					}
				}

				// Handle ECS capacity increase config
				if !step.EcsCapacityIncreaseConfig.IsNull() {
					var ecsConfigs []ecsCapacityIncreaseConfigModel
					step.EcsCapacityIncreaseConfig.ElementsAs(context.Background(), &ecsConfigs, false)

					if len(ecsConfigs) > 0 {
						ecsConfig := ecsConfigs[0]
						config := awstypes.EcsCapacityIncreaseConfiguration{
							CapacityMonitoringApproach: awstypes.EcsCapacityMonitoringApproach(ecsConfig.CapacityMonitoringApproach.ValueString()),
							TargetPercent:              aws.Int32(int32(ecsConfig.TargetPercent.ValueInt64())),
						}

						if !ecsConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(ecsConfig.TimeoutMinutes.ValueInt64()))
						}

						// Handle services
						if !ecsConfig.Services.IsNull() {
							var services []serviceModel
							ecsConfig.Services.ElementsAs(context.Background(), &services, false)

							config.Services = make([]awstypes.Service, len(services))
							for k, service := range services {
								config.Services[k] = awstypes.Service{
									ClusterArn: aws.String(service.ClusterArn.ValueString()),
									ServiceArn: aws.String(service.ServiceArn.ValueString()),
								}
								if !service.CrossAccountRole.IsNull() {
									config.Services[k].CrossAccountRole = aws.String(service.CrossAccountRole.ValueString())
								}
								if !service.ExternalId.IsNull() {
									config.Services[k].ExternalId = aws.String(service.ExternalId.ValueString())
								}
							}
						}

						// Handle ungraceful
						if !ecsConfig.Ungraceful.IsNull() {
							var ungracefuls []ecsUngracefulModel
							ecsConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

							if len(ungracefuls) > 0 {
								config.Ungraceful = &awstypes.EcsUngraceful{
									MinimumSuccessPercentage: aws.Int32(int32(ungracefuls[0].MinimumSuccessPercentage.ValueInt64())),
								}
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig{
							Value: config,
						}
					}
				}

				// Handle EKS resource scaling config
				if !step.EksResourceScalingConfig.IsNull() {
					var eksConfigs []eksResourceScalingConfigModel
					step.EksResourceScalingConfig.ElementsAs(context.Background(), &eksConfigs, false)

					if len(eksConfigs) > 0 {
						eksConfig := eksConfigs[0]
						config := awstypes.EksResourceScalingConfiguration{
							CapacityMonitoringApproach: awstypes.EksCapacityMonitoringApproach(eksConfig.CapacityMonitoringApproach.ValueString()),
							TargetPercent:              aws.Int32(int32(eksConfig.TargetPercent.ValueInt64())),
						}

						if !eksConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(eksConfig.TimeoutMinutes.ValueInt64()))
						}

						// Handle kubernetes resource type
						if !eksConfig.KubernetesResourceType.IsNull() {
							var resourceTypes []kubernetesResourceTypeModel
							eksConfig.KubernetesResourceType.ElementsAs(context.Background(), &resourceTypes, false)

							if len(resourceTypes) > 0 {
								config.KubernetesResourceType = &awstypes.KubernetesResourceType{
									ApiVersion: aws.String(resourceTypes[0].ApiVersion.ValueString()),
									Kind:       aws.String(resourceTypes[0].Kind.ValueString()),
								}
							}
						}

						// Handle EKS clusters
						if !eksConfig.EksClusters.IsNull() {
							var clusters []eksClusterModel
							eksConfig.EksClusters.ElementsAs(context.Background(), &clusters, false)

							config.EksClusters = make([]awstypes.EksCluster, len(clusters))
							for k, cluster := range clusters {
								config.EksClusters[k] = awstypes.EksCluster{
									ClusterArn: aws.String(cluster.ClusterArn.ValueString()),
								}
								if !cluster.CrossAccountRole.IsNull() {
									config.EksClusters[k].CrossAccountRole = aws.String(cluster.CrossAccountRole.ValueString())
								}
								if !cluster.ExternalId.IsNull() {
									config.EksClusters[k].ExternalId = aws.String(cluster.ExternalId.ValueString())
								}
							}
						}

						// Handle scaling resources
						if !eksConfig.ScalingResources.IsNull() {
							var scalingResources []scalingResourcesModel
							eksConfig.ScalingResources.ElementsAs(context.Background(), &scalingResources, false)

							config.ScalingResources = make([]map[string]map[string]awstypes.KubernetesScalingResource, len(scalingResources))
							for k, scalingResource := range scalingResources {
								if !scalingResource.Resources.IsNull() {
									var resources []kubernetesScalingResourceModel
									scalingResource.Resources.ElementsAs(context.Background(), &resources, false)

									regionMap := make(map[string]awstypes.KubernetesScalingResource)
									for _, resource := range resources {
										scalingResource := awstypes.KubernetesScalingResource{
											Name:      aws.String(resource.Name.ValueString()),
											Namespace: aws.String(resource.Namespace.ValueString()),
										}
										if !resource.HpaName.IsNull() {
											scalingResource.HpaName = aws.String(resource.HpaName.ValueString())
										}
										regionMap[resource.ResourceName.ValueString()] = scalingResource
									}
									config.ScalingResources[k] = map[string]map[string]awstypes.KubernetesScalingResource{
										scalingResource.Namespace.ValueString(): regionMap,
									}
								}
							}
						}

						// Handle ungraceful
						if !eksConfig.Ungraceful.IsNull() {
							var ungracefuls []eksUngracefulModel
							eksConfig.Ungraceful.ElementsAs(context.Background(), &ungracefuls, false)

							if len(ungracefuls) > 0 {
								config.Ungraceful = &awstypes.EksResourceScalingUngraceful{
									MinimumSuccessPercentage: aws.Int32(int32(ungracefuls[0].MinimumSuccessPercentage.ValueInt64())),
								}
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig{
							Value: config,
						}
					}
				}

				// Handle ARC routing control config
				if !step.ArcRoutingControlConfig.IsNull() {
					var routingConfigs []arcRoutingControlConfigModel
					step.ArcRoutingControlConfig.ElementsAs(context.Background(), &routingConfigs, false)

					if len(routingConfigs) > 0 {
						routingConfig := routingConfigs[0]
						config := awstypes.ArcRoutingControlConfiguration{
							RegionAndRoutingControls: make(map[string][]awstypes.ArcRoutingControlState),
						}

						if !routingConfig.CrossAccountRole.IsNull() {
							config.CrossAccountRole = aws.String(routingConfig.CrossAccountRole.ValueString())
						}
						if !routingConfig.ExternalId.IsNull() {
							config.ExternalId = aws.String(routingConfig.ExternalId.ValueString())
						}
						if !routingConfig.TimeoutMinutes.IsNull() {
							config.TimeoutMinutes = aws.Int32(int32(routingConfig.TimeoutMinutes.ValueInt64()))
						}

						// Handle region and routing controls
						if !routingConfig.RegionAndRoutingControls.IsNull() {
							var regionControls []regionAndRoutingControlsModel
							routingConfig.RegionAndRoutingControls.ElementsAs(context.Background(), &regionControls, false)

							for _, regionControl := range regionControls {
								var controlArns []string
								regionControl.RoutingControlArns.ElementsAs(context.Background(), &controlArns, false)

								controlStates := make([]awstypes.ArcRoutingControlState, len(controlArns))
								for l, arn := range controlArns {
									controlStates[l] = awstypes.ArcRoutingControlState{
										RoutingControlArn: aws.String(arn),
										State:             awstypes.RoutingControlStateChangeOn, // Default state
									}
								}
								config.RegionAndRoutingControls[regionControl.Region.ValueString()] = controlStates
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig{
							Value: config,
						}
					}
				}

				// Handle parallel config
				if !step.ParallelConfig.IsNull() {
					var parallelConfigs []parallelConfigModel
					step.ParallelConfig.ElementsAs(context.Background(), &parallelConfigs, false)

					if len(parallelConfigs) > 0 {
						parallelConfig := parallelConfigs[0]
						config := awstypes.ParallelExecutionBlockConfiguration{}

						// Handle parallel steps
						if !parallelConfig.Step.IsNull() {
							var parallelSteps []parallelStepModel
							parallelConfig.Step.ElementsAs(context.Background(), &parallelSteps, false)

							config.Steps = make([]awstypes.Step, len(parallelSteps))
							for k, parallelStep := range parallelSteps {
								config.Steps[k] = awstypes.Step{
									Name:               aws.String(parallelStep.Name.ValueString()),
									ExecutionBlockType: awstypes.ExecutionBlockType(parallelStep.ExecutionBlockType.ValueString()),
								}

								if !parallelStep.Description.IsNull() {
									config.Steps[k].Description = aws.String(parallelStep.Description.ValueString())
								}

								// Handle execution block configuration for parallel steps
								if !parallelStep.ExecutionApprovalConfig.IsNull() {
									var approvalConfigs []executionApprovalConfigModel
									parallelStep.ExecutionApprovalConfig.ElementsAs(context.Background(), &approvalConfigs, false)

									if len(approvalConfigs) > 0 {
										approvalConfig := approvalConfigs[0]
										config.Steps[k].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig{
											Value: awstypes.ExecutionApprovalConfiguration{
												ApprovalRole: aws.String(approvalConfig.ApprovalRole.ValueString()),
											},
										}
										if !approvalConfig.TimeoutMinutes.IsNull() {
											config.Steps[k].ExecutionBlockConfiguration.(*awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig).Value.TimeoutMinutes = aws.Int32(int32(approvalConfig.TimeoutMinutes.ValueInt64()))
										}
									}
								} else if !parallelStep.CustomActionLambdaConfig.IsNull() {
									var lambdaConfigs []customActionLambdaConfigModel
									parallelStep.CustomActionLambdaConfig.ElementsAs(context.Background(), &lambdaConfigs, false)

									if len(lambdaConfigs) > 0 {
										lambdaConfig := lambdaConfigs[0]
										lambdaConfigValue := awstypes.CustomActionLambdaConfiguration{
											RegionToRun:          awstypes.RegionToRunIn(lambdaConfig.RegionToRun.ValueString()),
											RetryIntervalMinutes: aws.Float32(float32(lambdaConfig.RetryIntervalMinutes.ValueFloat64())),
										}

										if !lambdaConfig.TimeoutMinutes.IsNull() {
											lambdaConfigValue.TimeoutMinutes = aws.Int32(int32(lambdaConfig.TimeoutMinutes.ValueInt64()))
										}

										// Handle lambdas
										if !lambdaConfig.Lambda.IsNull() {
											var lambdas []lambdaModel
											lambdaConfig.Lambda.ElementsAs(context.Background(), &lambdas, false)

											lambdaConfigValue.Lambdas = make([]awstypes.Lambdas, len(lambdas))
											for l, lambda := range lambdas {
												lambdaConfigValue.Lambdas[l] = awstypes.Lambdas{
													Arn: aws.String(lambda.ARN.ValueString()),
												}
												if !lambda.CrossAccountRole.IsNull() {
													lambdaConfigValue.Lambdas[l].CrossAccountRole = aws.String(lambda.CrossAccountRole.ValueString())
												}
												if !lambda.ExternalID.IsNull() {
													lambdaConfigValue.Lambdas[l].ExternalId = aws.String(lambda.ExternalID.ValueString())
												}
											}
										}

										config.Steps[k].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig{
											Value: lambdaConfigValue,
										}
									}
								}
							}
						}

						result[i].Steps[j].ExecutionBlockConfiguration = &awstypes.ExecutionBlockConfigurationMemberParallelConfig{
							Value: config,
						}
					}
				}
			}
		} else {
			result[i].Steps = []awstypes.Step{}
		}
	}

	return result
}

func flattenWorkflowsToFramework(ctx context.Context, workflows []awstypes.Workflow) (types.List, fwdiag.Diagnostics) {
	var diags fwdiag.Diagnostics

	if len(workflows) == 0 {
		return types.ListNull(getWorkflowObjectType()), diags
	}

	// Sort workflows by target action to ensure consistent ordering (activate first, then deactivate)
	sortedWorkflows := make([]awstypes.Workflow, len(workflows))
	copy(sortedWorkflows, workflows)

	// Simple sort: activate workflows first, then deactivate workflows
	var activateWorkflows, deactivateWorkflows []awstypes.Workflow
	for _, workflow := range sortedWorkflows {
		if workflow.WorkflowTargetAction == awstypes.WorkflowTargetActionActivate {
			activateWorkflows = append(activateWorkflows, workflow)
		} else {
			deactivateWorkflows = append(deactivateWorkflows, workflow)
		}
	}

	// Combine: activate workflows first, then deactivate workflows
	sortedWorkflows = append(activateWorkflows, deactivateWorkflows...)

	elements := make([]attr.Value, len(sortedWorkflows))
	for i, workflow := range sortedWorkflows {
		workflowAttrs := map[string]attr.Value{
			"workflow_target_action": types.StringValue(string(workflow.WorkflowTargetAction)),
		}

		if workflow.WorkflowTargetRegion != nil {
			workflowAttrs["workflow_target_region"] = types.StringValue(aws.ToString(workflow.WorkflowTargetRegion))
		} else {
			workflowAttrs["workflow_target_region"] = types.StringNull()
		}

		if workflow.WorkflowDescription != nil {
			workflowAttrs["workflow_description"] = types.StringValue(aws.ToString(workflow.WorkflowDescription))
		} else {
			workflowAttrs["workflow_description"] = types.StringNull()
		}

		// Handle steps
		if len(workflow.Steps) > 0 {
			stepElements := make([]attr.Value, len(workflow.Steps))
			for j, step := range workflow.Steps {
				stepAttrs := map[string]attr.Value{
					"name":                 types.StringValue(aws.ToString(step.Name)),
					"execution_block_type": types.StringValue(string(step.ExecutionBlockType)),
				}

				if step.Description != nil {
					stepAttrs["description"] = types.StringValue(aws.ToString(step.Description))
				} else {
					stepAttrs["description"] = types.StringNull()
				}

				// Handle execution block configuration
				if step.ExecutionBlockConfiguration != nil {
					// Initialize all execution block configs to null first
					stepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
					stepAttrs["route53_health_check_config"] = types.ListNull(getRoute53HealthCheckConfigObjectType())
					stepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
					stepAttrs["global_aurora_config"] = types.ListNull(getGlobalAuroraConfigObjectType())
					stepAttrs["ec2_asg_capacity_increase_config"] = types.ListNull(getEc2AsgCapacityIncreaseConfigObjectType())
					stepAttrs["ecs_capacity_increase_config"] = types.ListNull(getEcsCapacityIncreaseConfigObjectType())
					stepAttrs["eks_resource_scaling_config"] = types.ListNull(getEksResourceScalingConfigObjectType())
					stepAttrs["arc_routing_control_config"] = types.ListNull(getArcRoutingControlConfigObjectType())
					stepAttrs["parallel_config"] = types.ListNull(getParallelConfigObjectType())

					// Now each case only needs to set its specific config
					switch config := step.ExecutionBlockConfiguration.(type) {
					case *awstypes.ExecutionBlockConfigurationMemberExecutionApprovalConfig:
						approvalAttrs := map[string]attr.Value{
							"approval_role": types.StringValue(aws.ToString(config.Value.ApprovalRole)),
						}
						if config.Value.TimeoutMinutes != nil {
							approvalAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							approvalAttrs["timeout_minutes"] = types.Int64Null()
						}

						approvalObj, approvalDiags := types.ObjectValue(getExecutionApprovalConfigObjectType().AttrTypes, approvalAttrs)
						diags.Append(approvalDiags...)
						stepAttrs["execution_approval_config"] = types.ListValueMust(getExecutionApprovalConfigObjectType(), []attr.Value{approvalObj})

					case *awstypes.ExecutionBlockConfigurationMemberRoute53HealthCheckConfig:
						healthCheckAttrs := map[string]attr.Value{
							"hosted_zone_id": types.StringValue(aws.ToString(config.Value.HostedZoneId)),
							"record_name":    types.StringValue(aws.ToString(config.Value.RecordName)),
						}

						if config.Value.CrossAccountRole != nil {
							healthCheckAttrs["cross_account_role"] = types.StringValue(aws.ToString(config.Value.CrossAccountRole))
						} else {
							healthCheckAttrs["cross_account_role"] = types.StringNull()
						}

						if config.Value.ExternalId != nil {
							healthCheckAttrs["external_id"] = types.StringValue(aws.ToString(config.Value.ExternalId))
						} else {
							healthCheckAttrs["external_id"] = types.StringNull()
						}

						if config.Value.TimeoutMinutes != nil {
							healthCheckAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							healthCheckAttrs["timeout_minutes"] = types.Int64Null()
						}

						// Handle record sets
						if len(config.Value.RecordSets) > 0 {
							recordSetElements := make([]attr.Value, len(config.Value.RecordSets))
							for k, recordSet := range config.Value.RecordSets {
								recordSetAttrs := map[string]attr.Value{
									"record_set_identifier": types.StringValue(aws.ToString(recordSet.RecordSetIdentifier)),
									"region":                types.StringValue(aws.ToString(recordSet.Region)),
								}
								recordSetObj, recordSetDiags := types.ObjectValue(getRecordSetObjectType().AttrTypes, recordSetAttrs)
								diags.Append(recordSetDiags...)
								recordSetElements[k] = recordSetObj
							}
							healthCheckAttrs["record_sets"] = types.ListValueMust(getRecordSetObjectType(), recordSetElements)
						} else {
							healthCheckAttrs["record_sets"] = types.ListNull(getRecordSetObjectType())
						}

						healthCheckObj, healthCheckDiags := types.ObjectValue(getRoute53HealthCheckConfigObjectType().AttrTypes, healthCheckAttrs)
						diags.Append(healthCheckDiags...)
						stepAttrs["route53_health_check_config"] = types.ListValueMust(getRoute53HealthCheckConfigObjectType(), []attr.Value{healthCheckObj})

					case *awstypes.ExecutionBlockConfigurationMemberParallelConfig:
						// Handle parallel config flattening
						parallelStepElements := make([]attr.Value, len(config.Value.Steps))
						for k, parallelStep := range config.Value.Steps {
							parallelStepAttrs := map[string]attr.Value{
								"name":                 types.StringValue(aws.ToString(parallelStep.Name)),
								"execution_block_type": types.StringValue(string(parallelStep.ExecutionBlockType)),
							}

							if parallelStep.Description != nil {
								parallelStepAttrs["description"] = types.StringValue(aws.ToString(parallelStep.Description))
							} else {
								parallelStepAttrs["description"] = types.StringNull()
							}

							// Handle parallel step execution block configuration
							if parallelStep.ExecutionBlockConfiguration != nil {
								switch parallelConfig := parallelStep.ExecutionBlockConfiguration.(type) {
								case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
									// Flatten custom action lambda config for parallel step
									lambdaAttrs := map[string]attr.Value{
										"region_to_run":          types.StringValue(string(parallelConfig.Value.RegionToRun)),
										"retry_interval_minutes": types.Float64Value(float64(aws.ToFloat32(parallelConfig.Value.RetryIntervalMinutes))),
									}

									if parallelConfig.Value.TimeoutMinutes != nil {
										lambdaAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(parallelConfig.Value.TimeoutMinutes)))
									} else {
										lambdaAttrs["timeout_minutes"] = types.Int64Null()
									}

									// Handle lambdas
									if len(parallelConfig.Value.Lambdas) > 0 {
										lambdaElements := make([]attr.Value, len(parallelConfig.Value.Lambdas))
										for l, lambda := range parallelConfig.Value.Lambdas {
											lambdaElementAttrs := map[string]attr.Value{
												"arn": types.StringValue(aws.ToString(lambda.Arn)),
											}
											if lambda.CrossAccountRole != nil {
												lambdaElementAttrs["cross_account_role"] = types.StringValue(aws.ToString(lambda.CrossAccountRole))
											} else {
												lambdaElementAttrs["cross_account_role"] = types.StringNull()
											}
											if lambda.ExternalId != nil {
												lambdaElementAttrs["external_id"] = types.StringValue(aws.ToString(lambda.ExternalId))
											} else {
												lambdaElementAttrs["external_id"] = types.StringNull()
											}

											lambdaObj, lambdaDiags := types.ObjectValue(getLambdaObjectType().AttrTypes, lambdaElementAttrs)
											diags.Append(lambdaDiags...)
											lambdaElements[l] = lambdaObj
										}
										lambdaAttrs["lambda"] = types.ListValueMust(getLambdaObjectType(), lambdaElements)
									} else {
										lambdaAttrs["lambda"] = types.ListNull(getLambdaObjectType())
									}

									// Handle ungraceful (always null for now since we don't have the data)
									lambdaAttrs["ungraceful"] = types.ListNull(getUngracefulObjectType())

									lambdaObj, lambdaDiags := types.ObjectValue(getCustomActionLambdaConfigObjectType().AttrTypes, lambdaAttrs)
									diags.Append(lambdaDiags...)
									parallelStepAttrs["custom_action_lambda_config"] = types.ListValueMust(getCustomActionLambdaConfigObjectType(), []attr.Value{lambdaObj})
									parallelStepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
								default:
									parallelStepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
									parallelStepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
								}
							} else {
								parallelStepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
								parallelStepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
							}

							parallelStepObj, parallelStepDiags := types.ObjectValue(getParallelStepObjectType().AttrTypes, parallelStepAttrs)
							diags.Append(parallelStepDiags...)
							parallelStepElements[k] = parallelStepObj
						}

						parallelConfigAttrs := map[string]attr.Value{
							"step": types.ListValueMust(getParallelStepObjectType(), parallelStepElements),
						}

						parallelConfigObj, parallelConfigDiags := types.ObjectValue(getParallelConfigObjectType().AttrTypes, parallelConfigAttrs)
						diags.Append(parallelConfigDiags...)
						stepAttrs["parallel_config"] = types.ListValueMust(getParallelConfigObjectType(), []attr.Value{parallelConfigObj})

					case *awstypes.ExecutionBlockConfigurationMemberCustomActionLambdaConfig:
						// Implement custom action lambda config flattening
						lambdaAttrs := map[string]attr.Value{
							"region_to_run":          types.StringValue(string(config.Value.RegionToRun)),
							"retry_interval_minutes": types.Float64Value(float64(aws.ToFloat32(config.Value.RetryIntervalMinutes))),
						}

						if config.Value.TimeoutMinutes != nil {
							lambdaAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							lambdaAttrs["timeout_minutes"] = types.Int64Null()
						}

						// Handle lambdas
						if len(config.Value.Lambdas) > 0 {
							lambdaElements := make([]attr.Value, len(config.Value.Lambdas))
							for l, lambda := range config.Value.Lambdas {
								lambdaElementAttrs := map[string]attr.Value{
									"arn": types.StringValue(aws.ToString(lambda.Arn)),
								}
								if lambda.CrossAccountRole != nil {
									lambdaElementAttrs["cross_account_role"] = types.StringValue(aws.ToString(lambda.CrossAccountRole))
								} else {
									lambdaElementAttrs["cross_account_role"] = types.StringNull()
								}
								if lambda.ExternalId != nil {
									lambdaElementAttrs["external_id"] = types.StringValue(aws.ToString(lambda.ExternalId))
								} else {
									lambdaElementAttrs["external_id"] = types.StringNull()
								}

								lambdaObj, lambdaDiags := types.ObjectValue(getLambdaObjectType().AttrTypes, lambdaElementAttrs)
								diags.Append(lambdaDiags...)
								lambdaElements[l] = lambdaObj
							}
							lambdaAttrs["lambda"] = types.ListValueMust(getLambdaObjectType(), lambdaElements)
						} else {
							lambdaAttrs["lambda"] = types.ListNull(getLambdaObjectType())
						}

						// Handle ungraceful block properly
						if config.Value.Ungraceful != nil {
							ungracefulAttrs := map[string]attr.Value{
								"behavior": types.StringValue(string(config.Value.Ungraceful.Behavior)),
							}
							ungracefulObj, ungracefulDiags := types.ObjectValue(getUngracefulObjectType().AttrTypes, ungracefulAttrs)
							diags.Append(ungracefulDiags...)
							lambdaAttrs["ungraceful"] = types.ListValueMust(getUngracefulObjectType(), []attr.Value{ungracefulObj})
						} else {
							lambdaAttrs["ungraceful"] = types.ListNull(getUngracefulObjectType())
						}

						lambdaObj, lambdaDiags := types.ObjectValue(getCustomActionLambdaConfigObjectType().AttrTypes, lambdaAttrs)
						diags.Append(lambdaDiags...)
						stepAttrs["custom_action_lambda_config"] = types.ListValueMust(getCustomActionLambdaConfigObjectType(), []attr.Value{lambdaObj})

					case *awstypes.ExecutionBlockConfigurationMemberGlobalAuroraConfig:
						auroraAttrs := map[string]attr.Value{
							"behavior":                  types.StringValue(string(config.Value.Behavior)),
							"global_cluster_identifier": types.StringValue(aws.ToString(config.Value.GlobalClusterIdentifier)),
						}

						if config.Value.TimeoutMinutes != nil {
							auroraAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							auroraAttrs["timeout_minutes"] = types.Int64Null()
						}

						if config.Value.CrossAccountRole != nil {
							auroraAttrs["cross_account_role"] = types.StringValue(aws.ToString(config.Value.CrossAccountRole))
						} else {
							auroraAttrs["cross_account_role"] = types.StringNull()
						}

						if config.Value.ExternalId != nil {
							auroraAttrs["external_id"] = types.StringValue(aws.ToString(config.Value.ExternalId))
						} else {
							auroraAttrs["external_id"] = types.StringNull()
						}

						// Handle database cluster ARNs
						if len(config.Value.DatabaseClusterArns) > 0 {
							clusterArns, clusterArnsDiags := types.ListValueFrom(ctx, types.StringType, config.Value.DatabaseClusterArns)
							diags.Append(clusterArnsDiags...)
							auroraAttrs["database_cluster_arns"] = clusterArns
						} else {
							auroraAttrs["database_cluster_arns"] = types.ListNull(types.StringType)
						}

						// Handle ungraceful block
						if config.Value.Ungraceful != nil {
							ungracefulAttrs := map[string]attr.Value{
								"ungraceful": types.StringValue(string(config.Value.Ungraceful.Ungraceful)),
							}
							ungracefulObj, ungracefulDiags := types.ObjectValue(getGlobalAuroraUngracefulObjectType().AttrTypes, ungracefulAttrs)
							diags.Append(ungracefulDiags...)
							auroraAttrs["ungraceful"] = types.ListValueMust(getGlobalAuroraUngracefulObjectType(), []attr.Value{ungracefulObj})
						} else {
							auroraAttrs["ungraceful"] = types.ListNull(getGlobalAuroraUngracefulObjectType())
						}

						auroraObj, auroraDiags := types.ObjectValue(getGlobalAuroraConfigObjectType().AttrTypes, auroraAttrs)
						diags.Append(auroraDiags...)
						stepAttrs["global_aurora_config"] = types.ListValueMust(getGlobalAuroraConfigObjectType(), []attr.Value{auroraObj})

					case *awstypes.ExecutionBlockConfigurationMemberEc2AsgCapacityIncreaseConfig:
						ec2Attrs := map[string]attr.Value{
							"capacity_monitoring_approach": types.StringValue(string(config.Value.CapacityMonitoringApproach)),
							"target_percent":               types.Int64Value(int64(aws.ToInt32(config.Value.TargetPercent))),
						}

						if config.Value.TimeoutMinutes != nil {
							ec2Attrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							ec2Attrs["timeout_minutes"] = types.Int64Null()
						}

						// Handle ASGs
						if len(config.Value.Asgs) > 0 {
							asgElements := make([]attr.Value, len(config.Value.Asgs))
							for a, asg := range config.Value.Asgs {
								asgAttrs := map[string]attr.Value{
									"arn": types.StringValue(aws.ToString(asg.Arn)),
								}
								if asg.CrossAccountRole != nil {
									asgAttrs["cross_account_role"] = types.StringValue(aws.ToString(asg.CrossAccountRole))
								} else {
									asgAttrs["cross_account_role"] = types.StringNull()
								}
								if asg.ExternalId != nil {
									asgAttrs["external_id"] = types.StringValue(aws.ToString(asg.ExternalId))
								} else {
									asgAttrs["external_id"] = types.StringNull()
								}

								asgObj, asgDiags := types.ObjectValue(getAsgObjectType().AttrTypes, asgAttrs)
								diags.Append(asgDiags...)
								asgElements[a] = asgObj
							}
							ec2Attrs["asgs"] = types.ListValueMust(getAsgObjectType(), asgElements)
						} else {
							ec2Attrs["asgs"] = types.ListNull(getAsgObjectType())
						}

						// Handle ungraceful block
						if config.Value.Ungraceful != nil {
							ungracefulAttrs := map[string]attr.Value{
								"minimum_success_percentage": types.Int64Value(int64(aws.ToInt32(config.Value.Ungraceful.MinimumSuccessPercentage))),
							}
							ungracefulObj, ungracefulDiags := types.ObjectValue(getEc2UngracefulObjectType().AttrTypes, ungracefulAttrs)
							diags.Append(ungracefulDiags...)
							ec2Attrs["ungraceful"] = types.ListValueMust(getEc2UngracefulObjectType(), []attr.Value{ungracefulObj})
						} else {
							ec2Attrs["ungraceful"] = types.ListNull(getEc2UngracefulObjectType())
						}

						ec2Obj, ec2Diags := types.ObjectValue(getEc2AsgCapacityIncreaseConfigObjectType().AttrTypes, ec2Attrs)
						diags.Append(ec2Diags...)
						stepAttrs["ec2_asg_capacity_increase_config"] = types.ListValueMust(getEc2AsgCapacityIncreaseConfigObjectType(), []attr.Value{ec2Obj})

					case *awstypes.ExecutionBlockConfigurationMemberEcsCapacityIncreaseConfig:
						ecsAttrs := map[string]attr.Value{
							"capacity_monitoring_approach": types.StringValue(string(config.Value.CapacityMonitoringApproach)),
							"target_percent":               types.Int64Value(int64(aws.ToInt32(config.Value.TargetPercent))),
						}

						if config.Value.TimeoutMinutes != nil {
							ecsAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							ecsAttrs["timeout_minutes"] = types.Int64Null()
						}

						// Handle Services
						if len(config.Value.Services) > 0 {
							serviceElements := make([]attr.Value, len(config.Value.Services))
							for s, service := range config.Value.Services {
								serviceAttrs := map[string]attr.Value{
									"cluster_arn": types.StringValue(aws.ToString(service.ClusterArn)),
									"service_arn": types.StringValue(aws.ToString(service.ServiceArn)),
								}
								if service.CrossAccountRole != nil {
									serviceAttrs["cross_account_role"] = types.StringValue(aws.ToString(service.CrossAccountRole))
								} else {
									serviceAttrs["cross_account_role"] = types.StringNull()
								}
								if service.ExternalId != nil {
									serviceAttrs["external_id"] = types.StringValue(aws.ToString(service.ExternalId))
								} else {
									serviceAttrs["external_id"] = types.StringNull()
								}

								serviceObj, serviceDiags := types.ObjectValue(getServiceObjectType().AttrTypes, serviceAttrs)
								diags.Append(serviceDiags...)
								serviceElements[s] = serviceObj
							}
							ecsAttrs["services"] = types.ListValueMust(getServiceObjectType(), serviceElements)
						} else {
							ecsAttrs["services"] = types.ListNull(getServiceObjectType())
						}

						// Handle ungraceful block
						if config.Value.Ungraceful != nil {
							ungracefulAttrs := map[string]attr.Value{
								"minimum_success_percentage": types.Int64Value(int64(aws.ToInt32(config.Value.Ungraceful.MinimumSuccessPercentage))),
							}
							ungracefulObj, ungracefulDiags := types.ObjectValue(getEcsUngracefulObjectType().AttrTypes, ungracefulAttrs)
							diags.Append(ungracefulDiags...)
							ecsAttrs["ungraceful"] = types.ListValueMust(getEcsUngracefulObjectType(), []attr.Value{ungracefulObj})
						} else {
							ecsAttrs["ungraceful"] = types.ListNull(getEcsUngracefulObjectType())
						}

						ecsObj, ecsDiags := types.ObjectValue(getEcsCapacityIncreaseConfigObjectType().AttrTypes, ecsAttrs)
						diags.Append(ecsDiags...)
						stepAttrs["ecs_capacity_increase_config"] = types.ListValueMust(getEcsCapacityIncreaseConfigObjectType(), []attr.Value{ecsObj})

					case *awstypes.ExecutionBlockConfigurationMemberEksResourceScalingConfig:
						eksAttrs := map[string]attr.Value{
							"capacity_monitoring_approach": types.StringValue(string(config.Value.CapacityMonitoringApproach)),
							"target_percent":               types.Int64Value(int64(aws.ToInt32(config.Value.TargetPercent))),
						}

						if config.Value.TimeoutMinutes != nil {
							eksAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							eksAttrs["timeout_minutes"] = types.Int64Null()
						}

						// Handle EKS Clusters
						if len(config.Value.EksClusters) > 0 {
							clusterElements := make([]attr.Value, len(config.Value.EksClusters))
							for c, cluster := range config.Value.EksClusters {
								clusterAttrs := map[string]attr.Value{
									"cluster_arn": types.StringValue(aws.ToString(cluster.ClusterArn)),
								}
								if cluster.CrossAccountRole != nil {
									clusterAttrs["cross_account_role"] = types.StringValue(aws.ToString(cluster.CrossAccountRole))
								} else {
									clusterAttrs["cross_account_role"] = types.StringNull()
								}
								if cluster.ExternalId != nil {
									clusterAttrs["external_id"] = types.StringValue(aws.ToString(cluster.ExternalId))
								} else {
									clusterAttrs["external_id"] = types.StringNull()
								}

								clusterObj, clusterDiags := types.ObjectValue(getEksClusterObjectType().AttrTypes, clusterAttrs)
								diags.Append(clusterDiags...)
								clusterElements[c] = clusterObj
							}
							eksAttrs["eks_clusters"] = types.ListValueMust(getEksClusterObjectType(), clusterElements)
						} else {
							eksAttrs["eks_clusters"] = types.ListNull(getEksClusterObjectType())
						}

						// Handle Kubernetes Resource Type
						if config.Value.KubernetesResourceType != nil {
							k8sAttrs := map[string]attr.Value{
								"api_version": types.StringValue(aws.ToString(config.Value.KubernetesResourceType.ApiVersion)),
								"kind":        types.StringValue(aws.ToString(config.Value.KubernetesResourceType.Kind)),
							}
							k8sObj, k8sDiags := types.ObjectValue(getKubernetesResourceTypeObjectType().AttrTypes, k8sAttrs)
							diags.Append(k8sDiags...)
							eksAttrs["kubernetes_resource_type"] = types.ListValueMust(getKubernetesResourceTypeObjectType(), []attr.Value{k8sObj})
						} else {
							eksAttrs["kubernetes_resource_type"] = types.ListNull(getKubernetesResourceTypeObjectType())
						}

						// Handle Scaling Resources (slice of maps structure) - sort for consistent ordering
						if len(config.Value.ScalingResources) > 0 {
							scalingElements := make([]attr.Value, 0)
							for _, scalingResourceMap := range config.Value.ScalingResources {
								for namespace, resourceMap := range scalingResourceMap {
									scalingAttrs := map[string]attr.Value{
										"namespace": types.StringValue(namespace),
									}

									// Handle resources within scaling resource - sort by resource_name for consistency
									resourceElements := make([]attr.Value, 0)

									// Create a sorted slice of resource names for consistent ordering
									resourceNames := make([]string, 0, len(resourceMap))
									for resourceName := range resourceMap {
										resourceNames = append(resourceNames, resourceName)
									}
									// Sort by resource name to ensure consistent ordering
									sort.Strings(resourceNames)

									for _, resourceName := range resourceNames {
										resource := resourceMap[resourceName]
										resourceAttrs := map[string]attr.Value{
											"resource_name": types.StringValue(resourceName),
											"name":          types.StringValue(aws.ToString(resource.Name)),
											"namespace":     types.StringValue(aws.ToString(resource.Namespace)),
											"hpa_name":      types.StringValue(aws.ToString(resource.HpaName)),
										}
										resourceObj, resourceDiags := types.ObjectValue(getKubernetesScalingResourceObjectType().AttrTypes, resourceAttrs)
										diags.Append(resourceDiags...)
										resourceElements = append(resourceElements, resourceObj)
									}
									scalingAttrs["resources"] = types.ListValueMust(getKubernetesScalingResourceObjectType(), resourceElements)

									scalingObj, scalingDiags := types.ObjectValue(getScalingResourcesObjectType().AttrTypes, scalingAttrs)
									diags.Append(scalingDiags...)
									scalingElements = append(scalingElements, scalingObj)
								}
							}
							eksAttrs["scaling_resources"] = types.ListValueMust(getScalingResourcesObjectType(), scalingElements)
						} else {
							eksAttrs["scaling_resources"] = types.ListNull(getScalingResourcesObjectType())
						}

						// Handle ungraceful block
						if config.Value.Ungraceful != nil {
							ungracefulAttrs := map[string]attr.Value{
								"minimum_success_percentage": types.Int64Value(int64(aws.ToInt32(config.Value.Ungraceful.MinimumSuccessPercentage))),
							}
							ungracefulObj, ungracefulDiags := types.ObjectValue(getEksUngracefulObjectType().AttrTypes, ungracefulAttrs)
							diags.Append(ungracefulDiags...)
							eksAttrs["ungraceful"] = types.ListValueMust(getEksUngracefulObjectType(), []attr.Value{ungracefulObj})
						} else {
							eksAttrs["ungraceful"] = types.ListNull(getEksUngracefulObjectType())
						}

						eksObj, eksDiags := types.ObjectValue(getEksResourceScalingConfigObjectType().AttrTypes, eksAttrs)
						diags.Append(eksDiags...)
						stepAttrs["eks_resource_scaling_config"] = types.ListValueMust(getEksResourceScalingConfigObjectType(), []attr.Value{eksObj})

					case *awstypes.ExecutionBlockConfigurationMemberArcRoutingControlConfig:
						arcAttrs := map[string]attr.Value{}

						if config.Value.CrossAccountRole != nil {
							arcAttrs["cross_account_role"] = types.StringValue(aws.ToString(config.Value.CrossAccountRole))
						} else {
							arcAttrs["cross_account_role"] = types.StringNull()
						}

						if config.Value.ExternalId != nil {
							arcAttrs["external_id"] = types.StringValue(aws.ToString(config.Value.ExternalId))
						} else {
							arcAttrs["external_id"] = types.StringNull()
						}

						if config.Value.TimeoutMinutes != nil {
							arcAttrs["timeout_minutes"] = types.Int64Value(int64(aws.ToInt32(config.Value.TimeoutMinutes)))
						} else {
							arcAttrs["timeout_minutes"] = types.Int64Null()
						}

						// Handle Region and Routing Controls (map structure) - sort for consistent ordering
						if len(config.Value.RegionAndRoutingControls) > 0 {
							regionElements := make([]attr.Value, 0)

							// Sort regions for consistent ordering
							regions := make([]string, 0, len(config.Value.RegionAndRoutingControls))
							for region := range config.Value.RegionAndRoutingControls {
								regions = append(regions, region)
							}
							sort.Strings(regions)

							for _, region := range regions {
								routingControlStates := config.Value.RegionAndRoutingControls[region]
								regionAttrs := map[string]attr.Value{
									"region": types.StringValue(region),
								}

								// Extract routing control ARNs from the states
								routingControlArns := make([]string, len(routingControlStates))
								for i, state := range routingControlStates {
									routingControlArns[i] = aws.ToString(state.RoutingControlArn)
								}

								if len(routingControlArns) > 0 {
									controlArns, controlArnsDiags := types.ListValueFrom(ctx, types.StringType, routingControlArns)
									diags.Append(controlArnsDiags...)
									regionAttrs["routing_control_arns"] = controlArns
								} else {
									regionAttrs["routing_control_arns"] = types.ListNull(types.StringType)
								}

								regionObj, regionDiags := types.ObjectValue(getRegionAndRoutingControlsObjectType().AttrTypes, regionAttrs)
								diags.Append(regionDiags...)
								regionElements = append(regionElements, regionObj)
							}
							arcAttrs["region_and_routing_controls"] = types.ListValueMust(getRegionAndRoutingControlsObjectType(), regionElements)
						} else {
							arcAttrs["region_and_routing_controls"] = types.ListNull(getRegionAndRoutingControlsObjectType())
						}

						arcObj, arcDiags := types.ObjectValue(getArcRoutingControlConfigObjectType().AttrTypes, arcAttrs)
						diags.Append(arcDiags...)
						stepAttrs["arc_routing_control_config"] = types.ListValueMust(getArcRoutingControlConfigObjectType(), []attr.Value{arcObj})
					}
				} else {
					// No execution block configuration - all configs are null (already set above)
					stepAttrs["execution_approval_config"] = types.ListNull(getExecutionApprovalConfigObjectType())
					stepAttrs["route53_health_check_config"] = types.ListNull(getRoute53HealthCheckConfigObjectType())
					stepAttrs["custom_action_lambda_config"] = types.ListNull(getCustomActionLambdaConfigObjectType())
					stepAttrs["global_aurora_config"] = types.ListNull(getGlobalAuroraConfigObjectType())
					stepAttrs["ec2_asg_capacity_increase_config"] = types.ListNull(getEc2AsgCapacityIncreaseConfigObjectType())
					stepAttrs["ecs_capacity_increase_config"] = types.ListNull(getEcsCapacityIncreaseConfigObjectType())
					stepAttrs["eks_resource_scaling_config"] = types.ListNull(getEksResourceScalingConfigObjectType())
					stepAttrs["arc_routing_control_config"] = types.ListNull(getArcRoutingControlConfigObjectType())
					stepAttrs["parallel_config"] = types.ListNull(getParallelConfigObjectType())
				}

				stepObj, stepDiags := types.ObjectValue(getStepObjectType().AttrTypes, stepAttrs)
				diags.Append(stepDiags...)
				stepElements[j] = stepObj
			}

			stepsList, stepsDiags := types.ListValue(getStepObjectType(), stepElements)
			diags.Append(stepsDiags...)
			workflowAttrs["step"] = stepsList
		} else {
			workflowAttrs["step"] = types.ListNull(getStepObjectType())
		}

		workflowObj, objDiags := types.ObjectValue(getWorkflowObjectType().AttrTypes, workflowAttrs)
		diags.Append(objDiags...)
		elements[i] = workflowObj
	}

	result, resultDiags := types.ListValue(getWorkflowObjectType(), elements)
	diags.Append(resultDiags...)
	return result, diags
}

func getAssociatedAlarmObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                types.StringType,
			"alarm_type":          types.StringType,
			"resource_identifier": types.StringType,
			"cross_account_role":  types.StringType,
			"external_id":         types.StringType,
		},
	}
}

func getWorkflowObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"workflow_target_action": types.StringType,
			"workflow_target_region": types.StringType,
			"workflow_description":   types.StringType,
			"step":                   types.ListType{ElemType: getStepObjectType()},
		},
	}
}

func getStepObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                             types.StringType,
			"execution_block_type":             types.StringType,
			"description":                      types.StringType,
			"execution_approval_config":        types.ListType{ElemType: getExecutionApprovalConfigObjectType()},
			"route53_health_check_config":      types.ListType{ElemType: getRoute53HealthCheckConfigObjectType()},
			"custom_action_lambda_config":      types.ListType{ElemType: getCustomActionLambdaConfigObjectType()},
			"global_aurora_config":             types.ListType{ElemType: getGlobalAuroraConfigObjectType()},
			"ec2_asg_capacity_increase_config": types.ListType{ElemType: getEc2AsgCapacityIncreaseConfigObjectType()},
			"ecs_capacity_increase_config":     types.ListType{ElemType: getEcsCapacityIncreaseConfigObjectType()},
			"eks_resource_scaling_config":      types.ListType{ElemType: getEksResourceScalingConfigObjectType()},
			"arc_routing_control_config":       types.ListType{ElemType: getArcRoutingControlConfigObjectType()},
			"parallel_config":                  types.ListType{ElemType: getParallelConfigObjectType()},
		},
	}
}

func getRoute53HealthCheckConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"hosted_zone_id":     types.StringType,
			"record_name":        types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
			"timeout_minutes":    types.Int64Type,
			"record_sets":        types.ListType{ElemType: getRecordSetObjectType()},
		},
	}
}

func getRecordSetObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"record_set_identifier": types.StringType,
			"region":                types.StringType,
		},
	}
}

func getExecutionApprovalConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"approval_role":   types.StringType,
			"timeout_minutes": types.Int64Type,
		},
	}
}

func getCustomActionLambdaConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"region_to_run":          types.StringType,
			"retry_interval_minutes": types.Float64Type,
			"timeout_minutes":        types.Int64Type,
			"lambda":                 types.ListType{ElemType: getLambdaObjectType()},
			"ungraceful":             types.ListType{ElemType: getUngracefulObjectType()},
		},
	}
}

func getGlobalAuroraConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"behavior":                  types.StringType,
			"global_cluster_identifier": types.StringType,
			"database_cluster_arns":     types.ListType{ElemType: types.StringType},
			"cross_account_role":        types.StringType,
			"external_id":               types.StringType,
			"timeout_minutes":           types.Int64Type,
			"ungraceful":                types.ListType{ElemType: getGlobalAuroraUngracefulObjectType()},
		},
	}
}

func getGlobalAuroraUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"ungraceful": types.StringType,
		},
	}
}

func getEc2AsgCapacityIncreaseConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"capacity_monitoring_approach": types.StringType,
			"target_percent":               types.Int64Type,
			"timeout_minutes":              types.Int64Type,
			"asgs":                         types.ListType{ElemType: getAsgObjectType()},
			"ungraceful":                   types.ListType{ElemType: getEc2UngracefulObjectType()},
		},
	}
}

func getAsgObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"arn":                types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getEc2UngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"minimum_success_percentage": types.Int64Type,
		},
	}
}

func getEcsCapacityIncreaseConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"capacity_monitoring_approach": types.StringType,
			"target_percent":               types.Int64Type,
			"timeout_minutes":              types.Int64Type,
			"services":                     types.ListType{ElemType: getServiceObjectType()},
			"ungraceful":                   types.ListType{ElemType: getEcsUngracefulObjectType()},
		},
	}
}

func getServiceObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cluster_arn":        types.StringType,
			"service_arn":        types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getEcsUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"minimum_success_percentage": types.Int64Type,
		},
	}
}

func getEksResourceScalingConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"capacity_monitoring_approach": types.StringType,
			"target_percent":               types.Int64Type,
			"timeout_minutes":              types.Int64Type,
			"kubernetes_resource_type":     types.ListType{ElemType: getKubernetesResourceTypeObjectType()},
			"eks_clusters":                 types.ListType{ElemType: getEksClusterObjectType()},
			"scaling_resources":            types.ListType{ElemType: getScalingResourcesObjectType()},
			"ungraceful":                   types.ListType{ElemType: getEksUngracefulObjectType()},
		},
	}
}

func getKubernetesResourceTypeObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"api_version": types.StringType,
			"kind":        types.StringType,
		},
	}
}

func getEksClusterObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cluster_arn":        types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getScalingResourcesObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"namespace": types.StringType,
			"resources": types.ListType{ElemType: getKubernetesScalingResourceObjectType()},
		},
	}
}

func getKubernetesScalingResourceObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"resource_name": types.StringType,
			"name":          types.StringType,
			"namespace":     types.StringType,
			"hpa_name":      types.StringType,
		},
	}
}

func getEksUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"minimum_success_percentage": types.Int64Type,
		},
	}
}

func getArcRoutingControlConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"cross_account_role":          types.StringType,
			"external_id":                 types.StringType,
			"timeout_minutes":             types.Int64Type,
			"region_and_routing_controls": types.ListType{ElemType: getRegionAndRoutingControlsObjectType()},
		},
	}
}

func getRegionAndRoutingControlsObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"region":               types.StringType,
			"routing_control_arns": types.ListType{ElemType: types.StringType},
		},
	}
}

func getParallelConfigObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"step": types.ListType{ElemType: getParallelStepObjectType()},
		},
	}
}

func getParallelStepObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"name":                        types.StringType,
			"execution_block_type":        types.StringType,
			"description":                 types.StringType,
			"execution_approval_config":   types.ListType{ElemType: getExecutionApprovalConfigObjectType()},
			"custom_action_lambda_config": types.ListType{ElemType: getCustomActionLambdaConfigObjectType()},
		},
	}
}

func getLambdaObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"arn":                types.StringType,
			"cross_account_role": types.StringType,
			"external_id":        types.StringType,
		},
	}
}

func getUngracefulObjectType() types.ObjectType {
	return types.ObjectType{
		AttrTypes: map[string]attr.Type{
			"behavior": types.StringType,
		},
	}
}

func (r *resourcePlan) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state resourcePlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	plan, err := FindPlanByARN(ctx, conn, state.ID.ValueString())
	if tfresource.NotFound(err) {
		resp.State.RemoveResource(ctx)
		return
	}
	if err != nil {
		resp.Diagnostics.AddError("reading ARC Region Switch Plan", err.Error())
		return
	}

	state.ARN = types.StringValue(aws.ToString(plan.Arn))
	state.Name = types.StringValue(aws.ToString(plan.Name))
	state.ExecutionRole = types.StringValue(aws.ToString(plan.ExecutionRole))
	state.RecoveryApproach = types.StringValue(string(plan.RecoveryApproach))

	regions, diags := types.ListValueFrom(ctx, types.StringType, plan.Regions)
	resp.Diagnostics.Append(diags...)
	state.Regions = regions

	if plan.Description != nil {
		state.Description = types.StringValue(aws.ToString(plan.Description))
	} else {
		state.Description = types.StringNull()
	}

	if plan.PrimaryRegion != nil {
		state.PrimaryRegion = types.StringValue(aws.ToString(plan.PrimaryRegion))
	} else {
		state.PrimaryRegion = types.StringNull()
	}

	if plan.RecoveryTimeObjectiveMinutes != nil {
		state.RecoveryTimeObjectiveMinutes = types.Int64Value(int64(aws.ToInt32(plan.RecoveryTimeObjectiveMinutes)))
	} else {
		state.RecoveryTimeObjectiveMinutes = types.Int64Null()
	}

	// Handle associated alarms
	if len(plan.AssociatedAlarms) > 0 {
		alarmElements := make([]attr.Value, 0, len(plan.AssociatedAlarms))
		for alarmName, alarm := range plan.AssociatedAlarms {
			alarmAttrs := map[string]attr.Value{
				"name":                types.StringValue(alarmName),
				"alarm_type":          types.StringValue(string(alarm.AlarmType)),
				"resource_identifier": types.StringValue(aws.ToString(alarm.ResourceIdentifier)),
			}

			if alarm.CrossAccountRole != nil {
				alarmAttrs["cross_account_role"] = types.StringValue(aws.ToString(alarm.CrossAccountRole))
			} else {
				alarmAttrs["cross_account_role"] = types.StringNull()
			}

			if alarm.ExternalId != nil {
				alarmAttrs["external_id"] = types.StringValue(aws.ToString(alarm.ExternalId))
			} else {
				alarmAttrs["external_id"] = types.StringNull()
			}

			alarmObj, alarmDiags := types.ObjectValue(getAssociatedAlarmObjectType().AttrTypes, alarmAttrs)
			resp.Diagnostics.Append(alarmDiags...)
			alarmElements = append(alarmElements, alarmObj)
		}

		alarmSet, alarmSetDiags := types.SetValue(getAssociatedAlarmObjectType(), alarmElements)
		resp.Diagnostics.Append(alarmSetDiags...)
		state.AssociatedAlarms = alarmSet
	} else {
		state.AssociatedAlarms = types.SetNull(getAssociatedAlarmObjectType())
	}

	// Handle workflows
	if len(plan.Workflows) > 0 {
		workflows, diags := flattenWorkflowsToFramework(ctx, plan.Workflows)
		resp.Diagnostics.Append(diags...)
		state.Workflow = workflows
	} else {
		state.Workflow = types.ListNull(getWorkflowObjectType())
	}

	// Handle tags
	tags, err := ListTags(ctx, r.Meta().ARCRegionSwitchClient(ctx), state.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("listing tags for ARC Region Switch Plan", err.Error())
		return
	}
	setTagsOut(ctx, tags.Map())

	resp.Diagnostics.Append(resp.State.Set(ctx, state)...)
}

func (r *resourcePlan) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan, state resourcePlanModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	// Convert workflows from Framework List to slice
	var workflows []workflowModel
	resp.Diagnostics.Append(plan.Workflow.ElementsAs(ctx, &workflows, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Convert associated alarms from Framework Set to slice
	var alarms []associatedAlarmModel
	resp.Diagnostics.Append(plan.AssociatedAlarms.ElementsAs(ctx, &alarms, false)...)
	if resp.Diagnostics.HasError() {
		return
	}

	input := arcregionswitch.UpdatePlanInput{
		Arn:              aws.String(state.ID.ValueString()),
		ExecutionRole:    aws.String(plan.ExecutionRole.ValueString()),
		Workflows:        expandWorkflowsFromFramework(workflows),
		AssociatedAlarms: expandAssociatedAlarmsFromFramework(alarms),
	}

	if !plan.Description.Equal(state.Description) {
		if plan.Description.IsNull() {
			input.Description = aws.String("")
		} else {
			input.Description = aws.String(plan.Description.ValueString())
		}
	}

	if !plan.RecoveryTimeObjectiveMinutes.Equal(state.RecoveryTimeObjectiveMinutes) {
		if plan.RecoveryTimeObjectiveMinutes.IsNull() {
			input.RecoveryTimeObjectiveMinutes = aws.Int32(0)
		} else {
			input.RecoveryTimeObjectiveMinutes = aws.Int32(int32(plan.RecoveryTimeObjectiveMinutes.ValueInt64()))
		}
	}

	_, err := conn.UpdatePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("updating ARC Region Switch Plan", err.Error())
		return
	}

	// Handle tags update
	if !plan.TagsAll.Equal(state.TagsAll) {
		if err := UpdateTags(ctx, conn, plan.ID.ValueString(), state.TagsAll, plan.TagsAll); err != nil {
			resp.Diagnostics.AddError("updating tags", err.Error())
			return
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, plan)...)
}

func (r *resourcePlan) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state resourcePlanModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	conn := r.Meta().ARCRegionSwitchClient(ctx)

	input := arcregionswitch.DeletePlanInput{
		Arn: aws.String(state.ID.ValueString()),
	}

	_, err := conn.DeletePlan(ctx, &input)
	if err != nil {
		resp.Diagnostics.AddError("deleting ARC Region Switch Plan", err.Error())
		return
	}
}

func (r *resourcePlan) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}

func (r *resourcePlan) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	// Basic validation is handled by the schema validators
}

func (r *resourcePlan) Metadata(_ context.Context, request resource.MetadataRequest, response *resource.MetadataResponse) {
	response.TypeName = request.ProviderTypeName + "_arcregionswitch_plan"
}

func (r *resourcePlan) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = fwschema.Schema{
		Attributes: map[string]fwschema.Attribute{
			names.AttrARN: framework.ARNAttributeComputedOnly(),
			names.AttrID:  framework.IDAttribute(),
			names.AttrName: fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"execution_role": fwschema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					fwvalidators.ARN(),
				},
			},
			"recovery_approach": fwschema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				Validators: []validator.String{
					stringvalidator.OneOf("activeActive", "activePassive"),
				},
			},
			"regions": fwschema.ListAttribute{
				Required:    true,
				ElementType: types.StringType,
				PlanModifiers: []planmodifier.List{
					listplanmodifier.RequiresReplace(),
				},
			},
			names.AttrDescription: fwschema.StringAttribute{
				Optional: true,
			},
			"primary_region": fwschema.StringAttribute{
				Optional: true,
			},
			"recovery_time_objective_minutes": fwschema.Int64Attribute{
				Optional: true,
			},
			names.AttrTags:    tftags.TagsAttribute(),
			names.AttrTagsAll: tftags.TagsAttributeComputedOnly(),
		},
		Blocks: map[string]fwschema.Block{
			"associated_alarms": fwschema.SetNestedBlock{
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						names.AttrName: fwschema.StringAttribute{
							Required: true,
						},
						"alarm_type": fwschema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("applicationHealth", "trigger"),
							},
						},
						"resource_identifier": fwschema.StringAttribute{
							Required: true,
						},
						"cross_account_role": fwschema.StringAttribute{
							Optional: true,
						},
						"external_id": fwschema.StringAttribute{
							Optional: true,
						},
					},
				},
			},
			"workflow": fwschema.ListNestedBlock{
				NestedObject: fwschema.NestedBlockObject{
					Attributes: map[string]fwschema.Attribute{
						"workflow_target_action": fwschema.StringAttribute{
							Required: true,
							Validators: []validator.String{
								stringvalidator.OneOf("activate", "deactivate"),
							},
						},
						"workflow_target_region": fwschema.StringAttribute{
							Optional: true,
						},
						"workflow_description": fwschema.StringAttribute{
							Optional: true,
						},
					},
					Blocks: map[string]fwschema.Block{
						"step": fwschema.ListNestedBlock{
							NestedObject: fwschema.NestedBlockObject{
								Attributes: map[string]fwschema.Attribute{
									names.AttrName: fwschema.StringAttribute{
										Required: true,
									},
									"execution_block_type": fwschema.StringAttribute{
										Required: true,
									},
									names.AttrDescription: fwschema.StringAttribute{
										Optional: true,
									},
								},
								Blocks: map[string]fwschema.Block{
									"execution_approval_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"approval_role": fwschema.StringAttribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
										},
									},
									"route53_health_check_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"hosted_zone_id": fwschema.StringAttribute{
													Required: true,
												},
												"record_name": fwschema.StringAttribute{
													Required: true,
												},
												"cross_account_role": fwschema.StringAttribute{
													Optional: true,
												},
												"external_id": fwschema.StringAttribute{
													Optional: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"record_sets": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"record_set_identifier": fwschema.StringAttribute{
																Required: true,
															},
															"region": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"custom_action_lambda_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"region_to_run": fwschema.StringAttribute{
													Required: true,
												},
												"retry_interval_minutes": fwschema.Float64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"lambda": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrARN: fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"behavior": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"global_aurora_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"behavior": fwschema.StringAttribute{
													Required: true,
												},
												"global_cluster_identifier": fwschema.StringAttribute{
													Required: true,
												},
												"database_cluster_arns": fwschema.ListAttribute{
													ElementType: types.StringType,
													Required:    true,
												},
												"cross_account_role": fwschema.StringAttribute{
													Optional: true,
												},
												names.AttrExternalID: fwschema.StringAttribute{
													Optional: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"ungraceful": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"ec2_asg_capacity_increase_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"capacity_monitoring_approach": fwschema.StringAttribute{
													Required: true,
												},
												"target_percent": fwschema.Int64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"asgs": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrARN: fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"minimum_success_percentage": fwschema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"ecs_capacity_increase_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"capacity_monitoring_approach": fwschema.StringAttribute{
													Required: true,
												},
												"target_percent": fwschema.Int64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"services": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"cluster_arn": fwschema.StringAttribute{
																Required: true,
															},
															"service_arn": fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"minimum_success_percentage": fwschema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"eks_resource_scaling_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"capacity_monitoring_approach": fwschema.StringAttribute{
													Required: true,
												},
												"target_percent": fwschema.Int64Attribute{
													Required: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"kubernetes_resource_type": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"api_version": fwschema.StringAttribute{
																Required: true,
															},
															"kind": fwschema.StringAttribute{
																Required: true,
															},
														},
													},
												},
												"eks_clusters": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"cluster_arn": fwschema.StringAttribute{
																Required: true,
															},
															"cross_account_role": fwschema.StringAttribute{
																Optional: true,
															},
															names.AttrExternalID: fwschema.StringAttribute{
																Optional: true,
															},
														},
													},
												},
												"scaling_resources": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrNamespace: fwschema.StringAttribute{
																Required: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"resources": fwschema.ListNestedBlock{
																NestedObject: fwschema.NestedBlockObject{
																	Attributes: map[string]fwschema.Attribute{
																		"resource_name": fwschema.StringAttribute{
																			Required: true,
																		},
																		names.AttrName: fwschema.StringAttribute{
																			Required: true,
																		},
																		names.AttrNamespace: fwschema.StringAttribute{
																			Required: true,
																		},
																		"hpa_name": fwschema.StringAttribute{
																			Optional: true,
																		},
																	},
																},
															},
														},
													},
												},
												"ungraceful": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"minimum_success_percentage": fwschema.Int64Attribute{
																Required: true,
															},
														},
													},
												},
											},
										},
									},
									"arc_routing_control_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Attributes: map[string]fwschema.Attribute{
												"cross_account_role": fwschema.StringAttribute{
													Optional: true,
												},
												names.AttrExternalID: fwschema.StringAttribute{
													Optional: true,
												},
												"timeout_minutes": fwschema.Int64Attribute{
													Optional: true,
												},
											},
											Blocks: map[string]fwschema.Block{
												"region_and_routing_controls": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															"region": fwschema.StringAttribute{
																Required: true,
															},
															"routing_control_arns": fwschema.ListAttribute{
																ElementType: types.StringType,
																Required:    true,
															},
														},
													},
												},
											},
										},
									},
									"parallel_config": fwschema.ListNestedBlock{
										NestedObject: fwschema.NestedBlockObject{
											Blocks: map[string]fwschema.Block{
												"step": fwschema.ListNestedBlock{
													NestedObject: fwschema.NestedBlockObject{
														Attributes: map[string]fwschema.Attribute{
															names.AttrName: fwschema.StringAttribute{
																Required: true,
															},
															"execution_block_type": fwschema.StringAttribute{
																Required: true,
															},
															names.AttrDescription: fwschema.StringAttribute{
																Optional: true,
															},
														},
														Blocks: map[string]fwschema.Block{
															"execution_approval_config": fwschema.ListNestedBlock{
																NestedObject: fwschema.NestedBlockObject{
																	Attributes: map[string]fwschema.Attribute{
																		"approval_role": fwschema.StringAttribute{
																			Required: true,
																		},
																		"timeout_minutes": fwschema.Int64Attribute{
																			Optional: true,
																		},
																	},
																},
															},
															"custom_action_lambda_config": fwschema.ListNestedBlock{
																NestedObject: fwschema.NestedBlockObject{
																	Attributes: map[string]fwschema.Attribute{
																		"region_to_run": fwschema.StringAttribute{
																			Required: true,
																		},
																		"retry_interval_minutes": fwschema.Float64Attribute{
																			Required: true,
																		},
																		"timeout_minutes": fwschema.Int64Attribute{
																			Optional: true,
																		},
																	},
																	Blocks: map[string]fwschema.Block{
																		"lambda": fwschema.ListNestedBlock{
																			NestedObject: fwschema.NestedBlockObject{
																				Attributes: map[string]fwschema.Attribute{
																					names.AttrARN: fwschema.StringAttribute{
																						Required: true,
																					},
																					"cross_account_role": fwschema.StringAttribute{
																						Optional: true,
																					},
																					names.AttrExternalID: fwschema.StringAttribute{
																						Optional: true,
																					},
																				},
																			},
																		},
																		"ungraceful": fwschema.ListNestedBlock{
																			NestedObject: fwschema.NestedBlockObject{
																				Attributes: map[string]fwschema.Attribute{
																					"behavior": fwschema.StringAttribute{
																						Required: true,
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
	}
}
