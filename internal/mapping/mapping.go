package mapping

import (
	"fmt"

	"github.com/cdn77/terraform-provider-cdn77/internal/provider/cdn"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/object_storages"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/origin"
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/ssl"
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	ds_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	rsc_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type Resource string

const (
	Cdn                 = Resource("cdn")
	Cdns                = Resource("cdns")
	ObjectStorages      = Resource("object_storages")
	OriginAws           = Resource("origin_aws")
	OriginObjectStorage = Resource("origin_object_storage")
	OriginUrl           = Resource("origin_url")
	Ssl                 = Resource("ssl")
	Ssls                = Resource("ssls")
)

func ResourceFactory(rsc Resource) func() resource.Resource {
	return func() resource.Resource {
		schemaProvider, reader := resourceSchemaProviderAndReader(rsc)
		baseResource := util.NewBaseResource(string(rsc), schemaProvider, reader)

		switch rsc {
		case Cdn:
			return &cdn.Resource{BaseResource: baseResource}
		case OriginAws:
			return &origin.AwsResource{BaseResource: baseResource}
		case OriginObjectStorage:
			return &origin.ObjectStorageResource{BaseResource: baseResource}
		case OriginUrl:
			return &origin.UrlResource{BaseResource: baseResource}
		case Ssl:
			return &ssl.Resource{BaseResource: baseResource}
		default:
			panic(fmt.Sprintf("unexpected resource type %q", rsc))
		}
	}
}

func DataSourceFactory(rsc Resource) func() datasource.DataSource {
	return func() datasource.DataSource {
		schemaProvider, reader := dataSourceSchemaProviderAndReader(rsc)
		baseDataSource := util.NewBaseDataSource(string(rsc), schemaProvider, reader)

		switch rsc {
		case Cdn:
			return &cdn.DataSource{BaseDataSource: baseDataSource}
		case Cdns:
			return &cdn.AllDataSource{BaseDataSource: baseDataSource}
		case ObjectStorages:
			return &object_storages.DataSource{BaseDataSource: baseDataSource}
		case OriginAws:
			return &origin.AwsDataSource{BaseDataSource: baseDataSource}
		case OriginObjectStorage:
			return &origin.ObjectStorageDataSource{BaseDataSource: baseDataSource}
		case OriginUrl:
			return &origin.UrlDataSource{BaseDataSource: baseDataSource}
		case Ssl:
			return &ssl.DataSource{BaseDataSource: baseDataSource}
		case Ssls:
			return &ssl.AllDataSource{BaseDataSource: baseDataSource}
		default:
			panic(fmt.Sprintf("unexpected resource type %q", rsc))
		}
	}
}

func resourceSchemaProviderAndReader(rsc Resource) (func() rsc_schema.Schema, util.Reader) {
	switch rsc {
	case Cdn, Cdns:
		return cdn.CreateResourceSchema, util.NewUniversalReader(&cdn.Reader{})
	case OriginAws:
		return origin.CreateAwsResourceSchema, util.NewUniversalReader(&origin.AwsReader{})
	case OriginObjectStorage:
		return origin.CreateObjectStorageResourceSchema, util.NewUniversalReader(&origin.ObjectStorageReader{})
	case OriginUrl:
		return origin.CreateUrlResourceSchema, util.NewUniversalReader(&origin.UrlReader{})
	case Ssl:
		return ssl.CreateResourceSchema, util.NewUniversalReader(&ssl.Reader{})
	case Ssls:
		return ssl.CreateBaseResourceSchema, nil
	default:
		panic(fmt.Sprintf("unexpected resource type %q", rsc))
	}
}

func dataSourceSchemaProviderAndReader(rsc Resource) (func() ds_schema.Schema, util.Reader) {
	if rsc == ObjectStorages {
		return object_storages.CreateSchema, nil
	}

	schemaProvider, reader := resourceSchemaProviderAndReader(rsc)

	switch rsc {
	case Cdn, OriginAws, OriginObjectStorage, OriginUrl, Ssl:
		return toDataSourceSchema(schemaProvider, "id"), reader
	case Cdns, Ssls:
		return toDataSourceSchema(schemaProvider), reader
	default:
		panic(fmt.Sprintf("unexpected resource type %q", rsc))
	}
}

func toDataSourceSchema(schemaProvider func() rsc_schema.Schema, requiredAttrs ...string) func() ds_schema.Schema {
	return func() ds_schema.Schema {
		return util.NewResourceDataSourceSchemaConverter(requiredAttrs...).Convert(schemaProvider())
	}
}
