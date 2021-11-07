// This code currently needs https://github.com/tailscale/tailscale/pull/3275

package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"tailscale.com/client/tailscale"
	"tailscale.com/tsnet"
)

func main() {
	os.Setenv("TAILSCALE_USE_WIP_CODE", "true")
	// TODO: comment out this line to avoid having to re-login each time you start this program
	os.Setenv("TS_LOGIN", "1")
	os.Setenv("HOME", "/perm/tsnetdemo")
	hostname := flag.String("hostname", "tsnetdemo", "tailscale hostname")
	allowedUser := flag.String("allowed_user", "", "the name of a tailscale user to allow")
	flag.Parse()
	s := &tsnet.Server{
		Hostname: *hostname,
	}
	log.Printf("starting tailscale listener on hostname %s", *hostname)
	ln, err := s.Listen("tcp", ":443")
	if err != nil {
		log.Fatal(err)
	}
	ln = tls.NewListener(ln, &tls.Config{
		GetCertificate: tailscale.GetCertificate,
	})
	httpsrv := &http.Server{
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			who, err := tailscale.WhoIs(r.Context(), r.RemoteAddr)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if who.UserProfile.LoginName != *allowedUser || *allowedUser == "" {
				err := fmt.Sprintf("you are logged in as %q, but -allowed_user flag does not match!", who.UserProfile.LoginName)
				log.Printf("forbidden: %v", err)
				http.Error(w, err, http.StatusForbidden)
				return
			}
			fmt.Fprintf(w, "hey there, %q! this message is served via the tsnet package from gokrazy!", who.UserProfile.LoginName)
		}),
	}
	log.Fatal(httpsrv.Serve(ln))
}
