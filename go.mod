module github.com/nmasse-itix/keycloak-realm-import

go 1.15

require (
	github.com/cloudtrust/common-service v2.3.2+incompatible
	github.com/cloudtrust/keycloak-client/v3 v3.0.0
	github.com/fsnotify/fsnotify v1.4.9 // indirect
	github.com/google/uuid v1.0.0
	github.com/magiconair/properties v1.8.4 // indirect
	github.com/mitchellh/go-homedir v1.1.0
	github.com/mitchellh/mapstructure v1.4.1 // indirect
	github.com/pelletier/go-toml v1.8.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rakyll/statik v0.1.7
	github.com/spf13/afero v1.5.1 // indirect
	github.com/spf13/cast v1.3.1 // indirect
	github.com/spf13/cobra v1.1.1
	github.com/spf13/jwalterweatherman v1.1.0 // indirect
	github.com/spf13/viper v1.7.1
	golang.org/x/oauth2 v0.0.0-20210113205817-d3ed898aa8a3 // indirect
	golang.org/x/sys v0.0.0-20210113181707-4bcb84eeeb78 // indirect
	gopkg.in/ini.v1 v1.62.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/cloudtrust/keycloak-client/v3 => github.com/nmasse-itix/keycloak-client/v3 v3.0.0
