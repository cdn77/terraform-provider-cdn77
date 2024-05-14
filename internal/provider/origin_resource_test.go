package provider_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/cdn77/cdn77-client-go"
	"github.com/cdn77/terraform-provider-cdn77/internal/acctest"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider"
	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccOriginResource_Aws(t *testing.T) {
	client := acctest.GetClient(t)
	var originId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		CheckDestroy:             checkOriginsDestroyed(client),
		Steps: []resource.TestStep{
			{
				Config: `resource "cdn77_origin" "aws" {
					type = "aws"
					label = "some label"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.aws", plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.aws", "id", func(value string) error {
						originId = value

						return acctest.NotEqual(value, "")
					}),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "label", "some label"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "base_dir"),
					checkAwsOrigin(client, &originId, func(o *cdn77.S3OriginDetail) error {
						return errors.Join(
							acctest.NullField("aws_access_key_id", o.AwsAccessKeyId),
							acctest.NullField("aws_region", o.AwsRegion),
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "some label"),
							acctest.NullField("note", o.Note),
							acctest.NullField("port", o.Port),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeAws),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "aws" {
					type = "aws"
					label = "another label"
					note = "some note"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
					port = 12345
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.aws", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.aws", "id", func(value string) error {
						return acctest.EqualField("id", value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "note", "some note"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "port", "12345"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "base_dir"),
					checkAwsOrigin(client, &originId, func(o *cdn77.S3OriginDetail) error {
						return errors.Join(
							acctest.NullField("aws_access_key_id", o.AwsAccessKeyId),
							acctest.NullField("aws_region", o.AwsRegion),
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "some note"),
							acctest.NullFieldEqual("port", o.Port, 12345),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeAws),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "aws" {
					type = "aws"
					label = "another label"
					note = "some note"
					aws_access_key_id = "keyid"
					aws_access_key_secret = "keysecret"
					aws_region = "eu"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
					port = 12345
					base_dir = "some-dir"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.aws", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.aws", "id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "note", "some note"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "aws_access_key_id", "keyid"),
					resource.TestCheckResourceAttrSet("cdn77_origin.aws", "aws_access_key_secret"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "aws_region", "eu"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "port", "12345"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "base_dir", "some-dir"),
					checkAwsOrigin(client, &originId, func(o *cdn77.S3OriginDetail) error {
						return errors.Join(
							acctest.NullFieldEqual("aws_access_key_id", o.AwsAccessKeyId, "keyid"),
							acctest.NullFieldEqual("aws_region", o.AwsRegion, "eu"),
							acctest.NullFieldEqual("base_dir", o.BaseDir, "some-dir"),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "some note"),
							acctest.NullFieldEqual("port", o.Port, 12345),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeAws),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "aws" {
					type = "aws"
					label = "another label"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.aws", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.aws", "id", func(value string) error {
						return acctest.EqualField("id", value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "base_dir"),
					checkAwsOrigin(client, &originId, func(o *cdn77.S3OriginDetail) error {
						return errors.Join(
							acctest.NullField("aws_access_key_id", o.AwsAccessKeyId),
							acctest.NullField("aws_region", o.AwsRegion),
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullField("note", o.Note),
							acctest.NullField("port", o.Port),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeAws),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "aws" {
					type = "url"
					label = "another label"
					note = "another note"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.aws", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.aws", "id", func(value string) error {
						err := acctest.NotEqual(value, originId)
						originId = value

						return err
					}),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "note", "another note"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.aws", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.aws", "base_dir"),
					checkUrlOrigin(client, &originId, func(o *cdn77.UrlOriginDetail) error {
						return errors.Join(
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "another note"),
							acctest.NullField("port", o.Port),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeUrl),
						)
					}),
				),
			},
		},
	})
}

func TestAccOriginResource_ObjectStorage(t *testing.T) {
	client := acctest.GetClient(t)
	bucketName := "my-bucket-" + uuid.New().String()
	anotherBucketName := "my-bucket-" + uuid.New().String()
	var originId string
	var clusterId string

	const objectStoragesDataSourceConfig = `
		data "cdn77_object_storages" "all" {
		}

		locals {
			eu_cluster_id = one([for os in data.cdn77_object_storages.all.clusters : os.id if os.label == "EU"])
			us_cluster_id = one([for os in data.cdn77_object_storages.all.clusters : os.id if os.label == "US"])
		}
	`

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		CheckDestroy:             checkOriginsDestroyed(client),
		Steps: []resource.TestStep{
			{
				Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin" "os" {
					type = "object-storage"
					label = "some label"
					acl = "private"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.os", plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "id", func(value string) error {
						originId = value

						return acctest.NotEqual(value, "")
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr("cdn77_origin.os", "label", "some label"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "acl", "private"),
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "cluster_id", func(value string) error {
						clusterId = value

						return acctest.NotEqual(value, "")
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_id"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_secret"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "scheme"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "host"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "base_dir"),
					checkObjectStorageOrigin(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
						return errors.Join(
							acctest.EqualField("bucket_name", o.BucketName, bucketName),
							acctest.EqualField("label", o.Label, "some label"),
							acctest.NullField("note", o.Note),
							acctest.EqualField("type", o.Type, provider.OriginTypeObjectStorage),
						)
					}),
				),
			},
			{
				Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin" "os" {
					type = "object-storage"
					label = "another label"
					note = "some note"
					acl = "private"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.os", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr("cdn77_origin.os", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "note", "some note"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "acl", "private"),
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "cluster_id", func(value string) error {
						return acctest.Equal(value, clusterId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_id"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_secret"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "scheme"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "host"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "base_dir"),
					checkObjectStorageOrigin(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
						return errors.Join(
							acctest.EqualField("bucket_name", o.BucketName, bucketName),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "some note"),
							acctest.EqualField("type", o.Type, provider.OriginTypeObjectStorage),
						)
					}),
				),
			},
			{
				Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin" "os" {
					type = "object-storage"
					label = "another label"
					acl = "private"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.os", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr("cdn77_origin.os", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "acl", "private"),
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "cluster_id", func(value string) error {
						return acctest.Equal(value, clusterId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_id"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_secret"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "scheme"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "host"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "base_dir"),
					checkObjectStorageOrigin(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
						return errors.Join(
							acctest.EqualField("bucket_name", o.BucketName, bucketName),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullField("note", o.Note),
							acctest.EqualField("type", o.Type, provider.OriginTypeObjectStorage),
						)
					}),
				),
			},
			{
				Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin" "os" {
					type = "object-storage"
					label = "another label"
					acl = "authenticated-read"
					cluster_id = local.eu_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.os", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "id", func(value string) error {
						err := acctest.NotEqual(value, originId)
						originId = value

						return err
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr("cdn77_origin.os", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "acl", "authenticated-read"),
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "cluster_id", func(value string) error {
						return acctest.Equal(value, clusterId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_id"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_secret"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "scheme"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "host"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "base_dir"),
					checkObjectStorageOrigin(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
						return errors.Join(
							acctest.EqualField("bucket_name", o.BucketName, bucketName),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullField("note", o.Note),
							acctest.EqualField("type", o.Type, provider.OriginTypeObjectStorage),
						)
					}),
				),
			},
			{
				Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin" "os" {
					type = "object-storage"
					label = "another label"
					acl = "authenticated-read"
					cluster_id = local.us_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", bucketName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.os", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "id", func(value string) error {
						err := acctest.NotEqual(value, originId)
						originId = value

						return err
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr("cdn77_origin.os", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "acl", "authenticated-read"),
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "cluster_id", func(value string) error {
						err := acctest.NotEqual(value, clusterId)
						clusterId = value

						return err
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "bucket_name", bucketName),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_id"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_secret"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "scheme"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "host"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "base_dir"),
					checkObjectStorageOrigin(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
						return errors.Join(
							acctest.EqualField("bucket_name", o.BucketName, bucketName),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullField("note", o.Note),
							acctest.EqualField("type", o.Type, provider.OriginTypeObjectStorage),
						)
					}),
				),
			},
			{
				Config: acctest.Config(objectStoragesDataSourceConfig+`resource "cdn77_origin" "os" {
					type = "object-storage"
					label = "another label"
					acl = "authenticated-read"
					cluster_id = local.us_cluster_id
					bucket_name = "{bucketName}"
				}`, "bucketName", anotherBucketName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.os", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "id", func(value string) error {
						err := acctest.NotEqual(value, originId)
						originId = value

						return err
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "type", provider.OriginTypeObjectStorage),
					resource.TestCheckResourceAttr("cdn77_origin.os", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "acl", "authenticated-read"),
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "cluster_id", func(value string) error {
						return acctest.Equal(value, clusterId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "bucket_name", anotherBucketName),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_id"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "access_key_secret"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "scheme"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "host"),
					resource.TestCheckResourceAttrSet("cdn77_origin.os", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "base_dir"),
					checkObjectStorageOrigin(client, &originId, func(o *cdn77.ObjectStorageOriginDetail) error {
						return errors.Join(
							acctest.EqualField("bucket_name", o.BucketName, anotherBucketName),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullField("note", o.Note),
							acctest.EqualField("type", o.Type, provider.OriginTypeObjectStorage),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "os" {
					type = "url"
					label = "another label"
					note = "another note"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.os", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.os", "id", func(value string) error {
						err := acctest.NotEqual(value, originId)
						originId = value

						return err
					}),
					resource.TestCheckResourceAttr("cdn77_origin.os", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("cdn77_origin.os", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "note", "another note"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.os", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "acl"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "cluster_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "bucket_name"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.os", "base_dir"),
					checkUrlOrigin(client, &originId, func(o *cdn77.UrlOriginDetail) error {
						return errors.Join(
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "another note"),
							acctest.NullField("port", o.Port),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeUrl),
						)
					}),
				),
			},
		},
	})
}

func TestAccOriginResource_Url(t *testing.T) {
	client := acctest.GetClient(t)
	var originId string

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: acctest.GetProviderFactories(),
		CheckDestroy:             checkOriginsDestroyed(client),
		Steps: []resource.TestStep{
			{
				Config: `resource "cdn77_origin" "url" {
					type = "url"
					label = "some label"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.url", plancheck.ResourceActionCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.url", "id", func(value string) error {
						originId = value

						return acctest.NotEqual(value, "")
					}),
					resource.TestCheckResourceAttr("cdn77_origin.url", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("cdn77_origin.url", "label", "some label"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "base_dir"),
					checkUrlOrigin(client, &originId, func(o *cdn77.UrlOriginDetail) error {
						return errors.Join(
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "some label"),
							acctest.NullField("note", o.Note),
							acctest.NullField("port", o.Port),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeUrl),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "url" {
					type = "url"
					label = "another label"
					note = "some note"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
					port = 12345
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.url", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.url", "id", func(value string) error {
						return acctest.EqualField("id", value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.url", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("cdn77_origin.url", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "note", "some note"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "port", "12345"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "base_dir"),
					checkUrlOrigin(client, &originId, func(o *cdn77.UrlOriginDetail) error {
						return errors.Join(
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "some note"),
							acctest.NullFieldEqual("port", o.Port, 12345),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeUrl),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "url" {
					type = "url"
					label = "another label"
					note = "some note"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
					port = 12345
					base_dir = "some-dir"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.url", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.url", "id", func(value string) error {
						return acctest.Equal(value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.url", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("cdn77_origin.url", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "note", "some note"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "port", "12345"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "base_dir", "some-dir"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_region"),
					checkUrlOrigin(client, &originId, func(o *cdn77.UrlOriginDetail) error {
						return errors.Join(
							acctest.NullFieldEqual("base_dir", o.BaseDir, "some-dir"),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "some note"),
							acctest.NullFieldEqual("port", o.Port, 12345),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeUrl),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "url" {
					type = "url"
					label = "another label"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.url", plancheck.ResourceActionUpdate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.url", "id", func(value string) error {
						return acctest.EqualField("id", value, originId)
					}),
					resource.TestCheckResourceAttr("cdn77_origin.url", "type", provider.OriginTypeUrl),
					resource.TestCheckResourceAttr("cdn77_origin.url", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "note"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_id"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_access_key_secret"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "aws_region"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "base_dir"),
					checkUrlOrigin(client, &originId, func(o *cdn77.UrlOriginDetail) error {
						return errors.Join(
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullField("note", o.Note),
							acctest.NullField("port", o.Port),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeUrl),
						)
					}),
				),
			},
			{
				Config: `resource "cdn77_origin" "url" {
					type = "aws"
					label = "another label"
					note = "another note"
					aws_access_key_id = "keyid"
					aws_access_key_secret = "keysecret"
					aws_region = "eu"
					scheme = "http"
					host = "my-totally-random-custom-host.com"
				}`,
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectResourceAction("cdn77_origin.url", plancheck.ResourceActionDestroyBeforeCreate),
					},
				},
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrWith("cdn77_origin.url", "id", func(value string) error {
						err := acctest.NotEqual(value, originId)
						originId = value

						return err
					}),
					resource.TestCheckResourceAttr("cdn77_origin.url", "type", provider.OriginTypeAws),
					resource.TestCheckResourceAttr("cdn77_origin.url", "label", "another label"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "note", "another note"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "aws_access_key_id", "keyid"),
					resource.TestCheckResourceAttrSet("cdn77_origin.url", "aws_access_key_secret"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "aws_region", "eu"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "scheme", "http"),
					resource.TestCheckResourceAttr("cdn77_origin.url", "host", "my-totally-random-custom-host.com"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "port"),
					resource.TestCheckNoResourceAttr("cdn77_origin.url", "base_dir"),
					checkAwsOrigin(client, &originId, func(o *cdn77.S3OriginDetail) error {
						return errors.Join(
							acctest.NullFieldEqual("aws_access_key_id", o.AwsAccessKeyId, "keyid"),
							acctest.NullFieldEqual("aws_region", o.AwsRegion, "eu"),
							acctest.NullField("base_dir", o.BaseDir),
							acctest.EqualField("host", o.Host, "my-totally-random-custom-host.com"),
							acctest.EqualField("label", o.Label, "another label"),
							acctest.NullFieldEqual("note", o.Note, "another note"),
							acctest.NullField("port", o.Port),
							acctest.EqualField("scheme", o.Scheme, "http"),
							acctest.EqualField("type", o.Type, provider.OriginTypeAws),
						)
					}),
				),
			},
		},
	})
}

func checkAwsOrigin(
	client cdn77.ClientWithResponsesInterface,
	originId *string,
	fn func(o *cdn77.S3OriginDetail) error,
) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		response, err := client.OriginDetailAwsWithResponse(context.Background(), *originId)
		message := fmt.Sprintf("failed to get Origin[id=%s]: %%s", *originId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

func checkObjectStorageOrigin(
	client cdn77.ClientWithResponsesInterface,
	originId *string,
	fn func(o *cdn77.ObjectStorageOriginDetail) error,
) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		response, err := client.OriginDetailObjectStorageWithResponse(context.Background(), *originId)
		message := fmt.Sprintf("failed to get Origin[id=%s]: %%s", *originId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

func checkUrlOrigin(
	client cdn77.ClientWithResponsesInterface,
	originId *string,
	fn func(o *cdn77.UrlOriginDetail) error,
) func(*terraform.State) error {
	return func(_ *terraform.State) error {
		response, err := client.OriginDetailUrlWithResponse(context.Background(), *originId)
		message := fmt.Sprintf("failed to get Origin[id=%s]: %%s", *originId)

		if err = acctest.CheckResponse(message, response, err); err != nil {
			return err
		}

		return fn(response.JSON200)
	}
}

func checkOriginsDestroyed(client cdn77.ClientWithResponsesInterface) func(*terraform.State) error {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "cdn77_origin" {
				continue
			}

			originId := rs.Primary.Attributes["id"]

			switch rs.Primary.Attributes["type"] {
			case provider.OriginTypeAws:
				response, err := client.OriginDetailAwsWithResponse(context.Background(), originId)
				if err != nil {
					return fmt.Errorf("failed to fetch Origin: %w", err)
				}

				if response.JSON404 == nil {
					return errors.New("expected origin to be deleted")
				}
			case provider.OriginTypeObjectStorage:
				response, err := client.OriginDetailObjectStorageWithResponse(context.Background(), originId)
				if err != nil {
					return fmt.Errorf("failed to fetch Origin: %w", err)
				}

				if response.JSON404 == nil {
					return errors.New("expected origin to be deleted")
				}
			case provider.OriginTypeUrl:
				response, err := client.OriginDetailUrlWithResponse(context.Background(), originId)
				if err != nil {
					return fmt.Errorf("failed to fetch Origin: %w", err)
				}

				if response.JSON404 == nil {
					return errors.New("expected origin to be deleted")
				}
			default:
				return fmt.Errorf("unexpected Origin type: %s", rs.Primary.Attributes["type"])
			}
		}

		return nil
	}
}
