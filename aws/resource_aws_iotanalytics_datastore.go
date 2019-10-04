package aws

import (
	"log"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/iotanalytics"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/helper/validation"
)

func generateCustomerManagedS3Schema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"bucket": {
				Type:     schema.TypeString,
				Required: true,
			},
			"key_prefix": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"role_arn": {
				Type:     schema.TypeString,
				Required: true,
			},
		},
	}
}

func generateServiceManagedS3Schema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{},
	}
}

func generateStorageSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"customer_managed_s3": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"storage.0.service_managed_s3"},
				Elem:          generateCustomerManagedS3Schema(),
			},
			"service_managed_s3": {
				Type:          schema.TypeList,
				Optional:      true,
				MaxItems:      1,
				ConflictsWith: []string{"storage.0.customer_managed_s3"},
				Elem:          generateServiceManagedS3Schema(),
			},
		},
	}
}

func generateRetentionPeriodSchema() *schema.Resource {
	return &schema.Resource{
		Schema: map[string]*schema.Schema{
			"number_of_days": {
				Type:          schema.TypeInt,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.unlimited"},
				ValidateFunc:  validation.IntAtLeast(1),
			},
			"unlimited": {
				Type:          schema.TypeBool,
				Optional:      true,
				ConflictsWith: []string{"retention_period.0.number_of_days"},
			},
		},
	}
}

func resourceAwsIotAnalyticsDatastore() *schema.Resource {
	return &schema.Resource{
		Create: resourceAwsIotAnalyticsDatastoreCreate,
		Read:   resourceAwsIotAnalyticsDatastoreRead,
		Update: resourceAwsIotAnalyticsDatastoreUpdate,
		Delete: resourceAwsIotAnalyticsDatastoreDelete,

		Importer: &schema.ResourceImporter{
			State: schema.ImportStatePassthrough,
		},

		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"storage": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateStorageSchema(),
			},
			"retention_period": {
				Type:     schema.TypeSet,
				Optional: true,
				MaxItems: 1,
				Elem:     generateRetentionPeriodSchema(),
			},
		},
	}
}

func parseCustomerManagedS3(rawCustomerManagedS3 map[string]interface{}) *iotanalytics.CustomerManagedDatastoreS3Storage {
	bucket := rawCustomerManagedS3["bucket"].(string)
	roleArn := rawCustomerManagedS3["role_arn"].(string)
	customerManagedS3 := &iotanalytics.CustomerManagedDatastoreS3Storage{
		Bucket:  aws.String(bucket),
		RoleArn: aws.String(roleArn),
	}

	if v, ok := rawCustomerManagedS3["key_prefix"]; ok && len(v.(string)) >= 1 {
		customerManagedS3.KeyPrefix = aws.String(v.(string))
	}

	return customerManagedS3
}

func parseServiceManagedS3(rawServiceManagedS3 map[string]interface{}) *iotanalytics.ServiceManagedDatastoreS3Storage {
	return &iotanalytics.ServiceManagedDatastoreS3Storage{}
}

func parseStorage(rawDatastoreStorage map[string]interface{}) *iotanalytics.DatastoreStorage {

	var customerManagedS3 *iotanalytics.CustomerManagedDatastoreS3Storage
	if list := rawDatastoreStorage["customer_managed_s3"].([]interface{}); len(list) > 0 {
		rawCustomerManagedS3 := list[0].(map[string]interface{})
		customerManagedS3 = parseCustomerManagedS3(rawCustomerManagedS3)
	}

	var serviceManagedS3 *iotanalytics.ServiceManagedDatastoreS3Storage
	if list := rawDatastoreStorage["service_managed_s3"].([]interface{}); len(list) > 0 {
		rawServiceManagedS3 := list[0].(map[string]interface{})
		serviceManagedS3 = parseServiceManagedS3(rawServiceManagedS3)
	}

	return &iotanalytics.DatastoreStorage{
		CustomerManagedS3: customerManagedS3,
		ServiceManagedS3:  serviceManagedS3,
	}
}

func parseRetentionPeriod(rawRetentionPeriod map[string]interface{}) *iotanalytics.RetentionPeriod {

	var numberOfDays *int64
	if v, ok := rawRetentionPeriod["number_of_days"]; ok && int64(v.(int)) > 1 {
		numberOfDays = aws.Int64(int64(v.(int)))
	}
	var unlimited *bool
	if v, ok := rawRetentionPeriod["unlimited"]; ok {
		unlimited = aws.Bool(v.(bool))
	}
	return &iotanalytics.RetentionPeriod{
		NumberOfDays: numberOfDays,
		Unlimited:    unlimited,
	}
}

func resourceAwsIotAnalyticsDatastoreCreate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.CreateDatastoreInput{
		DatastoreName: aws.String(d.Get("name").(string)),
	}

	datastoreStorageSet := d.Get("storage").(*schema.Set).List()
	if len(datastoreStorageSet) >= 1 {
		rawDatastoreStorage := datastoreStorageSet[0].(map[string]interface{})
		params.DatastoreStorage = parseStorage(rawDatastoreStorage)
	}

	retentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(retentionPeriodSet) >= 1 {
		rawRetentionPeriod := retentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	log.Printf("[DEBUG] Create IoTAnalytics Datastore: %s", params)

	retrySecondsList := [6]int{1, 2, 5, 8, 10, 0}

	var err error

	// Primitive retry.
	// During testing datastore, problem was detected.
	// When we try to create datastore model and role arn that
	// will be assumed by datastore during one apply we get:
	// 'Unable to assume role, role ARN' error. However if we run apply
	// second time(when all required resources are created) datastore will be created successfully.
	// So we suppose that problem is that AWS return response of successful role arn creation before
	// process of creation is really ended, and then creation of datastore model fails.
	for _, sleepSeconds := range retrySecondsList {
		err = nil

		_, err = conn.CreateDatastore(params)
		if err == nil {
			break
		}

		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}

	if err != nil {
		return err
	}

	d.SetId(d.Get("name").(string))

	return resourceAwsIotAnalyticsDatastoreRead(d, meta)
}

func flattenCustomerManagedS3(customerManagedS3 *iotanalytics.CustomerManagedDatastoreS3Storage) map[string]interface{} {
	if customerManagedS3 == nil {
		return nil
	}

	rawCustomerManagedS3 := make(map[string]interface{})

	rawCustomerManagedS3["bucket"] = aws.StringValue(customerManagedS3.Bucket)
	rawCustomerManagedS3["role_arn"] = aws.StringValue(customerManagedS3.RoleArn)

	if customerManagedS3.KeyPrefix != nil {
		rawCustomerManagedS3["key_prefix"] = aws.StringValue(customerManagedS3.KeyPrefix)
	}

	return rawCustomerManagedS3
}

func flattenServiceManagedS3(serviceManagedS3 *iotanalytics.ServiceManagedDatastoreS3Storage) map[string]interface{} {
	if serviceManagedS3 == nil {
		return nil
	}

	rawServiceManagedS3 := make(map[string]interface{})
	return rawServiceManagedS3
}

func flattenStorage(datastoreStorage *iotanalytics.DatastoreStorage) map[string]interface{} {
	customerManagedS3 := flattenCustomerManagedS3(datastoreStorage.CustomerManagedS3)
	serviceManagedS3 := flattenServiceManagedS3(datastoreStorage.ServiceManagedS3)

	if customerManagedS3 == nil && serviceManagedS3 == nil {
		return nil
	}

	rawStorage := make(map[string]interface{})
	rawStorage["customer_managed_s3"] = wrapMapInList(customerManagedS3)
	rawStorage["service_managed_s3"] = wrapMapInList(serviceManagedS3)
	return rawStorage
}

func flattenRetentionPeriod(retentionPeriod *iotanalytics.RetentionPeriod) map[string]interface{} {
	rawRetentionPeriod := make(map[string]interface{})

	if retentionPeriod.NumberOfDays != nil {
		rawRetentionPeriod["number_of_days"] = aws.Int64Value(retentionPeriod.NumberOfDays)
	}
	if retentionPeriod.Unlimited != nil {
		rawRetentionPeriod["unlimited"] = aws.BoolValue(retentionPeriod.Unlimited)
	}

	return rawRetentionPeriod
}

func wrapMapInList(mapping map[string]interface{}) []interface{} {
	if mapping == nil {
		return make([]interface{}, 0)
	} else {
		return []interface{}{mapping}
	}
}

func resourceAwsIotAnalyticsDatastoreRead(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DescribeDatastoreInput{
		DatastoreName: aws.String(d.Id()),
	}

	log.Printf("[DEBUG] Reading IoTAnalytics Datastore: %s", params)

	out, err := conn.DescribeDatastore(params)

	if err != nil {
		return err
	}

	d.Set("name", out.Datastore.Name)
	storage := flattenStorage(out.Datastore.Storage)
	d.Set("storage", wrapMapInList(storage))
	retentionPeriod := flattenRetentionPeriod(out.Datastore.RetentionPeriod)
	d.Set("retention_period", wrapMapInList(retentionPeriod))

	return nil
}

func resourceAwsIotAnalyticsDatastoreUpdate(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.UpdateDatastoreInput{
		DatastoreName: aws.String(d.Get("name").(string)),
	}

	datastoreStorageSet := d.Get("storage").(*schema.Set).List()
	if len(datastoreStorageSet) >= 1 {
		rawDatastoreStorage := datastoreStorageSet[0].(map[string]interface{})
		params.DatastoreStorage = parseStorage(rawDatastoreStorage)
	}

	retentionPeriodSet := d.Get("retention_period").(*schema.Set).List()
	if len(retentionPeriodSet) >= 1 {
		rawRetentionPeriod := retentionPeriodSet[0].(map[string]interface{})
		params.RetentionPeriod = parseRetentionPeriod(rawRetentionPeriod)
	}

	log.Printf("[DEBUG] Updating IoTAnalytics Datastore: %s", params)

	retrySecondsList := [6]int{1, 2, 5, 8, 10, 0}

	var err error

	// Primitive retry.
	// Full explanation can be found in function `resourceAwsIotAnalyticsDatastoreCreate`.
	// We suppose that such error can appear during update also, if you update
	// role arn.
	for _, sleepSeconds := range retrySecondsList {
		err = nil

		_, err = conn.UpdateDatastore(params)
		if err == nil {
			break
		}

		time.Sleep(time.Duration(sleepSeconds) * time.Second)
	}

	if err != nil {
		return err
	}

	return resourceAwsIotAnalyticsDatastoreRead(d, meta)
}

func resourceAwsIotAnalyticsDatastoreDelete(d *schema.ResourceData, meta interface{}) error {
	conn := meta.(*AWSClient).iotanalyticsconn

	params := &iotanalytics.DeleteDatastoreInput{
		DatastoreName: aws.String(d.Id()),
	}
	log.Printf("[DEBUG] Delete IoTAnalytics Datastore: %s", params)
	_, err := conn.DeleteDatastore(params)

	return err
}
