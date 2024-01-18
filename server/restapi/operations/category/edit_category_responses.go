// Code generated by go-swagger; DO NOT EDIT.

package category

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"estore-backend/server/models"
)

// EditCategoryOKCode is the HTTP code returned for type EditCategoryOK
const EditCategoryOKCode int = 200

/*
EditCategoryOK OK

swagger:response editCategoryOK
*/
type EditCategoryOK struct {

	/*
	  In: Body
	*/
	Payload *models.Category `json:"body,omitempty"`
}

// NewEditCategoryOK creates EditCategoryOK with default headers values
func NewEditCategoryOK() *EditCategoryOK {

	return &EditCategoryOK{}
}

// WithPayload adds the payload to the edit category o k response
func (o *EditCategoryOK) WithPayload(payload *models.Category) *EditCategoryOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the edit category o k response
func (o *EditCategoryOK) SetPayload(payload *models.Category) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EditCategoryOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*
EditCategoryDefault Error

swagger:response editCategoryDefault
*/
type EditCategoryDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewEditCategoryDefault creates EditCategoryDefault with default headers values
func NewEditCategoryDefault(code int) *EditCategoryDefault {
	if code <= 0 {
		code = 500
	}

	return &EditCategoryDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the edit category default response
func (o *EditCategoryDefault) WithStatusCode(code int) *EditCategoryDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the edit category default response
func (o *EditCategoryDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the edit category default response
func (o *EditCategoryDefault) WithPayload(payload *models.Error) *EditCategoryDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the edit category default response
func (o *EditCategoryDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EditCategoryDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
