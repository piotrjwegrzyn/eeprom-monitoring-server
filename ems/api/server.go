//go:generate go tool oapi-codegen --config oapi/config.yaml oapi/api.yaml

package api

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"net/http"

	oapi "pi-wegrzyn/ems/api/oapi/generated"
	"pi-wegrzyn/ems/cookies"
	"pi-wegrzyn/ems/storage"
	"pi-wegrzyn/ems/templates"
)

type Repository interface {
	CreateDevice(ctx context.Context, device storage.Device) error
	Device(ctx context.Context, id uint) (storage.Device, error)
	Devices(ctx context.Context) ([]storage.Device, error)
	UpdateDevice(ctx context.Context, device storage.Device) error
	DeleteDevice(ctx context.Context, id uint) error
}

type Cookies interface {
	Create(ctx context.Context, w http.ResponseWriter, username string)
	Delete(ctx context.Context, w http.ResponseWriter, token *string)
}

type Config struct {
	User     string `envconfig:"ADMIN_USER"`
	Password string `envconfig:"ADMIN_PASSWORD"`
}

type StaticFiles struct {
	CSS     []byte
	Favicon []byte
}

type APIServer struct {
	config     Config
	repository Repository
	cookies    Cookies

	templateEx  *templates.Executor
	staticFiles *StaticFiles
}

func NewServerAPI(
	config Config,
	repository Repository,
	cookieStore *cookies.Store,
	executor *templates.Executor,
	staticFiles *StaticFiles,
) http.Handler {
	return oapi.Handler(oapi.NewStrictHandler(&APIServer{
		config:      config,
		repository:  repository,
		cookies:     cookieStore,
		templateEx:  executor,
		staticFiles: staticFiles,
	}, []oapi.StrictMiddlewareFunc{
		NewLoggerMiddleware(),
		NewAuthMiddleware(config, cookieStore),
	}))
}

func (s *APIServer) Get(ctx context.Context, request oapi.GetRequestObject) (oapi.GetResponseObject, error) {
	devices, err := s.repository.Devices(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "error getting devices", slog.Any("error", err))
		return oapi.Get500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error getting devices",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	page, err := s.templateEx.ExecuteIndex(devices)
	if err != nil {
		slog.ErrorContext(ctx, "error executing template", slog.Any("error", err))
		return oapi.Get500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error executing template",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	return oapi.Get200TexthtmlResponse{
		PageTexthtmlResponse: oapi.PageTexthtmlResponse{
			Body:          page,
			ContentLength: int64(page.Len()),
		},
	}, nil
}

func (s *APIServer) GetEdit(ctx context.Context, request oapi.GetEditRequestObject) (oapi.GetEditResponseObject, error) {
	device, err := s.repository.Device(ctx, request.Params.EditId)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.ErrorContext(ctx, "device not found", slog.Any("error", err))
			return oapi.PageRedirectResponse{
				Headers: oapi.PageRedirectResponseHeaders{
					Location: "/",
				},
			}, nil
		}
		slog.ErrorContext(ctx, "database error", slog.Any("error", err))

		return oapi.GetEdit500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "database error",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	page, err := s.templateEx.ExecuteNewEdit(templates.EditPageContent(device, ""))
	if err != nil {
		slog.ErrorContext(ctx, "error executing template", slog.Any("error", err))
		return oapi.GetEdit500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error executing template",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	return oapi.GetEdit200TexthtmlResponse{
		PageTexthtmlResponse: oapi.PageTexthtmlResponse{
			Body:          page,
			ContentLength: int64(page.Len()),
		},
	}, nil
}

func (s *APIServer) PostEdit(ctx context.Context, request oapi.PostEditRequestObject) (oapi.PostEditResponseObject, error) {
	form, err := templates.ParseForm(request.Body)
	if err != nil {
		slog.ErrorContext(ctx, "error parsing form", slog.Any("error", err))
		return oapi.PostEdit500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error parsing form",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	if err := form.Validate(); err != nil {
		slog.ErrorContext(ctx, "validation error", slog.Any("error", err))

		return s.postEditError(
			ctx,
			storage.Device{
				ID:        form.EditId,
				Hostname:  form.Hostname,
				IPAddress: form.Ip,
				Login:     form.Login,
			},
			err,
		), nil
	}

	device, err := s.repository.Device(ctx, form.EditId)
	if err != nil {
		if err == sql.ErrNoRows {
			slog.ErrorContext(ctx, "device not found", slog.Any("error", err))
			return oapi.PageRedirectResponse{
				Headers: oapi.PageRedirectResponseHeaders{
					Location: "/",
				},
			}, nil
		}
		slog.ErrorContext(ctx, "database error", slog.Any("error", err))

		return oapi.PostEdit500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "database error",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	device.Hostname = form.Hostname
	device.IPAddress = form.Ip
	device.Login = form.Login
	if form.PasswordClear != nil {
		device.Password = *form.Password
	}

	switch {
	case form.KeyClear != nil:
		device.Keyfile = []byte{}
	case len(form.Key) != 0:
		device.Keyfile = form.Key
	default:
	}

	if err = s.repository.UpdateDevice(ctx, device); err != nil {
		slog.ErrorContext(ctx, "database error", slog.Any("error", err))
	}

	return oapi.PostEdit303Response{
		Headers: oapi.PageRedirectResponseHeaders{
			Location: "/",
		},
	}, nil
}

func (s *APIServer) postEditError(ctx context.Context, device storage.Device, err error) oapi.PostEditResponseObject {
	page, err2 := s.templateEx.ExecuteNewEdit(templates.EditPageContent(device, err.Error()))
	if err2 != nil {
		slog.ErrorContext(ctx, "error executing template", slog.Any("error", err2))

		return oapi.PostEdit500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error executing template",
				ErrorDetails: ptr(errors.Join(err, err2).Error()),
			},
		}
	}

	return oapi.PostEdit200TexthtmlResponse{
		PageTexthtmlResponse: oapi.PageTexthtmlResponse{
			Body:          page,
			ContentLength: int64(page.Len()),
		},
	}
}

func (s *APIServer) GetNew(ctx context.Context, request oapi.GetNewRequestObject) (oapi.GetNewResponseObject, error) {
	page, err := s.templateEx.ExecuteNewEdit(templates.NewPageContent())
	if err != nil {
		slog.ErrorContext(ctx, "error executing template", slog.Any("error", err))
		return oapi.GetNew500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error executing template",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	return oapi.GetNew200TexthtmlResponse{
		PageTexthtmlResponse: oapi.PageTexthtmlResponse{
			Body:          page,
			ContentLength: int64(page.Len()),
		},
	}, nil
}

func (s *APIServer) PostNew(ctx context.Context, request oapi.PostNewRequestObject) (oapi.PostNewResponseObject, error) {
	form, err := templates.ParseForm(request.Body)
	if err != nil {
		slog.ErrorContext(ctx, "error parsing form", slog.Any("error", err))
		return oapi.PostNew500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error parsing form",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	if err := form.Validate(); err != nil {
		slog.ErrorContext(ctx, "validation error", slog.Any("error", err))

		return s.postNewError(
			ctx,
			storage.Device{
				Hostname:  form.Hostname,
				IPAddress: form.Ip,
				Login:     form.Login,
			},
			err,
		), nil
	}

	if err = s.repository.CreateDevice(ctx, storage.Device{
		Hostname:  form.Hostname,
		IPAddress: form.Ip,
		Login:     form.Login,
		Password:  *form.Password,
		Keyfile:   form.Key,
	}); err != nil {
		slog.ErrorContext(ctx, "database error", slog.Any("error", err))
	}

	return oapi.PostNew303Response{
		Headers: oapi.PageRedirectResponseHeaders{
			Location: "/",
		},
	}, nil
}

func (s *APIServer) postNewError(ctx context.Context, device storage.Device, err error) oapi.PostNewResponseObject {
	page, err2 := s.templateEx.ExecuteNewEdit(templates.NewPageContentWithError(device, err.Error()))
	if err2 != nil {
		slog.ErrorContext(ctx, "error executing template", slog.Any("error", err2))
		return oapi.PostNew500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error executing template",
				ErrorDetails: ptr(errors.Join(err, err2).Error()),
			},
		}
	}

	return oapi.PostNew200TexthtmlResponse{
		PageTexthtmlResponse: oapi.PageTexthtmlResponse{
			Body:          page,
			ContentLength: int64(page.Len()),
		},
	}
}

func (s *APIServer) PostDelete(ctx context.Context, request oapi.PostDeleteRequestObject) (oapi.PostDeleteResponseObject, error) {
	err := s.repository.DeleteDevice(ctx, request.Body.DeleteId)
	switch err {
	case nil:
	case sql.ErrNoRows:
		slog.ErrorContext(ctx, "device not found", slog.Any("error", err))
	default:
		slog.ErrorContext(ctx, "database error", slog.Any("error", err))
		return oapi.PostDelete500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "database error",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	return oapi.PostDelete303Response{
		Headers: oapi.PageRedirectResponseHeaders{
			Location: "/",
		},
	}, nil
}

func (s *APIServer) GetSignin(ctx context.Context, request oapi.GetSigninRequestObject) (oapi.GetSigninResponseObject, error) {
	page, err := s.templateEx.ExecuteSignIn("")
	if err != nil {
		slog.ErrorContext(ctx, "error executing template", slog.Any("error", err))
		return oapi.GetSignin500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error executing template",
				ErrorDetails: ptr(err.Error()),
			},
		}, nil
	}

	return oapi.GetSignin200TexthtmlResponse{
		PageTexthtmlResponse: oapi.PageTexthtmlResponse{
			Body:          page,
			ContentLength: int64(page.Len()),
		},
	}, nil
}

func (s *APIServer) PostSignin(ctx context.Context, request oapi.PostSigninRequestObject) (oapi.PostSigninResponseObject, error) {
	login := string(request.Body.Login)
	password := request.Body.Password

	if s.config.User != login ||
		s.config.Password != password {
		return s.postSigninError(ctx, errors.New("wrong credentials, try again")), nil
	}

	v := PostSignInVisiter(func(w http.ResponseWriter) error {
		s.cookies.Create(ctx, w, string(login))
		w.Header().Add("Location", "/")
		w.WriteHeader(http.StatusSeeOther)

		return nil
	})

	return v, nil
}

func (s *APIServer) postSigninError(ctx context.Context, err error) oapi.PostSigninResponseObject {
	page, err2 := s.templateEx.ExecuteSignIn(err.Error())
	if err2 != nil {
		slog.ErrorContext(ctx, "error executing template", slog.Any("error", err2))
		return oapi.PostSignin500JSONResponse{
			PageErrorJSONResponse: oapi.PageErrorJSONResponse{
				Error:        "error executing template",
				ErrorDetails: ptr(errors.Join(err, err2).Error()),
			},
		}
	}

	return oapi.PostSignin200TexthtmlResponse{
		PageTexthtmlResponse: oapi.PageTexthtmlResponse{
			Body:          page,
			ContentLength: int64(page.Len()),
		},
	}
}

func (s *APIServer) GetLogout(ctx context.Context, request oapi.GetLogoutRequestObject) (oapi.GetLogoutResponseObject, error) {
	v := LogoutVisiter(func(w http.ResponseWriter) error {
		s.cookies.Delete(ctx, w, request.Params.SessionToken)
		w.Header().Add("Location", "/signin")
		w.WriteHeader(http.StatusSeeOther)

		return nil
	})

	return v, nil
}

func (s *APIServer) GetStaticStyleCss(ctx context.Context, request oapi.GetStaticStyleCssRequestObject) (oapi.GetStaticStyleCssResponseObject, error) {
	return oapi.GetStaticStyleCss200TextcssResponse{Body: bytes.NewReader(s.staticFiles.CSS)}, nil
}

func (s *APIServer) GetStaticFaviconIco(ctx context.Context, request oapi.GetStaticFaviconIcoRequestObject) (oapi.GetStaticFaviconIcoResponseObject, error) {
	return oapi.GetStaticFaviconIco200ImagexIconResponse{Body: bytes.NewReader(s.staticFiles.Favicon)}, nil
}

func ptr[T any](v T) *T {
	return &v
}

type PostSignInVisiter func(w http.ResponseWriter) error

func (v PostSignInVisiter) VisitPostSigninResponse(w http.ResponseWriter) error {
	return v(w)
}

type LogoutVisiter func(w http.ResponseWriter) error

func (v LogoutVisiter) VisitGetLogoutResponse(w http.ResponseWriter) error {
	return v(w)
}
