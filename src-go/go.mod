module gate-master

go 1.21

toolchain go1.24.2

require (
	github.com/davecheney/i2c v0.0.0-20140823063045-caf08501bef2
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/gorilla/securecookie v1.1.2
	github.com/mattn/go-sqlite3 v1.14.24
	gocv.io/x/gocv opencv-4.6.0
	golang.org/x/crypto v0.31.0
	gopkg.in/mail.v2 v2.3.1
)

require gopkg.in/alexcesaro/quotedprintable.v3 v3.0.0-20150716171945-2caba252f4dc // indirect
