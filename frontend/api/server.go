//go:generate oapi-codegen --config oapi/config.yaml oapi/api.yaml

package api

import (
	"context"
	"errors"
	"net/http"
	"text/template"

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

type StrictServerImpl struct {
	config     Config
	repository Repository
	cookies    *map[string]Cookie
	templates  map[string]*template.Template
}

func NewServerAPI(config Config, repository Repository, cookies *map[string]Cookie) (http.Handler, error) {
	templates, err := initTemplates(config.TemplatesDir)
	if err != nil {
		return nil, err
	}

	// TODO: read static files

	return oapi.Handler(oapi.NewStrictHandler(&StrictServerImpl{
		config:     config,
		repository: repository,
		cookies:    cookies,
		templates:  templates,
	}, []oapi.StrictMiddlewareFunc{NewAuthMiddleware(config, cookies)})), nil
}

func (s *StrictServerImpl) Get(ctx context.Context, request oapi.GetRequestObject) (oapi.GetResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) PostDelete(ctx context.Context, request oapi.PostDeleteRequestObject) (oapi.PostDeleteResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) GetEdit(ctx context.Context, request oapi.GetEditRequestObject) (oapi.GetEditResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) PostEdit(ctx context.Context, request oapi.PostEditRequestObject) (oapi.PostEditResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) GetFaviconIco(ctx context.Context, request oapi.GetFaviconIcoRequestObject) (oapi.GetFaviconIcoResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) GetLogout(ctx context.Context, request oapi.GetLogoutRequestObject) (oapi.GetLogoutResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) GetNew(ctx context.Context, request oapi.GetNewRequestObject) (oapi.GetNewResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) PostNew(ctx context.Context, request oapi.PostNewRequestObject) (oapi.PostNewResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) GetSignin(ctx context.Context, request oapi.GetSigninRequestObject) (oapi.GetSigninResponseObject, error) {
	return newTemplateWriter(s.templates[TemplateSignIn], nil), nil
}

func (s *StrictServerImpl) PostSignin(ctx context.Context, request oapi.PostSigninRequestObject) (oapi.PostSigninResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

func (s *StrictServerImpl) GetStyleCss(ctx context.Context, request oapi.GetStyleCssRequestObject) (oapi.GetStyleCssResponseObject, error) {
	// TODO
	return nil, errors.New("not implemented")
}

type templateWriter struct {
	*template.Template
	content any
}

func newTemplateWriter(templ *template.Template, content any) templateWriter {
	return templateWriter{templ, content}
}

func (t templateWriter) VisitGetSigninResponse(w http.ResponseWriter) error {
	return t.Execute(w, t.content)
}
