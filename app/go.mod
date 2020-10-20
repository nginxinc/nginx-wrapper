module github.com/nginxinc/nginx-wrapper/app

go 1.15

replace github.com/nginxinc/nginx-wrapper/lib => ../lib

require (
	github.com/nginxinc/nginx-wrapper/lib v0.0.2

	github.com/davecgh/go-spew v1.1.1
	github.com/elliotchance/orderedmap v1.3.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/go-eden/common v0.1.8 // indirect
	github.com/go-eden/slf4go v1.0.7
	github.com/google/uuid v1.1.2
	github.com/iancoleman/strcase v0.1.2
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/mapstructure v1.3.3 // indirect
	github.com/oleiade/reflections v1.0.0 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pkg/errors v0.9.1
	github.com/sirupsen/logrus v1.7.0
	github.com/spf13/afero v1.4.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.0
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	github.com/tevino/abool v1.2.0

	golang.org/x/sys v0.0.0-20201014080544-cc95f250f6bc // indirect

	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/oleiade/reflections.v1 v1.0.0
)
