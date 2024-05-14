package util

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/resource/schema/defaults"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func Int64NullDefault() defaults.Int64 {
	return nullInt64Default{}
}

type nullInt64Default struct{}

func (d nullInt64Default) Description(ctx context.Context) string {
	return d.MarkdownDescription(ctx)
}

func (nullInt64Default) MarkdownDescription(_ context.Context) string {
	return "value defaults to null"
}

func (nullInt64Default) DefaultInt64(_ context.Context, _ defaults.Int64Request, resp *defaults.Int64Response) {
	resp.PlanValue = types.Int64Null()
}
