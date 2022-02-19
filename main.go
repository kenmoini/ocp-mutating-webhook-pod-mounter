package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/golang/glog"
	"github.com/kenmoini/ocp-mutating-webhook-pod-mounter/pkg/admission"
	"github.com/kenmoini/ocp-mutating-webhook-pod-mounter/pkg/shared"
	"github.com/sirupsen/logrus"
	admissionv1 "k8s.io/api/admission/v1"
)

func main() {
	setLogger()

	var parameters WhSvrParameters

	// get command line parameters
	flag.IntVar(&parameters.port, "port", 8443, "Webhook server port.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.StringVar(&parameters.cfgFile, "sidecarCfgFile", "/etc/webhook/config/config.yaml", "File containing the mutation configuration.")
	flag.Parse()

	// handle our core application
	http.HandleFunc("/validate-pods", ServeValidatePods)
	http.HandleFunc("/mutate-pods", ServeMutatePods)
	http.HandleFunc("/health", ServeHealth)

	// Consume the configuration file
	cfg, err := shared.LoadConfig(parameters.cfgFile)
	if err != nil {
		glog.Errorf("Failed to load configuration: %v", err)
	}

	shared.CfgFile = cfg

	// start the server
	// listens to clear text http on port 8080 unless TLS env var is set to "true"
	if os.Getenv("TLS") == "true" {
		cert := parameters.certFile
		key := parameters.keyFile
		logrus.Print("Listening on port " + strconv.Itoa(parameters.port) + "...")
		logrus.Fatal(http.ListenAndServeTLS(":"+strconv.Itoa(parameters.port), cert, key, nil))
	} else {
		logrus.Print("Listening on port " + strconv.Itoa(parameters.port) + "...")
		logrus.Fatal(http.ListenAndServe(":"+strconv.Itoa(parameters.port), nil))
	}
}

// ServeMutatePods returns an admission review with pod mutations as a json patch
// in the review response
func ServeMutatePods(w http.ResponseWriter, r *http.Request) {
	logger := logrus.WithField("uri", r.RequestURI)
	logger.Debug("received mutation request")

	in, err := parseRequest(*r)
	if err != nil {
		logger.Error(err)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	adm := admission.Admitter{
		Logger:  logger,
		Request: in.Request,
	}

	out, err := adm.MutatePodReview()
	if err != nil {
		e := fmt.Sprintf("could not generate admission response: %v", err)
		logger.Error(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	jout, err := json.Marshal(out)
	if err != nil {
		e := fmt.Sprintf("could not parse admission response: %v", err)
		logger.Error(e)
		http.Error(w, e, http.StatusInternalServerError)
		return
	}

	logger.Debug("sending response")
	logger.Debugf("%s", jout)
	fmt.Fprintf(w, "%s", jout)
}

// parseRequest extracts an AdmissionReview from an http.Request if possible
func parseRequest(r http.Request) (*admissionv1.AdmissionReview, error) {
	if r.Header.Get("Content-Type") != "application/json" {
		return nil, fmt.Errorf("Content-Type: %q should be %q",
			r.Header.Get("Content-Type"), "application/json")
	}

	bodybuf := new(bytes.Buffer)
	bodybuf.ReadFrom(r.Body)
	body := bodybuf.Bytes()

	if len(body) == 0 {
		return nil, fmt.Errorf("admission request body is empty")
	}

	var a admissionv1.AdmissionReview

	if err := json.Unmarshal(body, &a); err != nil {
		return nil, fmt.Errorf("could not parse admission review request: %v", err)
	}

	if a.Request == nil {
		return nil, fmt.Errorf("admission review can't be used: Request field is nil")
	}

	return &a, nil
}
