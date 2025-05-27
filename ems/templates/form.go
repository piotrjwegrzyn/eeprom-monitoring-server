package templates

import (
	"bytes"
	"errors"
	"log/slog"
	"mime/multipart"
	"net"
	"regexp"
	"strconv"
)

const LoginPattern string = `^[a-zA-Z][\-a-zA-Z0-9_\.]*[a-zA-Z0-9]$`

type Form struct {
	Hostname string
	Ip       string
	IPType   int
	Login    string
	Password *string
	Key      []byte

	EditId        uint
	PasswordClear *string
	KeyClear      *string
}

func (f *Form) Validate() error {
	if f.IPType != 4 && f.IPType != 6 {
		f.IPType = 4
		return errors.New("wrong IP type provided")
	}

	if f.Hostname == "" || f.Ip == "" || f.Login == "" {
		return errors.New("empty fields")
	}

	if res, err := regexp.MatchString(LoginPattern, f.Login); err != nil || !res {
		f.Login = ""
		return errors.New("wrong login")
	}

	ip := net.ParseIP(f.Ip)
	if ip == nil || (ip.To4() == nil && f.IPType == 4) || (ip.To16() == nil && f.IPType == 6) {
		f.Ip = ""
		return errors.New("wrong IP address")
	}

	if len(f.Key) > 20480 {
		f.Key = nil
		return errors.New("wrong key size")
	}

	return nil
}

func ParseForm(formReader *multipart.Reader) (*Form, error) {
	form := Form{}

	for {
		part, err := formReader.NextPart()
		if err != nil {
			break
		}
		defer func() {
			if err := part.Close(); err != nil {
				slog.Error("cannot close part", slog.Any("error", err))
			}
		}()

		buf := new(bytes.Buffer)
		if _, err = buf.ReadFrom(part); err != nil {
			return &Form{}, err
		}

		switch part.FormName() {
		case "hostname":
			form.Hostname = buf.String()
		case "ip":
			form.Ip = buf.String()
		case "ip-type":
			form.IPType, err = strconv.Atoi(buf.String())
			if err != nil {
				return &Form{}, err
			}
		case "login":
			form.Login = buf.String()
		case "password":
			password := buf.String()
			form.Password = &password
		case "key":
			form.Key = buf.Bytes()
		case "edit-id":
			editId, err := strconv.ParseUint(buf.String(), 10, 32)
			if err != nil {
				return &Form{}, err
			}
			form.EditId = uint(editId)
		case "password-clear":
			passwordClear := buf.String()
			form.PasswordClear = &passwordClear
		case "key-clear":
			keyClear := buf.String()
			form.KeyClear = &keyClear
		default:
			slog.Warn("unknown form field, skipping", slog.Any("formName", part.FormName()), slog.Any("fileName", part.FileName()))
		}
	}

	return &form, nil
}
