package templates

import (
	"errors"
	"testing"
)

func TestForm_Validate(t *testing.T) {
	tcs := []struct {
		name string
		form Form
		err  error
	}{
		{
			name: "valid",
			form: Form{
				Hostname: "hostname",
				Ip:       "127.0.0.1",
				Login:    "login",
				IPType:   4,
			},
			err: nil,
		},
		{
			name: "wrong IP type",
			form: Form{
				IPType: 3,
			},
			err: errors.New("wrong IP type provided"),
		},
		{
			name: "empty fields",
			form: Form{
				Hostname: "",
				Ip:       "",
				Login:    "",
				IPType:   4,
			},
			err: errors.New("empty fields"),
		},
		{
			name: "wrong login",
			form: Form{
				Hostname: "hostname",
				Ip:       "127.0.0.1",
				Login:    "l",
				IPType:   4,
			},
			err: errors.New("wrong login"),
		},
		{
			name: "wrong IP address",
			form: Form{
				Hostname: "hostname",
				Ip:       "127.0.0.1.1abc",
				Login:    "login",
				IPType:   4,
			},
			err: errors.New("wrong IP address"),
		},
		{
			name: "wrong key size",
			form: Form{
				Hostname: "hostname",
				Ip:       "127.0.0.1",
				Login:    "login",
				IPType:   4,
				Key:      make([]byte, 20481),
			},
			err: errors.New("wrong key size"),
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.form.Validate()

			if err == nil {
				if tc.err != nil {
					t.Errorf("expected error %v, got nil", tc.err)
				}

				return
			}

			if err.Error() != tc.err.Error() {
				t.Errorf("expected error %v, got %v", tc.err, err)
			}
		})
	}
}
