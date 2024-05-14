package util

import (
	"fmt"
	"strings"

	ds_schema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rsc_schema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type ResourceDataSourceSchemaConverter struct {
	requiredAttrs       map[string]struct{}
	requiredNestedAttrs map[string][]string
}

func NewResourceDataSourceSchemaConverter(requiredAttrs ...string) *ResourceDataSourceSchemaConverter {
	requiredAttrsMap := make(map[string]struct{})
	requiredNestedAttrsMap := make(map[string][]string)

	for _, requiredAttr := range requiredAttrs {
		if i := strings.Index(requiredAttr, "."); i != -1 {
			attr := requiredAttr[:i]
			requiredNestedAttrsMap[attr] = append(requiredNestedAttrsMap[attr], requiredAttr[i+1:])

			continue
		}

		requiredAttrsMap[requiredAttr] = struct{}{}
	}

	return &ResourceDataSourceSchemaConverter{
		requiredAttrs:       requiredAttrsMap,
		requiredNestedAttrs: requiredNestedAttrsMap,
	}
}

func (c *ResourceDataSourceSchemaConverter) Convert(rsc rsc_schema.Schema) ds_schema.Schema {
	if rsc.Blocks != nil {
		panic("Resource to DataSource schema converter doesnt' support blocks")
	}

	return ds_schema.Schema{
		Attributes:          c.convertAttributes(rsc.Attributes),
		Description:         rsc.Description,
		MarkdownDescription: rsc.MarkdownDescription,
		DeprecationMessage:  rsc.DeprecationMessage,
	}
}

func (c *ResourceDataSourceSchemaConverter) convertAttributes(
	rscAttrs map[string]rsc_schema.Attribute,
) map[string]ds_schema.Attribute {
	dsAttrs := make(map[string]ds_schema.Attribute, len(rscAttrs))

	for name, rscAttr := range rscAttrs {
		_, isRequired := c.requiredAttrs[name]
		var dsAttr ds_schema.Attribute

		switch rscAttr := rscAttr.(type) {
		case rsc_schema.BoolAttribute:
			dsAttr = ds_schema.BoolAttribute{
				CustomType:          rscAttr.CustomType,
				Required:            isRequired,
				Computed:            !isRequired,
				Sensitive:           rscAttr.Sensitive,
				Description:         rscAttr.Description,
				MarkdownDescription: rscAttr.MarkdownDescription,
				DeprecationMessage:  rscAttr.DeprecationMessage,
				Validators:          If(isRequired, rscAttr.Validators, nil),
			}
		case rsc_schema.StringAttribute:
			dsAttr = ds_schema.StringAttribute{
				CustomType:          rscAttr.CustomType,
				Required:            isRequired,
				Computed:            !isRequired,
				Sensitive:           rscAttr.Sensitive,
				Description:         rscAttr.Description,
				MarkdownDescription: rscAttr.MarkdownDescription,
				DeprecationMessage:  rscAttr.DeprecationMessage,
				Validators:          If(isRequired, rscAttr.Validators, nil),
			}
		case rsc_schema.Int64Attribute:
			dsAttr = ds_schema.Int64Attribute{
				CustomType:          rscAttr.CustomType,
				Required:            isRequired,
				Computed:            !isRequired,
				Sensitive:           rscAttr.Sensitive,
				Description:         rscAttr.Description,
				MarkdownDescription: rscAttr.MarkdownDescription,
				DeprecationMessage:  rscAttr.DeprecationMessage,
				Validators:          If(isRequired, rscAttr.Validators, nil),
			}
		case rsc_schema.SetAttribute:
			dsAttr = ds_schema.SetAttribute{
				ElementType:         rscAttr.ElementType,
				CustomType:          rscAttr.CustomType,
				Required:            isRequired,
				Computed:            !isRequired,
				Sensitive:           rscAttr.Sensitive,
				Description:         rscAttr.Description,
				MarkdownDescription: rscAttr.MarkdownDescription,
				DeprecationMessage:  rscAttr.DeprecationMessage,
				Validators:          If(isRequired, rscAttr.Validators, nil),
			}
		case rsc_schema.MapAttribute:
			dsAttr = ds_schema.MapAttribute{
				ElementType:         rscAttr.ElementType,
				CustomType:          rscAttr.CustomType,
				Required:            isRequired,
				Computed:            !isRequired,
				Sensitive:           rscAttr.Sensitive,
				Description:         rscAttr.Description,
				MarkdownDescription: rscAttr.MarkdownDescription,
				DeprecationMessage:  rscAttr.DeprecationMessage,
				Validators:          If(isRequired, rscAttr.Validators, nil),
			}
		case rsc_schema.SingleNestedAttribute:
			childConverter := NewResourceDataSourceSchemaConverter(c.requiredNestedAttrs[name]...)
			dsAttr = ds_schema.SingleNestedAttribute{
				Attributes:          childConverter.convertAttributes(rscAttr.Attributes),
				CustomType:          rscAttr.CustomType,
				Required:            isRequired,
				Computed:            !isRequired,
				Sensitive:           rscAttr.Sensitive,
				Description:         rscAttr.Description,
				MarkdownDescription: rscAttr.MarkdownDescription,
				DeprecationMessage:  rscAttr.DeprecationMessage,
				Validators:          If(isRequired, rscAttr.Validators, nil),
			}
		default:
			const message = `Resource to DataSource schema converter encountered unsupported Attribute type "%T"`

			panic(fmt.Sprintf(message, rscAttr))
		}

		dsAttrs[name] = dsAttr
	}

	return dsAttrs
}
