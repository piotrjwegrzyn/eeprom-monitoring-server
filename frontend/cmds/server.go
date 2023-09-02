package cmds

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"pi-wegrzyn/utils"
)

type server struct {
	config    utils.Config
	templates templates
	cookies   cookies
}

func NewServer(config *utils.Config, templatesDir string) (*server, error) {
	templates, err := initTemplates(templatesDir)
	if err != nil {
		return nil, err
	}

	return &server{
		cookies:   make(map[string]cookie),
		config:    *config,
		templates: templates,
	}, nil
}

func (s *server) isSignedIn(w http.ResponseWriter, r *http.Request) bool {
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
	if s.isSignedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")

		if res, _ := regexp.MatchString("^([-_.a-zA-Z0-9]){2,32}$", login); res && ((s.config).Users)[login] == password {
			s.cookies.createCookie(w, login)
			log.Printf("User %s signed in successfully (cookie: %s)\n", login, s.cookies[login])
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

	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	database, err := utils.ConnectToDatabase(&s.config.MySQL)
	if err != nil {
		log.Printf("Error while opening database: %v\n", err)
		s.templates["index.html"].Execute(w, nil)
		return
	}

	devices, err := database.GetDevices()
	if err != nil {
		log.Printf("Database error: %v\n", err)
		s.templates["index.html"].Execute(w, nil)
		return
	}

	s.templates["index.html"].Execute(w, devices)
	defer database.Close()
}

func (s *server) newHtml(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	if r.Method == http.MethodPost {
		hostname := strings.TrimSpace(r.FormValue("hostname"))
		ip := strings.TrimSpace(r.FormValue("ip"))
		login := strings.TrimSpace(r.FormValue("login"))
		password := r.FormValue("password")

		ipType, err := strconv.Atoi(strings.TrimSpace(r.FormValue("ip-type")))
		if err != nil {
			log.Printf("Error while parsing IP type: %v", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		if err = validateFormInput(&ipType, &hostname, &ip, &login); err != nil {
			log.Printf("Unsuccessful validation: %v\n", err)
			s.templates["new.html"].Execute(w, NewEdit{"New", utils.Device{Hostname: hostname, IP: ip, Login: login}, ipType, prettifyError(err)})
			return
		}

		key, err := retriveSSHKey(r)
		if err != nil {
			s.templates["new.html"].Execute(w, NewEdit{"New", utils.Device{Hostname: hostname, IP: ip, Login: login}, ipType, prettifyError(err)})
			return
		}

		database, err := utils.ConnectToDatabase(&s.config.MySQL)
		if err != nil {
			log.Printf("Error while opening database: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		defer database.Close()

		if err = database.InsertDevice(utils.Device{Hostname: hostname, IP: ip, Login: login, Password: password, Key: key}); err != nil {
			log.Printf("Database error: %v\n", err)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.templates["new.html"].Execute(w, NewEdit{"New", utils.Device{}, 4, ""})
}

func (s *server) editHtml(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("edit-id")))
	if err != nil {
		log.Printf("Error while parsing device ID: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	database, err := utils.ConnectToDatabase(&s.config.MySQL)
	if err != nil {
		log.Printf("Error while opening database: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer database.Close()

	device, err := database.GetDevice(id)
	if err != nil {
		log.Printf("Database error: %v\n", err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var deviceIpType int
	if net.ParseIP(device.IP).To4() != nil {
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
			log.Printf("Error while parsing IP type: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		err = validateFormInput(&ipType, &hostname, &ip, &login)
		if err != nil {
			log.Printf("Unsuccessful validation: %v\n", err)
			s.templates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, prettifyError(err)})
			return
		}

		newKey, err := retriveSSHKey(r)
		if err != nil {
			log.Printf("Error occurred while retriving SSH key: %v\n", err)
			s.templates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, prettifyError(err)})
			return
		}

		if len(newKey) > 0 {
			device.Key = newKey
		} else if r.FormValue("key-clear") != "" {
			device.Key = []byte{}
		}

		if strings.TrimSpace(r.FormValue("password-clear")) != "" {
			device.Password = r.FormValue("password")
		}

		device.Hostname, device.IP, device.Login = hostname, ip, login

		if err := database.UpdateDevice(device); err != nil {
			log.Printf("Database error: %v\n", err)
		}

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	s.templates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, ""})
}

func (s *server) delete(w http.ResponseWriter, r *http.Request) {
	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	if r.Method == http.MethodPost {
		id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("delete-id")))
		if err != nil {
			log.Printf("Wrong device ID was provided (%s)\n", r.FormValue("delete-id"))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		database, err := utils.ConnectToDatabase(&s.config.MySQL)
		if err != nil {
			log.Printf("Error while opening database: %v\n", err)
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		defer database.Close()

		if err = database.DeleteDevice(id); err != nil {
			log.Printf("Database error: %v\n", err)
		}
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *server) logOut(w http.ResponseWriter, r *http.Request) {
	s.cookies.deleteCookie(w, r)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func (s *server) protected(handler func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if !s.isSignedIn(w, r) {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}

		handler(w, r)
	}
}

func (s *server) Start(staticDir string) error {
	log.Println("Frontend module started")

	http.HandleFunc("/signin", s.signInHtml)
	http.HandleFunc("/", s.protected(s.indexHtml))
	http.HandleFunc("/new", s.protected(s.newHtml))
	http.HandleFunc("/edit", s.protected(s.editHtml))
	http.HandleFunc("/delete", s.protected(s.delete))
	http.HandleFunc("/logout", s.logOut)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	log.Printf("Serving HTTP on 0.0.0.0 port %[1]d (http://0.0.0.0:%[1]d/) ...\n", s.config.Port)

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

	if res, _ := regexp.MatchString(LoginPattern, *login); !res {
		*login = ""
		return errors.New("wrong login")
	}

	if net.ParseIP(*ip) == nil {
		*ip = ""
		return errors.New("wrong IP address")
	}

	return nil
}

func retriveSSHKey(r *http.Request) (key []byte, err error) {
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
