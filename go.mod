module kommandeur

go 1.15

require (
	github.com/BurntSushi/toml v0.3.1
	github.com/abbot/go-http-auth v0.0.0-00010101000000-000000000000
	github.com/go-check/check v0.0.0-00010101000000-000000000000
	github.com/gorilla/mux v1.7.3
	github.com/traefik/traefik/v2 v2.3.6

)

// Docker v19.03.6
replace github.com/docker/docker => github.com/docker/engine v1.4.2-0.20200204220554-5f6d6f3f2203

// Containous forks
replace (
	github.com/abbot/go-http-auth => github.com/containous/go-http-auth v0.4.1-0.20200324110947-a37a7636d23e
	github.com/go-check/check => github.com/containous/check v0.0.0-20170915194414-ca0bf163426a
	github.com/gorilla/mux => github.com/containous/mux v0.0.0-20181024131434-c33f32e26898
	github.com/mailgun/minheap => github.com/containous/minheap v0.0.0-20190809180810-6e71eb837595
	github.com/mailgun/multibuf => github.com/containous/multibuf v0.0.0-20190809014333-8b6c9a7e6bba
)
