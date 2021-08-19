package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	admissionv1 "k8s.io/api/admission/v1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

// toAdmissionResponse is a helper function to create an AdmissionResponse
// with an embedded error
func toAdmissionResponse(err error) *admissionv1.AdmissionResponse {
	return &admissionv1.AdmissionResponse{
		Result: &metav1.Status{
			Message: err.Error(),
		},
	}
}

type admitFunc func(admissionv1.AdmissionReview) *admissionv1.AdmissionResponse

// serve handles the http portion of a request prior to handing to an admit
// function
func serve(w http.ResponseWriter, r *http.Request, admit admitFunc) {
	var body []byte
	if r.Body != nil {
		if data, err := ioutil.ReadAll(r.Body); err == nil {
			body = data
		}
	}

	// verify the content type is accurate
	contentType := r.Header.Get("Content-Type")
	if contentType != "application/json" {
		klog.Errorf("contentType=%s, expect application/json", contentType)
		return
	}

	klog.Infoln(fmt.Sprintf("handling request: %s", body))

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := admissionv1.AdmissionReview{}

	// The AdmissionReview that will be returned
	responseAdmissionReview := admissionv1.AdmissionReview{}

	if err := json.Unmarshal(body, &requestedAdmissionReview); err != nil {
		klog.Error(err)
		responseAdmissionReview.Response = toAdmissionResponse(err)
	} else {
		// pass to admitFunc
		responseAdmissionReview.Response = admit(requestedAdmissionReview)
	}

	// Return the same UID
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID
	// Return the same ApiVersion
	responseAdmissionReview.APIVersion = requestedAdmissionReview.APIVersion
	// Return same kind
	responseAdmissionReview.Kind = requestedAdmissionReview.Kind

	klog.V(2).Info(fmt.Sprintf("sending response: %v", responseAdmissionReview.Response))

	respBytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		klog.Error(err)
	}
	if _, err := w.Write(respBytes); err != nil {
		klog.Error(err)
	}
}

func mutate(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.Infoln("entering the mutate func")

	// Preparing our review Response
	reviewResponse := admissionv1.AdmissionResponse{}
	// As we are mutating here and are just adding an annotation, we will allow this operation
	reviewResponse.Allowed = true

	// Getting the current date for the timestamp annotation
	date := time.Now()

	// Creating our Patch Operation (This is just for demonstration purposes) - see also: https://tools.ietf.org/html/rfc6902
	addTimeStampAnnotation := `[{ "op": "add", "path": "/metadata/annotations/deployment_timestamp", "value": "` + date.String() + `" }]`

	// Adding the Timestamp to the Object
	reviewResponse.Patch = []byte(addTimeStampAnnotation)

	pt := admissionv1.PatchTypeJSONPatch
	reviewResponse.PatchType = &pt
	return &reviewResponse
}

func validate(ar admissionv1.AdmissionReview) *admissionv1.AdmissionResponse {
	klog.Info("entering the validate func")

	// Preparing our review response
	reviewResponse := admissionv1.AdmissionResponse{}
	reviewResponse.Allowed = true

	raw := ar.Request.Object.Raw

	var deploy appsv1.Deployment
	// Unmarshalling the data in the deployment struct
	if err := json.Unmarshal(raw, &deploy); err != nil {
		klog.Errorf("Could not unmarshal raw object: %v", err)
		return toAdmissionResponse(err)
	}
	// Check if RunAsNonRoot is set
	// It may be set to false in edge cases, but it needs to be set
	if reflect.ValueOf(deploy.Spec.Template.Spec.SecurityContext.RunAsNonRoot).IsNil() {
		err := errors.New("need to set RunAsNonRoot")
		return toAdmissionResponse(err)
	}

	return &reviewResponse
}
