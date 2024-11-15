package config

import (
	"crypto/x509"
	"log"

	"github.com/jackc/pgx"
)

var RootCAs *x509.CertPool
var DbConn *pgx.Conn
var AppKey string = "base64:qZCQSIA7VPk8Zxuc+lk/LJOeyoxTnU/hpesawf8gL2s="

func certPool() {
	var err error
	RootCAs, err = x509.SystemCertPool()
	if err != nil || RootCAs == nil {
		RootCAs = x509.NewCertPool()
		log.Println("Using new cert pool.")
	}
}

func Bootstrap() {
	certPool()
}
