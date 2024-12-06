//go:generate oapi-codegen --config oapi/config.yaml oapi/api.yaml

package api

import (
	"context"
	"errors"

	oapi "pi-wegrzyn/frontend/api/oapi/generated"
	"pi-wegrzyn/storage"
)

type Repository interface {
	CreateDevice(ctx context.Context, device storage.Device) error
	Device(ctx context.Context, id uint) (storage.Device, error)
	Devices(ctx context.Context) ([]storage.Device, error)
	UpdateDevice(ctx context.Context, device storage.Device) error
	DeleteDevice(ctx context.Context, id uint) error
}

type ServerAPI struct {
	repository Repository
}

func NewServerAPI(repository Repository) *ServerAPI {
	return &ServerAPI{
		repository: repository,
	}
}

func (s *ServerAPI) Get(ctx context.Context, request oapi.GetRequestObject) (oapi.GetResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) PostDelete(ctx context.Context, request oapi.PostDeleteRequestObject) (oapi.PostDeleteResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) GetEdit(ctx context.Context, request oapi.GetEditRequestObject) (oapi.GetEditResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) PostEdit(ctx context.Context, request oapi.PostEditRequestObject) (oapi.PostEditResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) GetFaviconIco(ctx context.Context, request oapi.GetFaviconIcoRequestObject) (oapi.GetFaviconIcoResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) GetLogout(ctx context.Context, request oapi.GetLogoutRequestObject) (oapi.GetLogoutResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) GetNew(ctx context.Context, request oapi.GetNewRequestObject) (oapi.GetNewResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) PostNew(ctx context.Context, request oapi.PostNewRequestObject) (oapi.PostNewResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) GetSignin(ctx context.Context, request oapi.GetSigninRequestObject) (oapi.GetSigninResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) PostSignin(ctx context.Context, request oapi.PostSigninRequestObject) (oapi.PostSigninResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *ServerAPI) GetStyleCss(ctx context.Context, request oapi.GetStyleCssRequestObject) (oapi.GetStyleCssResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}
