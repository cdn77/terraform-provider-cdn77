package util

import (
	"context"
	"net/http"

	"github.com/cdn77/cdn77-client-go/v2"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
)

type Reader interface {
	RemoveMissingResource() Reader
	Read(
		ctx context.Context,
		client cdn77.ClientWithResponsesInterface,
		reqStateProvider StateProvider,
		respState *tfsdk.State,
		diags *diag.Diagnostics,
	)
	Fill(ctx context.Context, client cdn77.ClientWithResponsesInterface, model *any) diag.Diagnostics
}

type StateProvider interface {
	Get(ctx context.Context, target any) diag.Diagnostics
}

type GenericReader[M any, R Response, Rok any] interface {
	ErrMessage() string
	Fetch(ctx context.Context, client cdn77.ClientWithResponsesInterface, model M) (R, *Rok, error)
	Process(ctx context.Context, model M, detail *Rok, diags *diag.Diagnostics) M
}

type UniversalReader[M any, R Response, Rok any] struct {
	reader                GenericReader[M, R, Rok]
	removeMissingResource bool
}

func NewUniversalReader[M any, R Response, Rok any](reader GenericReader[M, R, Rok]) *UniversalReader[M, R, Rok] {
	return &UniversalReader[M, R, Rok]{reader: reader}
}

func (r UniversalReader[M, R, Rok]) RemoveMissingResource() Reader {
	return UniversalReader[M, R, Rok]{reader: r.reader, removeMissingResource: true}
}

func (r UniversalReader[M, R, Rok]) Read(
	ctx context.Context,
	client cdn77.ClientWithResponsesInterface,
	reqStateProvider StateProvider,
	respState *tfsdk.State,
	diags *diag.Diagnostics,
) {
	var model M
	if diags.Append(reqStateProvider.Get(ctx, &model)...); diags.HasError() {
		return
	}

	response, responseOk, err := r.reader.Fetch(ctx, client, model)
	if err != nil {
		diags.AddError(r.reader.ErrMessage(), err.Error())

		return
	}

	if r.removeMissingResource && response.StatusCode() == http.StatusNotFound {
		respState.RemoveResource(ctx)

		return
	}

	ProcessResponse(diags, response, r.reader.ErrMessage(), responseOk, func(detail *Rok) {
		if model = r.reader.Process(ctx, model, detail, diags); diags.HasError() {
			return
		}

		diags.Append(respState.Set(ctx, model)...)
	})
}

func (r UniversalReader[M, R, Rok]) Fill(
	ctx context.Context,
	client cdn77.ClientWithResponsesInterface,
	model *any,
) (diags diag.Diagnostics) {
	response, responseOk, err := r.reader.Fetch(ctx, client, (*model).(M))
	if err != nil {
		return diag.Diagnostics{diag.NewErrorDiagnostic(r.reader.ErrMessage(), err.Error())}
	}

	ProcessResponse(&diags, response, r.reader.ErrMessage(), responseOk, func(detail *Rok) {
		*model = r.reader.Process(ctx, (*model).(M), detail, &diags)
	})

	return diags
}
