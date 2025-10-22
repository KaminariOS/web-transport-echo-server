package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"flag"
	"fmt"
	"io"
	"log"
	"math/big"
	"net"
	"net/http"
	"time"

	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

var (
	apiAddressFlag          = flag.String("api-address", "0.0.0.0:8000", "api bind address formatted as <host>:<port>")
	webTransportAddressFlag = flag.String("web-transport-address", "0.0.0.0:4443", "web transport bind address formatted as <host>:<port>")
)

func main() {
	cert, err := generateTLSCertificate()
	if err != nil {
		log.Fatalf("failed to generate TLS certificate", err)
	}
	go func() {
		err := listenAndServeAPI(cert)
		if err != nil {
			log.Fatalf("API server closed: %", err)
		}
	}()
	err = listenAndServeWebTransport(cert)
	if err != nil {
		log.Fatalf("WebTransport server closed: %", err)
	}
}

// listenAndServeAPI starts the API server.
func listenAndServeAPI(cert *tls.Certificate) error {
	server := http.Server{
		Addr: *apiAddressFlag,
	}
	certHash := fmt.Sprintf("%x", sha256.Sum256(cert.Leaf.Raw))
	server.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "*")
		w.Header().Set("Access-Control-Allow-Headers", "*")
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
		} else {
			w.Write([]byte(certHash))
		}
	})
	log.Printf("API server listening on: %s\n", server.Addr)
	return server.ListenAndServe()
}

// listenAndServeWebTransport starts the WebTransport server.
func listenAndServeWebTransport(cert *tls.Certificate) error {
	server := webtransport.Server{
		CheckOrigin: func(r *http.Request) bool {
			return true // allow connections from all origins
		},
	}
	server.H3 = http3.Server{
		Addr:    *webTransportAddressFlag,
		Handler: webTransportEchoHandler(&server),
		TLSConfig: &tls.Config{
			Certificates: []tls.Certificate{*cert},
			NextProtos:   []string{http3.NextProtoH3},
		},
	}
	log.Printf("WebTransport server listening on: %s\n", server.H3.Addr)
	return server.ListenAndServe()
}

// webTransportEchoHandler echoes stream data back to the sender.
func webTransportEchoHandler(server *webtransport.Server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		conn, err := server.Upgrade(w, r)
		if err != nil {
			log.Printf("connection upgrade failed: %v", err)
			w.WriteHeader(500)
			return
		}
		log.Println("connection accepted")
		for {
			str, err := conn.AcceptStream(r.Context())
			if err != nil {
				break
			}
			_, err = io.CopyBuffer(str, str, make([]byte, 100))
			if err != nil {
				break
			}
		}
	})
}

// generateTLSCertificate returns a new self signed TLS certifiate.
//
// WebTransport allows connections using self signed certificates when
// the `serverCertificateHashes` option is used.
//
// https://developer.mozilla.org/en-US/docs/Web/API/WebTransport/WebTransport#servercertificatehashes
func generateTLSCertificate() (*tls.Certificate, error) {
	certTempl := &x509.Certificate{
		SerialNumber:          big.NewInt(2025),
		Subject:               pkix.Name{},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
		DNSNames:              []string{"localhost"},
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	caPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, err
	}
	caBytes, err := x509.CreateCertificate(rand.Reader, certTempl, certTempl, caPrivateKey.Public(), caPrivateKey)
	if err != nil {
		return nil, err
	}
	ca, err := x509.ParseCertificate(caBytes)
	if err != nil {
		return nil, err
	}
	return &tls.Certificate{
		Certificate: [][]byte{ca.Raw},
		PrivateKey:  caPrivateKey,
		Leaf:        ca,
	}, nil
}
