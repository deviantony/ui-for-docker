module github.com/portainer/portainer/api

go 1.16

require (
	github.com/Microsoft/go-winio v0.4.16
	github.com/asaskevich/govalidator v0.0.0-20190424111038-f61b66f89f4a
	github.com/boltdb/bolt v1.3.1
	github.com/containerd/containerd v1.3.1 // indirect
	github.com/coreos/go-semver v0.3.0
	github.com/dchest/uniuri v0.0.0-20160212164326-8902c56451e9
	github.com/dgrijalva/jwt-go v3.2.0+incompatible
	github.com/docker/cli v0.0.0-20191126203649-54d085b857e9
	github.com/docker/docker v0.0.0-00010101000000-000000000000
	github.com/g07cha/defender v0.0.0-20180505193036-5665c627c814
	github.com/go-git/go-git/v5 v5.3.0
	github.com/go-ldap/ldap/v3 v3.1.8
	github.com/gofrs/uuid v3.2.0+incompatible
	github.com/gorilla/mux v1.7.3
	github.com/gorilla/securecookie v1.1.1
	github.com/gorilla/websocket v1.4.1
	github.com/joho/godotenv v1.3.0
	github.com/jpillora/chisel v0.0.0-20190724232113-f3a8df20e389
	github.com/json-iterator/go v1.1.8
	github.com/koding/websocketproxy v0.0.0-20181220232114-7ed82d81a28c
	github.com/mattn/go-shellwords v1.0.6 // indirect
	github.com/mitchellh/mapstructure v1.1.2 // indirect
	github.com/orcaman/concurrent-map v0.0.0-20190826125027-8c72a8bb44f6
	github.com/pkg/errors v0.9.1
	github.com/portainer/libcompose v0.5.3
	github.com/portainer/libcrypto v0.0.0-20210422035235-c652195c5c3a
	github.com/portainer/libhttp v0.0.0-20190806161843-ba068f58be33
	github.com/stretchr/testify v1.6.1
	golang.org/x/crypto v0.0.0-20210322153248-0c34fe9e7dc2
	golang.org/x/oauth2 v0.0.0-20190604053449-0f29369cfe45
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	gopkg.in/yaml.v3 v3.0.0-20200313102051-9f266ea9e77c
	k8s.io/api v0.17.2
	k8s.io/apimachinery v0.17.2
	k8s.io/client-go v0.17.2
)

replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203

replace golang.org/x/sys => golang.org/x/sys v0.0.0-20190826190057-c7b8b68b1456
