package cmds

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"pi-wegrzyn/storage"
	"pi-wegrzyn/utils"
)

type server struct {
	config    utils.Config
	db        *storage.DB
	templates templates
	cookies   cookies
}

func NewServer(config *utils.Config, templatesDir string, db *storage.DB) (*server, error) {
	templates, err := initTemplates(templatesDir)
	if err != nil {
		return nil, err
	}

	return &server{
		config:    *config,
		db:        db,
		templates: templates,
		cookies:   make(map[string]cookie),
	}, nil
}

func (s *server) isSignedIn(r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return false
	}

	session, exists := s.cookies[cookie.Value]
	if !exists || session.isExpired() {
		delete(s.cookies, cookie.Value)
		return false
	}

	return true
}

func (s *server) signInHtml(w http.ResponseWriter, r *http.Request) {
	if s.isSignedIn(r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	slog.InfoContext(r.Context(), fmt.Sprintf("%s%s (%s / %s)", r.Host, r.URL, r.Method, r.Proto))

	if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")

		if res, err := regexp.MatchString("^([-_.a-zA-Z0-9]){2,32}$", login); err == nil && res && s.config.Users[login] == password {
			s.cookies.createCookie(r.Context(), w, login)
			slog.InfoContext(r.Context(), "user signed in successfully", slog.Any("username", login))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		s.templates["signin.html"].Execute(w, "Wrong credentials, try again")
		return
	}

	s.templates["signin.html"].Execute(w, "")
}

func (s *server) indexHtml(w http.ResponseWriter, r *http.Request) {
	if r.RequestURI != "/" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	slog.InfoContext(r.Context(), fmt.Sprintf("%s%s (%s / %s)", r.Host, r.URL, r.Method, r.Proto))

	devices, err := s.db.Devices(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "database error", slog.Any("error", err))
		s.templates["index.html"].Execute(w, nil)
		return
	}

	s.templates["index.html"].Execute(w, devices)
}

func (s *server) newHtml(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), fmt.Sprintf("%s%s (%s / %s)", r.Host, r.URL, r.Method, r.Proto))

	if r.Method == http.MethodPost {
		hostname := strings.TrimSpace(r.FormValue("hostname"))
		ip := strings.TrimSpace(r.FormValue("ip"))
		login := strings.TrimSpace(r.FormValue("login"))
		password := r.FormValue("password")

		ipType, err := strconv.Atoi(strings.TrimSpace(r.FormValue("ip-type")))
		if err != nil {
			slog.ErrorContext(r.Context(), "error while parsing IP type", slog.Any("error", err))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if err = validateFormInput(&ipType, &hostname, &ip, &login); err != nil {
			slog.ErrorContext(r.Context(), "validation error", slog.Any("error", err))
			s.templates["new.html"].Execute(w, NewEdit{"New", storage.Device{Hostname: hostname, IPAddress: ip, Login: login}, ipType, prettifyError(err)})
			return
		}

		key, err := retrieveSSHKey(r)
		if err != nil {
			s.templates["new.html"].Execute(w, NewEdit{"New", storage.Device{Hostname: hostname, IPAddress: ip, Login: login}, ipType, prettifyError(err)})
			return
		}

		if err = s.db.CreateDevice(r.Context(), storage.Device{Hostname: hostname, IPAddress: ip, Login: login, Password: password, Keyfile: key}); err != nil {
			slog.ErrorContext(r.Context(), "database error", slog.Any("error", err))
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.templates["new.html"].Execute(w, NewEdit{"New", storage.Device{}, 4, ""})
}

func (s *server) editHtml(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), fmt.Sprintf("%s%s (%s / %s)", r.Host, r.URL, r.Method, r.Proto))

	id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("edit-id")))
	if err != nil {
		slog.ErrorContext(r.Context(), "wrong device ID provided", slog.Any("deviceID", r.FormValue("edit-id")), slog.Any("error", err))
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	device, err := s.db.Device(r.Context(), uint(id))
	if err != nil {
		slog.ErrorContext(r.Context(), "database error", slog.Any("error", err))
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var deviceIpType int
	if net.ParseIP(device.IPAddress).To4() != nil {
		deviceIpType = 4
	} else {
		deviceIpType = 6
	}

	if r.Method == http.MethodPost {
		hostname := strings.TrimSpace(r.FormValue("hostname"))
		ip := strings.TrimSpace(r.FormValue("ip"))
		login := strings.TrimSpace(r.FormValue("login"))

		ipType, err := strconv.Atoi(strings.TrimSpace(r.FormValue("ip-type")))
		if err != nil {
			slog.ErrorContext(r.Context(), "error while parsing IP type", slog.Any("error", err))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		err = validateFormInput(&ipType, &hostname, &ip, &login)
		if err != nil {
			slog.ErrorContext(r.Context(), "validation error", slog.Any("error", err))
			s.templates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, prettifyError(err)})
			return
		}

		newKey, err := retrieveSSHKey(r)
		if err != nil {
			slog.InfoContext(r.Context(), "error occurred while retrieving SSH key", slog.Any("error", err))
			s.templates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, prettifyError(err)})
			return
		}

		if len(newKey) > 0 {
			device.Keyfile = newKey
		} else if r.FormValue("key-clear") != "" {
			device.Keyfile = []byte{}
		}

		if strings.TrimSpace(r.FormValue("password-clear")) != "" {
			device.Password = r.FormValue("password")
		}

		device.Hostname, device.IPAddress, device.Login = hostname, ip, login

		if err := s.db.UpdateDevice(r.Context(), device); err != nil {
			slog.ErrorContext(r.Context(), "database error", slog.Any("error", err))
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.templates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, ""})
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	slog.InfoContext(r.Context(), fmt.Sprintf("%s%s (%s / %s)", r.Host, r.URL, r.Method, r.Proto))

	if r.Method == http.MethodPost {
		id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("delete-id")))
		if err != nil {
			slog.ErrorContext(r.Context(), "wrong device ID provided", slog.Any("deviceID", r.FormValue("delete-id")), slog.Any("error", err))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if err = s.db.DeleteDevice(r.Context(), uint(id)); err != nil {
			slog.ErrorContext(r.Context(), "database error", slog.Any("error", err))
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) logOut(w http.ResponseWriter, r *http.Request) {
	s.cookies.deleteCookie(r.Context(), w, r)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func (s *server) protected(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.isSignedIn(r) {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		handler(w, r)
	}
}

func (s *server) Start(ctx context.Context, staticDir string) error {
	slog.InfoContext(ctx, "Frontend module started")

	http.HandleFunc("/signin", s.signInHtml)
	http.HandleFunc("/", s.protected(s.indexHtml))
	http.HandleFunc("/new", s.protected(s.newHtml))
	http.HandleFunc("/edit", s.protected(s.editHtml))
	http.HandleFunc("/delete", s.protected(s.delete))
	http.HandleFunc("/logout", s.logOut)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	slog.InfoContext(ctx, fmt.Sprintf("Serving HTTP on 0.0.0.0 port %[1]d (http://0.0.0.0:%[1]d/) ...", s.config.Port))

	return http.ListenAndServe(fmt.Sprintf(":%d", s.config.Port), nil)
}

func validateFormInput(ipType *int, hostname *string, ip *string, login *string) error {
	if *ipType != 4 && *ipType != 6 {
		*ipType = 4
		*hostname = ""
		*ip = ""
		*login = ""
		return errors.New("wrong IP type provided")
	}

	if *hostname == "" || *ip == "" || *login == "" {
		return errors.New("empty fields")
	}

	if res, err := regexp.MatchString(LoginPattern, *login); err != nil || !res {
		*login = ""
		return errors.New("wrong login")
	}

	if net.ParseIP(*ip) == nil {
		*ip = ""
		return errors.New("wrong IP address")
	}

	return nil
}

func retrieveSSHKey(r *http.Request) (key []byte, err error) {
	keyFile, keyFileHeader, err := r.FormFile("key")
	if err != nil {
		if err == http.ErrMissingFile {
			return nil, nil
		}

		return nil, err
	}

	if keyFileHeader.Size > 20480 || keyFileHeader.Size == 0 {
		return nil, errors.New("wrong key size")
	}

	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, keyFile); err != nil {
		return nil, err
	}
	defer keyFile.Close()

	key = buf.Bytes()

	return key, nil
}

func prettifyError(err error) string {
	return strings.ToUpper(err.Error()[:1]) + err.Error()[1:] + ", try again"
}
