module pi-wegrzyn/backend

go 1.19

replace pi-wegrzyn/common => ../common

require (
	golang.org/x/crypto v0.1.0
	pi-wegrzyn/common v0.0.0-00010101000000-000000000000
)

require (
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	golang.org/x/sys v0.1.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
