package origin

import (
	"github.com/cdn77/terraform-provider-cdn77/internal/util"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/oapi-codegen/nullable"
)

const (
	TypeAws           = "aws"
	TypeObjectStorage = "object-storage"
	TypeUrl           = "url"
)

type SharedModel struct {
	Id    types.String `tfsdk:"id"`
	Label types.String `tfsdk:"label"`
	Note  types.String `tfsdk:"note"`
}

func NewSharedModel(id types.String, label string, note nullable.Nullable[string]) SharedModel {
	return SharedModel{
		Id:    id,
		Label: types.StringValue(label),
		Note:  util.NullableToStringValue(note),
	}
}

func WithSharedSchemaAttrs(s schema.Schema) schema.Schema {
	s.Attributes["id"] = schema.StringAttribute{
		Description:   "Origin ID (UUID)",
		Computed:      true,
		PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
	}
	s.Attributes["label"] = schema.StringAttribute{
		Description: "The label helps you to identify your Origin",
		Required:    true,
	}
	s.Attributes["note"] = schema.StringAttribute{
		Description: "Optional note for the Origin",
		Optional:    true,
	}

	return s
}
