module github.com/portainer/portainer/api

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.17
	github.com/asaskevich/govalidator v0.0.0-20200428143746-21a406dcc535
	github.com/boltdb/bolt v1.3.1
	github.com/coreos/go-semver v0.3.0
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/cli v20.10.7+incompatible
	github.com/docker/docker v20.10.7+incompatible
	github.com/g07cha/defender v0.0.0-20180505193036-5665c627c814
	github.com/go-git/go-git/v5 v5.3.0
	github.com/go-ldap/ldap/v3 v3.1.8
	github.com/gofrs/uuid v3.3.0+incompatible
	github.com/gorilla/mux v1.8.0
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.2
	github.com/joho/godotenv v1.3.0
	github.com/jpillora/chisel v0.0.0-20190724232113-f3a8df20e389
	github.com/json-iterator/go v1.1.10
	github.com/koding/websocketproxy v0.0.0-20181220232114-7ed82d81a28c
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.9.1
	github.com/portainer/docker-compose-wrapper v0.0.0-20210719115040-b82f1d4be772
	github.com/portainer/libcrypto v0.0.0-20190723020515-23ebe86ab2c2
	github.com/portainer/libhttp v0.0.0-20190806161843-ba068f58be33
	github.com/stretchr/testify v1.7.0
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/oauth2 v0.0.0-20201109201403-9fd604954f58
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	k8s.io/api v0.20.6
	k8s.io/apimachinery v0.20.6
	k8s.io/client-go v0.20.6
)

replace github.com/jaguilar/vt100 => github.com/tonistiigi/vt100 v0.0.0-20190402012908-ad4c4a574305

// replace github.com/portainer/docker-compose-wrapper => ../../docker-compose-wrapper
