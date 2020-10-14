module github.com/nginxinc/nginx-wrapper/plugins/example

go 1.15

replace github.com/nginxinc/nginx-wrapper/lib => ../../lib

require (
	github.com/nginxinc/nginx-wrapper/lib v0.0.0-00010101000000-000000000000
	github.com/go-eden/slf4go v1.0.7
)
