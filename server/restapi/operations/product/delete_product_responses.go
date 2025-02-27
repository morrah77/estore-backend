// Code generated by go-swagger; DO NOT EDIT.

package product

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"

	"github.com/go-openapi/runtime"

	"estore-backend/server/models"
)

// DeleteProductNoContentCode is the HTTP code returned for type DeleteProductNoContent
const DeleteProductNoContentCode int = 204

/*
DeleteProductNoContent Deleted

swagger:response deleteProductNoContent
*/
type DeleteProductNoContent struct {
}

// NewDeleteProductNoContent creates DeleteProductNoContent with default headers values
func NewDeleteProductNoContent() *DeleteProductNoContent {

	return &DeleteProductNoContent{}
}

// WriteResponse to the client
func (o *DeleteProductNoContent) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.Header().Del(runtime.HeaderContentType) //Remove Content-Type on empty responses

	rw.WriteHeader(204)
}

/*
DeleteProductDefault Error

swagger:response deleteProductDefault
*/
type DeleteProductDefault struct {
	_statusCode int

	/*
	  In: Body
	*/
	Payload *models.Error `json:"body,omitempty"`
}

// NewDeleteProductDefault creates DeleteProductDefault with default headers values
func NewDeleteProductDefault(code int) *DeleteProductDefault {
	if code <= 0 {
		code = 500
	}

	return &DeleteProductDefault{
		_statusCode: code,
	}
}

// WithStatusCode adds the status to the delete product default response
func (o *DeleteProductDefault) WithStatusCode(code int) *DeleteProductDefault {
	o._statusCode = code
	return o
}

// SetStatusCode sets the status to the delete product default response
func (o *DeleteProductDefault) SetStatusCode(code int) {
	o._statusCode = code
}

// WithPayload adds the payload to the delete product default response
func (o *DeleteProductDefault) WithPayload(payload *models.Error) *DeleteProductDefault {
	o.Payload = payload
	return o
}

// SetPayload sets the payload to the delete product default response
func (o *DeleteProductDefault) SetPayload(payload *models.Error) {
	o.Payload = payload
}

// WriteResponse to the client
func (o *DeleteProductDefault) WriteResponse(rw http.ResponseWriter, producer runtime.Producer) {

	rw.WriteHeader(o._statusCode)
	if o.Payload != nil {
		payload := o.Payload
		if err := producer.Produce(rw, payload); err != nil {
			panic(err) // let the recovery middleware deal with this
		}
	}
}
