package aws

import (
	"log"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/greengrass"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

func resourceAwsGreengrassGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsGreengrassGroupCreate,
		Read:   resourceAwsGreengrassGroupRead,
		Update: resourceAwsGreengrassGroupUpdate,
		Delete: resourceAwsGreengrassGroupDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"arn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"group_version": {
				Type:     schema.TypeSet,
				MaxItems: 1,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"connector_definition_version_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"core_definition_version_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"device_definition_version_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"function_definition_version_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"logger_definition_version_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"resource_definition_version_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"subscription_definition_version_arn": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func createGroupVersion(d *schema.ResourceData, conn *greengrass.Greengrass) error {
	var raw map[string]interface{}
	if v := d.Get("group_version").(*schema.Set).List(); len(v) != 0 {
		raw = v[0].(map[string]interface{})
	} else {
		return nil
	}

	params := &greengrass.CreateGroupVersionInput{
		GroupId: aws.String(d.Id()),
	}

	if v := os.Getenv("AMZN_CLIENT_TOKEN"); v != "" {
		params.AmznClientToken = aws.String(v)
	}

	if v, ok := raw["connector_definition_version_arn"]; ok {
		params.ConnectorDefinitionVersionArn = aws.String(v.(string))
	}

	if v, ok := raw["core_definition_version_arn"]; ok {
		params.CoreDefinitionVersionArn = aws.String(v.(string))
	}

	if v, ok := raw["device_definition_version_arn"]; ok {
		params.DeviceDefinitionVersionArn = aws.String(v.(string))
	}

	if v, ok := raw["function_definition_version_arn"]; ok {
		params.FunctionDefinitionVersionArn = aws.String(v.(string))
	}

	if v, ok := raw["logger_definition_version_arn"]; ok {
		params.LoggerDefinitionVersionArn = aws.String(v.(string))
	}

	if v, ok := raw["resource_definition_version_arn"]; ok {
		params.ResourceDefinitionVersionArn = aws.String(v.(string))
	}

	if v, ok := raw["subscription_definition_version_arn"]; ok {
		params.SubscriptionDefinitionVersionArn = aws.String(v.(string))
	}

	_, err := conn.CreateGroupVersion(params)

	if err != nil {
		return err
	}

	return nil
}

func resourceAwsGreengrassGroupCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.CreateGroupInput{
		Name: aws.String(d.Get("name").(string)),
	}

	log.Printf("[DEBUG] Creating Greengrass Group: %s", params)
	out, err := conn.CreateGroup(params)
	if err != nil {
		return err
	}

	d.SetId(*out.Id)

	err = createGroupVersion(d, conn)

	if err != nil {
		return err
	}

	return resourceAwsGreengrassGroupRead(d, meta)
}

func readGroupVersion(latestVersion string, d *schema.ResourceData, conn *greengrass.Greengrass) error {
	params := &greengrass.GetGroupVersionInput{
		GroupId:        aws.String(d.Id()),
		GroupVersionId: aws.String(latestVersion),
	}
	log.Printf("[DEBUG] Reading Greengrass Group Version: %s", params)
	out, err := conn.GetGroupVersion(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Group Version: %s", out)

	flattenGroupVersion := make(map[string]interface{})

	flattenGroupVersion["connector_definition_version_arn"] = out.Definition.ConnectorDefinitionVersionArn
	flattenGroupVersion["core_definition_version_arn"] = out.Definition.CoreDefinitionVersionArn
	flattenGroupVersion["device_definition_version_arn"] = out.Definition.DeviceDefinitionVersionArn
	flattenGroupVersion["function_definition_version_arn"] = out.Definition.FunctionDefinitionVersionArn
	flattenGroupVersion["logger_definition_version_arn"] = out.Definition.LoggerDefinitionVersionArn
	flattenGroupVersion["resource_definition_version_arn"] = out.Definition.ResourceDefinitionVersionArn
	flattenGroupVersion["subscription_definition_version_arn"] = out.Definition.SubscriptionDefinitionVersionArn

	d.Set("group_version", []map[string]interface{}{flattenGroupVersion})
	return nil
}

func resourceAwsGreengrassGroupRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.GetGroupInput{
		GroupId: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Reading Greengrass Group: %s", params)
	out, err := conn.GetGroup(params)

	if err != nil {
		return err
	}

	log.Printf("[DEBUG] Received Greengrass Group: %s", out)

	d.Set("arn", out.Arn)
	d.Set("name", out.Name)
	d.Set("group_id", out.Id)

	if out.LatestVersion != nil {
		err = readGroupVersion(*out.LatestVersion, d, conn)
	}
	if err != nil {
		return err
	}
	return nil
}

func resourceAwsGreengrassGroupUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.UpdateGroupInput{
		Name:    aws.String(d.Get("name").(string)),
		GroupId: aws.String(d.Get("group_id").(string)),
	}

	_, err := conn.UpdateGroup(params)
	if err != nil {
		return err
	}

	if d.HasChange("group_version") {
		err = createGroupVersion(d, conn)
		if err != nil {
			return err
		}
	}

	return resourceAwsGreengrassGroupRead(d, meta)
}

func resourceAwsGreengrassGroupDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).greengrassconn

	params := &greengrass.DeleteGroupInput{
		GroupId: aws.String(d.Get("group_id").(string)),
	}
	log.Printf("[DEBUG] Deleting Greengrass Group: %s", params)

	_, err := conn.DeleteGroup(params)
	if err != nil {
		return nil
	}

	return nil
}
