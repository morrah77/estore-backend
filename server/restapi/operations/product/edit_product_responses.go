// Code generated by go-swagger; DO NOT EDIT.

package product

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"estore-backend/server/models"
)

// EditProductOKCode is the HTTP code returned for type EditProductOK
const EditProductOKCode int = 200

/*
EditProductOK OK

swagger:response editProductOK
*/
type EditProductOK struct {

	/*
	  In: Body
	*/
	Payload *models.Product `json:"body,omitempty"`
}

// NewEditProductOK creates EditProductOK with default headers values
func NewEditProductOK() *EditProductOK {

	return &EditProductOK{}
}

// WithPayload adds the payload to the edit product o k response
func (o *EditProductOK) WithPayload(payload *models.Product) *EditProductOK {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the edit product o k response
func (o *EditProductOK) SetPayload(payload *models.Product) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EditProductOK) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(200)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}

/*
EditProductDefault Error

swagger:response editProductDefault
*/
type EditProductDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewEditProductDefault creates EditProductDefault with default headers values
func NewEditProductDefault(code int) *EditProductDefault {
	if code <= 0 {
		code = 500
	}

	return &EditProductDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the edit product default response
func (o *EditProductDefault) WithStatusCode(code int) *EditProductDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the edit product default response
func (o *EditProductDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the edit product default response
func (o *EditProductDefault) WithPayload(payload *models.Error) *EditProductDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the edit product default response
func (o *EditProductDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *EditProductDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
