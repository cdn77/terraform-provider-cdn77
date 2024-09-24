package origin

import (
	"github.com/cdn77/terraform-provider-cdn77/internal/provider/shared"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
)

type UrlModel struct {
	SharedModel
	shared.UrlModel
}

func CreateUrlResourceSchema() schema.Schema {
	return WithSharedSchemaAttrs(shared.WithUrlSchemaAttrs(schema.Schema{
		MarkdownDescription: "URL Origin resource allows you to manage your custom URL Origins",
		Attributes:          map[string]schema.Attribute{},
	}))
}
