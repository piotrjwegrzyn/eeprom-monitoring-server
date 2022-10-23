package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"path"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	"time"

	common "pi-wegrzyn/common"

	"github.com/google/uuid"
)

func GetPattern(version int) string {
	if version == 6 {
		return Ipv6Pattern
	}
	return Ipv4Pattern
}

type NewEdit struct {
	Action       string
	Device       common.Device
	IpVersion    int
	ErrorMessage string
}

type Session struct {
	Login  string
	Expiry time.Time
}

func (s Session) isExpired() bool {
	return s.Expiry.Before(time.Now())
}

var Sessions map[string]Session
var PageTemplates map[string]*template.Template
var PageConfig *common.Config

const Ipv4Pattern string = `^((25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])\.){3,3}(25[0-5]|(2[0-4]|1{0,1}[0-9]){0,1}[0-9])$`
const Ipv6Pattern string = `^((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])(\.(25[0-5]|2[0-4][0-9]|1[0-9][0-9]|[1-9]?[0-9])){3}))|:)))(%.+)?$`
const LoginPattern string = `^[a-zA-Z][-a-zA-Z0-9\.\_]*[a-zA-Z0-9]$`

func InitTemplates(templatesDir *string) (err error) {
	functions := template.FuncMap{
		"ToUpper":    strings.ToUpper,
		"ToLower":    strings.ToLower,
		"GetPattern": GetPattern,
	}
	for _, t := range []string{"index.html", "new.html", "signin.html"} {
		PageTemplates[t], err = template.New(t).Funcs(functions).ParseFiles(path.Join(*templatesDir, t))
		if err != nil {
			return err
		}
	}

	return nil
}

func CreateCookie(w http.ResponseWriter, username string) {
	token := uuid.NewString()
	expiration := time.Now().Add(15 * time.Minute)

	Sessions[token] = Session{
		Login:  username,
		Expiry: expiration,
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   token,
		Expires: expiration,
	})
}

func DeleteCookie(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err == http.ErrNoCookie {
			http.Redirect(w, r, "/signin", http.StatusSeeOther)
			return
		}
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	token := cookie.Value

	delete(Sessions, token)

	http.SetCookie(w, &http.Cookie{
		Name:    "session_token",
		Value:   "",
		Expires: time.Now(),
	})
}

func IsSignedIn(w http.ResponseWriter, r *http.Request) bool {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		if err != http.ErrNoCookie {
			log.Println(err)
		}
		return false
	}
	token := cookie.Value

	userSession, exists := Sessions[token]
	if !exists {
		return false
	} else if userSession.isExpired() {
		delete(Sessions, token)
		return false
	}
	return true
}

func ValidateFormInput(ipType *int, hostname *string, ip *string, login *string) error {
	if *ipType != 4 && *ipType != 6 {
		*ipType = 4
		*hostname = ""
		*ip = ""
		*login = ""
		return fmt.Errorf("Wrong IP type provided")
	}
	if *hostname == "" || *ip == "" || *login == "" {
		return fmt.Errorf("Empty fields, try again")
	}
	if res, _ := regexp.MatchString(LoginPattern, *login); !res {
		*login = ""
		return fmt.Errorf("Wrong login, try again")
	}
	if net.ParseIP(*ip) == nil {
		*ip = ""
		return fmt.Errorf("Wrong IP address, try again")
	}

	return nil
}

func GetKey(r *http.Request) (key []byte, msg error) {
	keyFile, keyFileHeader, err := r.FormFile("key")
	if err != nil {
		if err != http.ErrMissingFile {
			log.Println(err)
			msg = fmt.Errorf("Error with retriving file")
			return
		}
	} else {
		if keyFileHeader.Size > 20480 || keyFileHeader.Size == 0 {
			msg = fmt.Errorf("Wrong key size, try again")
			return
		}
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, keyFile); err != nil {
			log.Println(err)
			msg = fmt.Errorf("Error with reading file")
			return
		} else {
			key = buf.Bytes()
		}
		defer keyFile.Close()
	}
	return
}

func SignInHtml(w http.ResponseWriter, r *http.Request) {
	if IsSignedIn(w, r) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	if r.Method == http.MethodPost {
		login := r.FormValue("login")
		password := r.FormValue("password")

		if res, _ := regexp.MatchString("^([-_.a-zA-Z0-9]){2,32}$", login); res && ((*PageConfig).Users)[login] == password {
			CreateCookie(w, login)
			log.Printf("User %s signed in successfully (cookie: %s)\n", login, Sessions[login])
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		} else {
			PageTemplates["signin.html"].Execute(w, "Wrong credentials, try again")
			return
		}
	}

	PageTemplates["signin.html"].Execute(w, "")
}

func IndexHtml(w http.ResponseWriter, r *http.Request) {
	if !IsSignedIn(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	if r.RequestURI != "/" {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	database, err := common.ConnectToDatabase(&PageConfig.Database)
	if err != nil {
		log.Println("Error while opening database")
		PageTemplates["index.html"].Execute(w, nil)
		return
	}

	devices, err := common.GetDevices(database)
	if err != nil {
		log.Println("Error while operating with database")
		PageTemplates["index.html"].Execute(w, nil)
		return
	}

	PageTemplates["index.html"].Execute(w, devices)
	defer database.Close()
}

func NewHtml(w http.ResponseWriter, r *http.Request) {
	if !IsSignedIn(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	if r.Method == http.MethodPost {
		hostname := strings.TrimSpace(r.FormValue("hostname"))
		ip := strings.TrimSpace(r.FormValue("ip"))
		login := strings.TrimSpace(r.FormValue("login"))
		password := r.FormValue("password")

		ipType, err := strconv.Atoi(strings.TrimSpace(r.FormValue("ip-type")))
		if err != nil {
			log.Println("Error while parsing IP type")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		err = ValidateFormInput(&ipType, &hostname, &ip, &login)
		if err != nil {
			PageTemplates["new.html"].Execute(w, NewEdit{"New", common.Device{Hostname: hostname, Ip: ip, Login: login}, ipType, err.Error()})
			return
		}

		key, err := GetKey(r)
		if err != nil {
			PageTemplates["new.html"].Execute(w, NewEdit{"New", common.Device{Hostname: hostname, Ip: ip, Login: login}, ipType, err.Error()})
			return
		}

		database, err := common.ConnectToDatabase(&PageConfig.Database)
		if err != nil {
			log.Println("Error while opening database")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		err = common.InsertDevice(database, common.Device{Hostname: hostname, Ip: ip, Login: login, Password: password, Key: key})
		if err != nil {
			log.Println(err)
		}

		defer database.Close()

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	PageTemplates["new.html"].Execute(w, NewEdit{"New", common.Device{}, 4, ""})
}

func EditHtml(w http.ResponseWriter, r *http.Request) {
	if !IsSignedIn(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("edit-id")))
	if err != nil {
		log.Println("Error while parsing device ID")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	database, err := common.ConnectToDatabase(&PageConfig.Database)
	if err != nil {
		log.Println("Error while opening database")
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	defer database.Close()

	device, err := common.GetDevice(database, id)
	if err != nil {
		log.Println(err)
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	var deviceIpType int
	if net.ParseIP(device.Ip).To4() != nil {
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
			log.Println("Error while parsing IP type")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		err = ValidateFormInput(&ipType, &hostname, &ip, &login)
		if err != nil {
			PageTemplates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, err.Error()})
			return
		}

		newKey, err := GetKey(r)
		if err != nil {
			PageTemplates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, err.Error()})
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

		device.Hostname, device.Ip, device.Login = hostname, ip, login

		common.UpdateDevice(database, device)

		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	PageTemplates["new.html"].Execute(w, NewEdit{"Edit", device, deviceIpType, ""})
}

func Delete(w http.ResponseWriter, r *http.Request) {
	if !IsSignedIn(w, r) {
		http.Redirect(w, r, "/signin", http.StatusSeeOther)
		return
	}

	log.Printf("%s%s (%s / %s)\n", r.Host, r.URL, r.Method, r.Proto)

	if r.Method == http.MethodPost {
		id, err := strconv.Atoi(strings.TrimSpace(r.FormValue("delete-id")))
		if err != nil {
			log.Printf("Wrong device ID was provided (%s)\n", r.FormValue("delete-id"))
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		database, err := common.ConnectToDatabase(&PageConfig.Database)
		if err != nil {
			log.Println("Error while opening database")
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}

		err = common.DeleteDevice(database, id)
		if err != nil {
			log.Println(err)
		}

		defer database.Close()
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func LogOut(w http.ResponseWriter, r *http.Request) {
	DeleteCookie(w, r)
	http.Redirect(w, r, "/signin", http.StatusSeeOther)
}

func StartServer(serverConfig *common.Config, templatesDir *string, staticDir *string) error {
	Sessions = make(map[string]Session)
	PageTemplates = make(map[string]*template.Template)
	if err := InitTemplates(templatesDir); err != nil {
		return err
	}
	PageConfig = serverConfig

	http.HandleFunc("/signin", SignInHtml)
	http.HandleFunc("/", IndexHtml)
	http.HandleFunc("/new", NewHtml)
	http.HandleFunc("/edit", EditHtml)
	http.HandleFunc("/delete", Delete)
	http.HandleFunc("/logout", LogOut)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(*staticDir))))

	log.Printf("Serving HTTP on 0.0.0.0 port %[1]d (http://0.0.0.0:%[1]d/) ...\n", serverConfig.Port)

	return http.ListenAndServe(fmt.Sprintf(":%d", serverConfig.Port), nil)
}
