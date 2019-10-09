package aws

import (
	"log"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/helper/validation"
)

func generateVariableSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"string_value": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"double_value": {
				Type:     schema.TypeFloat,
				Optional: true,
			},
			"dataset_content_version_value": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"dataset_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
			"output_file_uri_value": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"file_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateContainerDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"image": {
				Type:     schema.TypeString,
				Required: true,
			},
			"execution_role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"resource_configuration": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"compute_type": {
							Type:     schema.TypeString,
							Required: true,
						},
						"volume_size_in_gb": {
							Type:     schema.TypeInt,
							Required: true,
						},
					},
				},
			},
			"variable": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateVariableSchema(),
			},
		},
	}
}

func generateQueryFilterSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"delta_time": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"offset_seconds": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"time_expression": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateSqlQueryDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"sql_query": {
				Type:     schema.TypeString,
				Required: true,
			},
			"filter": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateQueryFilterSchema(),
			},
		},
	}
}

func generateDatasetActionSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"query_action": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateSqlQueryDatasetActionSchema(),
			},
		},
	}
}

func generateS3DestinationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key": {
				Type:     schema.TypeString,
				Required: true,
			},
			"role_arn": {
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validateArn,
			},
			"glue_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"database_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"table_name": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateDatasetContentDeliveryDestinationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"iotevents_destination": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"input_name": {
							Type:     schema.TypeString,
							Required: true,
						},
						"role_arn": {
							Type:         schema.TypeString,
							Required:     true,
							ValidateFunc: validateArn,
						},
					},
				},
			},
			"s3_destination": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateS3DestinationSchema(),
			},
		},
	}
}

func generateDatasetContentDeliveryRuleSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"entry_name": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"destination": {
				Type:     schema.TypeSet,
				Required: true,
				MaxItems: 1,
				Elem:     generateDatasetContentDeliveryDestinationSchema(),
			},
		},
	}
}

func generateDatasetTriggerSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"schedule": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"expression": {
							Type:     schema.TypeString,
							Required: true,
						},
					},
				},
			},
		},
	}
}

func generateVersioningConfigurationSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"max_versions": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"versioning_configuration.0.unlimited"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"unlimited": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"versioning_configuration.0.max_versions"},
			},
		},
	}
}

func resourceAwsIotAnalyticsDataset() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsDatasetCreate,
		Read:   resourceAwsIotAnalyticsDatasetRead,
		Update: resourceAwsIotAnalyticsDatasetUpdate,
		Delete: resourceAwsIotAnalyticsDatasetDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"action": {
				Type:     schema.TypeSet,
				Required: true,
				MinItems: 1,
				Elem:     generateDatasetActionSchema(),
			},
			"content_delivery_rule": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem:     generateDatasetContentDeliveryRuleSchema(),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateRetentionPeriodSchema(),
			},
			"trigger": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 5,
				Elem:     generateDatasetTriggerSchema(),
			},
			"versioning_configuration": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateVersioningConfigurationSchema(),
			},
		},
	}
}

func parseVariable(rawVariable map[string]interface{}) *iotanalytics.Variable {
	variable := &iotanalytics.Variable{
		Name: aws.String(rawVariable["name"].(string)),
	}

	if v, ok := rawVariable["string_value"]; ok {
		variable.StringValue = aws.String(v.(string))
	}

	if v, ok := rawVariable["double_value"]; ok {
		variable.DoubleValue = aws.Float64(v.(float64))
	}

	rawDatasetContentVersionValueSet := rawVariable["dataset_content_version_value"].(*schema.Set).List()
	if len(rawDatasetContentVersionValueSet) > 0 {
		rawDatasetContentVersionValue := rawDatasetContentVersionValueSet[0].(map[string]interface{})
		datasetContentVersionValue := &iotanalytics.DatasetContentVersionValue{
			DatasetName: aws.String(rawDatasetContentVersionValue["dataset_name"].(string)),
		}
		variable.DatasetContentVersionValue = datasetContentVersionValue
	}

	rawOutputFileUriValueSet := rawVariable["output_file_uri_value"].(*schema.Set).List()
	if len(rawOutputFileUriValueSet) > 0 {
		rawOutputFileUriValue := rawOutputFileUriValueSet[0].(map[string]interface{})
		outputFileUriValue := &iotanalytics.OutputFileUriValue{
			FileName: aws.String(rawOutputFileUriValue["file_name"].(string)),
		}
		variable.OutputFileUriValue = outputFileUriValue
	}

	return variable
}

func parseContainerAction(rawContainerAction map[string]interface{}) *iotanalytics.ContainerDatasetAction {
	containerAction := &iotanalytics.ContainerDatasetAction{
		Image:            aws.String(rawContainerAction["image"].(string)),
		ExecutionRoleArn: aws.String(rawContainerAction["execution_role_arn"].(string)),
	}

	rawResourceConfiguration := rawContainerAction["resource_configuration"].(*schema.Set).List()[0].(map[string]interface{})
	containerAction.ResourceConfiguration = &iotanalytics.ResourceConfiguration{
		ComputeType:    aws.String(rawResourceConfiguration["compute_type"].(string)),
		VolumeSizeInGB: aws.Int64(int64(rawResourceConfiguration["volume_size_in_gb"].(int))),
	}

	variables := make([]*iotanalytics.Variable, 0)
	rawVariables := rawContainerAction["variable"].(*schema.Set).List()
	for _, rawVar := range rawVariables {
		variable := parseVariable(rawVar.(map[string]interface{}))
		variables = append(variables, variable)
	}
	containerAction.Variables = variables

	return containerAction
}

func parseQueryFilter(rawQueryFilter map[string]interface{}) *iotanalytics.QueryFilter {
	rawDeltaTime := rawQueryFilter["delta_time"].(*schema.Set).List()[0].(map[string]interface{})
	deltaTime := &iotanalytics.DeltaTime{
		OffsetSeconds:  aws.Int64(int64(rawDeltaTime["offset_seconds"].(int))),
		TimeExpression: aws.String(rawDeltaTime["time_expression"].(string)),
	}
	queryFilter := &iotanalytics.QueryFilter{
		DeltaTime: deltaTime,
	}
	return queryFilter
}

func parseSqlQueryAction(rawSqlQueryAction map[string]interface{}) *iotanalytics.SqlQueryDatasetAction {
	sqlQueryAction := &iotanalytics.SqlQueryDatasetAction{
		SqlQuery: aws.String(rawSqlQueryAction["sql_query"].(string)),
	}

	filters := make([]*iotanalytics.QueryFilter, 0)
	rawFilters := rawSqlQueryAction["filter"].(*schema.Set).List()
	for _, rawFilter := range rawFilters {
		filter := parseQueryFilter(rawFilter.(map[string]interface{}))
		filters = append(filters, filter)
	}
	sqlQueryAction.Filters = filters
	return sqlQueryAction

}

func parseDatasetAction(rawAction map[string]interface{}) *iotanalytics.DatasetAction {
	action := &iotanalytics.DatasetAction{
		ActionName: aws.String(rawAction["name"].(string)),
	}

	rawQueryActionSet := rawAction["query_action"].(*schema.Set).List()
	if len(rawQueryActionSet) > 0 {
		rawQueryAction := rawQueryActionSet[0].(map[string]interface{})
		action.QueryAction = parseSqlQueryAction(rawQueryAction)
	}

	return action
}

func parseS3Destination(rawS3Destination map[string]interface{}) *iotanalytics.S3DestinationConfiguration {
	s3Destination := &iotanalytics.S3DestinationConfiguration{
		Bucket:  aws.String(rawS3Destination["bucket"].(string)),
		Key:     aws.String(rawS3Destination["key"].(string)),
		RoleArn: aws.String(rawS3Destination["role_arn"].(string)),
	}

	rawGlueConfigurationSet := rawS3Destination["glue_configuration"].(*schema.Set).List()
	if len(rawGlueConfigurationSet) > 0 {
		rawGlueConfiguration := rawGlueConfigurationSet[0].(map[string]interface{})
		s3Destination.GlueConfiguration = &iotanalytics.GlueConfiguration{
			DatabaseName: aws.String(rawGlueConfiguration["database_name"].(string)),
			TableName:    aws.String(rawGlueConfiguration["table_name"].(string)),
		}
	}

	return s3Destination
}

func parseIotEventsDestination(rawIotEventsDestination map[string]interface{}) *iotanalytics.IotEventsDestinationConfiguration {
	return &iotanalytics.IotEventsDestinationConfiguration{
		InputName: aws.String(rawIotEventsDestination["input_name"].(string)),
		RoleArn:   aws.String(rawIotEventsDestination["role_arn"].(string)),
	}
}

func parseDestination(rawDestination map[string]interface{}) *iotanalytics.DatasetContentDeliveryDestination {
	destination := &iotanalytics.DatasetContentDeliveryDestination{}

	rawIotEventsDestinationSet := rawDestination["iotevents_destination"].(*schema.Set).List()
	if len(rawIotEventsDestinationSet) > 0 {
		rawIotEventsDestination := rawIotEventsDestinationSet[0].(map[string]interface{})
		destination.IotEventsDestinationConfiguration = parseIotEventsDestination(rawIotEventsDestination)
	}

	rawS3DestinationSet := rawDestination["s3_destination"].(*schema.Set).List()
	if len(rawS3DestinationSet) > 0 {
		rawS3Destination := rawS3DestinationSet[0].(map[string]interface{})
		destination.S3DestinationConfiguration = parseS3Destination(rawS3Destination)
	}

	return destination
}

func parseContentDeliveryRule(rawContentDeliveryRule map[string]interface{}) *iotanalytics.DatasetContentDeliveryRule {
	rawDestination := rawContentDeliveryRule["destination"].(*schema.Set).List()[0].(map[string]interface{})
	datasetContentDeliveryRule := &iotanalytics.DatasetContentDeliveryRule{
		Destination: parseDestination(rawDestination),
	}

	if rawEntryName, ok := rawContentDeliveryRule["entry_name"]; ok {
		datasetContentDeliveryRule.EntryName = aws.String(rawEntryName.(string))
	}

	return datasetContentDeliveryRule
}

func parseTrigger(rawTrigger map[string]interface{}) *iotanalytics.DatasetTrigger {
	trigger := &iotanalytics.DatasetTrigger{}

	rawScheduleSet := rawTrigger["schedule"].(*schema.Set).List()
	if len(rawScheduleSet) > 0 {
		rawSchedule := rawScheduleSet[0].(map[string]interface{})
		trigger.Schedule = &iotanalytics.Schedule{
			Expression: aws.String(rawSchedule["expression"].(string)),
		}
	}

	return trigger
}

func parseVersioningConfiguration(rawVersioningConfiguration map[string]interface{}) *iotanalytics.VersioningConfiguration {
	var maxVersion *int64
	if v, ok := rawVersioningConfiguration["max_versions"]; ok && int64(v.(int)) > 1 {
		maxVersion = aws.Int64(int64(v.(int)))
	}
	var unlimited *bool
	if v, ok := rawVersioningConfiguration["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.VersioningConfiguration{
		MaxVersions: maxVersion,
		Unlimited:   unlimited,
	}
}

func resourceAwsIotAnalyticsDatasetCreate(d *schema.ResourceData, meta interface{}) error {
	// TODO: make function that return structure of ready-to-use fields to fill
	// CreateDatasetInput and UpdateDatasetInput structures
	conn := meta.(*AWSClient).iotanalyticsconn

	name := d.Get("name").(string)
	params := &iotanalytics.CreateDatasetInput{
		DatasetName: aws.String(name),
	}

	rawActions := d.Get("action").(*schema.Set).List()
	actions := make([]*iotanalytics.DatasetAction, 0)
	for _, rawAction := range rawActions {
		action := parseDatasetAction(rawAction.(map[string]interface{}))
		actions = append(actions, action)
	}
	params.Actions = actions

	rawContentDeliveryRules := d.Get("content_delivery_rule").(*schema.Set).List()
	contentDeliveryRules := make([]*iotanalytics.DatasetContentDeliveryRule, 0)
	for _, rawRule := range rawContentDeliveryRules {
		rule := parseContentDeliveryRule(rawRule.(map[string]interface{}))
		contentDeliveryRules = append(contentDeliveryRules, rule)
	}
	params.ContentDeliveryRules = contentDeliveryRules

	rawTriggers := d.Get("trigger").(*schema.Set).List()
	triggers := make([]*iotanalytics.DatasetTrigger, 0)
	for _, rawTrigger := range rawTriggers {
		trigger := parseTrigger(rawTrigger.(map[string]interface{}))
		triggers = append(triggers, trigger)
	}
	params.Triggers = triggers

	rawRetentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(rawRetentionPeriodSet) > 0 {
		rawRetentionPeriod := rawRetentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	rawVersioningConfigurationSet := d.Get("versioning_configuration").(*schema.Set).List()
	if len(rawVersioningConfigurationSet) > 0 {
		rawVersioningConfiguration := rawVersioningConfigurationSet[0].(map[string]interface{})
		params.VersioningConfiguration = parseVersioningConfiguration(rawVersioningConfiguration)
	}

	log.Printf("[DEBUG] Creating IoT Analytics Dataset: %s", params)
	_, err := conn.CreateDataset(params)

	if err != nil {
		return err
	}

	d.SetId(name)

	return resourceAwsIotAnalyticsDatasetRead(d, meta)
}

func flattenVariable(variable *iotanalytics.Variable) map[string]interface{} {
	rawVariable := make(map[string]interface{})
	rawVariable["name"] = aws.StringValue(variable.Name)

	if variable.StringValue != nil {
		rawVariable["string_value"] = aws.StringValue(variable.StringValue)
	}

	if variable.DoubleValue != nil {
		rawVariable["string_value"] = aws.StringValue(variable.StringValue)
	}

	if variable.OutputFileUriValue != nil {
		outputFileUriValue := map[string]interface{}{
			"file_name": aws.StringValue(variable.OutputFileUriValue.FileName),
		}
		rawVariable["output_file_uri_value"] = wrapMapInList(outputFileUriValue)
	}

	if variable.DatasetContentVersionValue != nil {
		datasetContentVersionValue := map[string]interface{}{
			"dataset_name": aws.StringValue(variable.DatasetContentVersionValue.DatasetName),
		}
		rawVariable["dataset_content_version_value"] = wrapMapInList(datasetContentVersionValue)
	}

	return rawVariable
}

func flattenContainerAction(containerAction *iotanalytics.ContainerDatasetAction) map[string]interface{} {
	rawContainerAction := make(map[string]interface{})
	rawContainerAction["image"] = aws.StringValue(containerAction.Image)
	rawContainerAction["execution_role_arn"] = aws.StringValue(containerAction.ExecutionRoleArn)

	rawResourceConfiguration := map[string]interface{}{
		"compute_type":      aws.StringValue(containerAction.ResourceConfiguration.ComputeType),
		"volume_size_in_gb": aws.Int64Value(containerAction.ResourceConfiguration.VolumeSizeInGB),
	}
	rawContainerAction["resource_configuration"] = wrapMapInList(rawResourceConfiguration)

	rawVariables := make([]map[string]interface{}, 0)
	for _, variable := range containerAction.Variables {
		rawVariables = append(rawVariables, flattenVariable(variable))
	}
	rawContainerAction["variable"] = rawVariables

	return rawContainerAction
}

func flattenQueryFilter(queryFilter *iotanalytics.QueryFilter) map[string]interface{} {
	rawDeltaTime := map[string]interface{}{
		"offset_seconds":  aws.Int64Value(queryFilter.DeltaTime.OffsetSeconds),
		"time_expression": aws.StringValue(queryFilter.DeltaTime.TimeExpression),
	}
	rawQueryFilter := make(map[string]interface{})
	rawQueryFilter["delta_time"] = wrapMapInList(rawDeltaTime)
	return rawQueryFilter
}

func flattenSqlQueryAction(sqlQueryAction *iotanalytics.SqlQueryDatasetAction) map[string]interface{} {
	rawSqlQueryAction := make(map[string]interface{})
	rawSqlQueryAction["sql_query"] = aws.StringValue(sqlQueryAction.SqlQuery)

	rawFilters := make([]map[string]interface{}, 0)
	for _, filter := range sqlQueryAction.Filters {
		rawFilters = append(rawFilters, flattenQueryFilter(filter))
	}
	rawSqlQueryAction["filter"] = rawFilters
	return rawSqlQueryAction
}

func flattenDatasetAction(action *iotanalytics.DatasetAction) map[string]interface{} {
	rawAction := make(map[string]interface{})
	rawAction["name"] = aws.StringValue(action.ActionName)

	if action.QueryAction != nil {
		rawQueryAction := flattenSqlQueryAction(action.QueryAction)
		rawAction["query_action"] = wrapMapInList(rawQueryAction)
	}

	return rawAction
}

func flattenS3Destination(s3Destination *iotanalytics.S3DestinationConfiguration) map[string]interface{} {
	rawS3Destination := make(map[string]interface{})
	rawS3Destination["bucket"] = aws.StringValue(s3Destination.Bucket)
	rawS3Destination["key"] = aws.StringValue(s3Destination.Key)
	rawS3Destination["role_arn"] = aws.StringValue(s3Destination.RoleArn)

	if s3Destination.GlueConfiguration != nil {
		rawGlueConfiguration := map[string]interface{}{
			"database_name": aws.StringValue(s3Destination.GlueConfiguration.DatabaseName),
			"table_name":    aws.StringValue(s3Destination.GlueConfiguration.TableName),
		}
		rawS3Destination["glue_configuration"] = wrapMapInList(rawGlueConfiguration)
	}
	return rawS3Destination
}

func flattenIotEventsDestination(iotEventsDestination *iotanalytics.IotEventsDestinationConfiguration) map[string]interface{} {
	rawIotEventsDestination := map[string]interface{}{
		"input_name": aws.StringValue(iotEventsDestination.InputName),
		"role_arn":   aws.StringValue(iotEventsDestination.RoleArn),
	}
	return rawIotEventsDestination
}

func flattenDestination(destination *iotanalytics.DatasetContentDeliveryDestination) map[string]interface{} {
	rawDestination := make(map[string]interface{})

	if destination.IotEventsDestinationConfiguration != nil {
		rawIotEventsDestination := flattenIotEventsDestination(destination.IotEventsDestinationConfiguration)
		rawDestination["iotevents_destination"] = wrapMapInList(rawIotEventsDestination)
	}

	if destination.S3DestinationConfiguration != nil {
		rawS3Destination := flattenS3Destination(destination.S3DestinationConfiguration)
		rawDestination["s3_destination"] = wrapMapInList(rawS3Destination)
	}

	return rawDestination
}

func flattenContentDeliveryRule(datasetContentDeliveryRule *iotanalytics.DatasetContentDeliveryRule) map[string]interface{} {
	rawContentDeliveryRule := make(map[string]interface{})

	rawDestination := flattenDestination(datasetContentDeliveryRule.Destination)
	rawContentDeliveryRule["destination"] = wrapMapInList(rawDestination)

	if datasetContentDeliveryRule.EntryName != nil {
		rawContentDeliveryRule["entry_name"] = aws.StringValue(datasetContentDeliveryRule.EntryName)
	}

	return rawContentDeliveryRule
}

func flattenTrigger(trigger *iotanalytics.DatasetTrigger) map[string]interface{} {
	rawTrigger := make(map[string]interface{})

	if trigger.Schedule != nil {
		rawSchedule := map[string]interface{}{
			"expression": aws.StringValue(trigger.Schedule.Expression),
		}
		rawTrigger["schedule"] = wrapMapInList(rawSchedule)
	}

	return rawTrigger
}

func flattenVersioningConfiguration(versioningConfiguration *iotanalytics.VersioningConfiguration) map[string]interface{} {
	if versioningConfiguration == nil {
		return nil
	}

	rawVersioningConfiguration := make(map[string]interface{})

	if versioningConfiguration.MaxVersions != nil {
		rawVersioningConfiguration["max_versions"] = aws.Int64Value(versioningConfiguration.MaxVersions)
	}
	if versioningConfiguration.Unlimited != nil {
		rawVersioningConfiguration["unlimited"] = aws.BoolValue(versioningConfiguration.Unlimited)
	}

	return rawVersioningConfiguration
}

func resourceAwsIotAnalyticsDatasetRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DescribeDatasetInput{
		DatasetName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading IoT Analytics Dataset: %s", params)
	out, err := conn.DescribeDataset(params)

	if err != nil {
		return err
	}

	d.Set("name", out.Dataset.Name)

	rawActions := make([]map[string]interface{}, 0)
	for _, action := range out.Dataset.Actions {
		rawActions = append(rawActions, flattenDatasetAction(action))
	}
	d.Set("action", rawActions)

	rawContentDeliveryRules := make([]map[string]interface{}, 0)
	for _, rule := range out.Dataset.ContentDeliveryRules {
		rawContentDeliveryRules = append(rawContentDeliveryRules, flattenContentDeliveryRule(rule))
	}
	d.Set("content_delivery_rule", rawContentDeliveryRules)

	rawRetentionPeriod := flattenRetentionPeriod(out.Dataset.RetentionPeriod)
	d.Set("retention_period", wrapMapInList(rawRetentionPeriod))

	rawTriggers := make([]map[string]interface{}, 0)
	for _, trigger := range out.Dataset.Triggers {
		rawTriggers = append(rawTriggers, flattenTrigger(trigger))
	}
	d.Set("trigger", rawTriggers)

	rawVersioningConfiguration := flattenVersioningConfiguration(out.Dataset.VersioningConfiguration)
	d.Set("versioning_configuration", wrapMapInList(rawVersioningConfiguration))
	return nil
}

func resourceAwsIotAnalyticsDatasetUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	name := d.Get("name").(string)
	params := &iotanalytics.UpdateDatasetInput{
		DatasetName: aws.String(name),
	}

	rawActions := d.Get("action").(*schema.Set).List()
	actions := make([]*iotanalytics.DatasetAction, 0)
	for _, rawAction := range rawActions {
		action := parseDatasetAction(rawAction.(map[string]interface{}))
		actions = append(actions, action)
	}
	params.Actions = actions

	rawContentDeliveryRules := d.Get("content_delivery_rule").(*schema.Set).List()
	contentDeliveryRules := make([]*iotanalytics.DatasetContentDeliveryRule, 0)
	for _, rawRule := range rawContentDeliveryRules {
		rule := parseContentDeliveryRule(rawRule.(map[string]interface{}))
		contentDeliveryRules = append(contentDeliveryRules, rule)
	}
	params.ContentDeliveryRules = contentDeliveryRules

	rawTriggers := d.Get("trigger").(*schema.Set).List()
	triggers := make([]*iotanalytics.DatasetTrigger, 0)
	for _, rawTrigger := range rawTriggers {
		trigger := parseTrigger(rawTrigger.(map[string]interface{}))
		triggers = append(triggers, trigger)
	}
	params.Triggers = triggers

	rawRetentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(rawRetentionPeriodSet) > 0 {
		rawRetentionPeriod := rawRetentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	rawVersioningConfigurationSet := d.Get("versioning_configuration").(*schema.Set).List()
	if len(rawVersioningConfigurationSet) > 0 {
		rawVersioningConfiguration := rawVersioningConfigurationSet[0].(map[string]interface{})
		params.VersioningConfiguration = parseVersioningConfiguration(rawVersioningConfiguration)
	}

	log.Printf("[DEBUG] Creating IoT Analytics Dataset: %s", params)
	_, err := conn.UpdateDataset(params)

	if err != nil {
		return err
	}

	return resourceAwsIotAnalyticsDatasetRead(d, meta)

}

func resourceAwsIotAnalyticsDatasetDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeleteDatasetInput{
		DatasetName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Deleting IoT Analytics Dataset: %s", params)
	_, err := conn.DeleteDataset(params)

	return err
}
