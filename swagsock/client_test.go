package swagsock

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"
	"github.com/go-openapi/strfmt"

	"github.com/stretchr/testify/assert"

	// use these generated packages for testing
	"github.com/elakito/swagsock/examples/greeter/models"
)

func TestClient(t *testing.T) {
	conf := NewConfig()
	conf.Heartbeat = 5
	ph := CreateProtocolHandler(conf)
	defer ph.Destroy()
	hh := &testEchoHandler{mediator: conf.ResponseMediator}

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ph.Serve(hh, w, r)
	}))
	defer ts.Close()

	transport := NewTransport("ws" + ts.URL[4:])
	assert.NotNil(t, transport)
	defer transport.Close()

	client := New(transport, strfmt.Default)

	// synchronous variants
	pingOK, err := client.Ping(nil)
	assert.Nil(t, err)
	assert.NotNil(t, pingOK.Payload)

	echoOK, err := client.Echo(NewEchoParams().WithBody("hola"))
	assert.Nil(t, err)
	assert.NotNil(t, echoOK.Payload)

	// asynchronous variants
	received := make(chan struct{})
	var resid string
	var resval interface{}
	rid, err := client.PingAsync(nil, func(reqid string, r *PingOK, e error) {
		resid = reqid
		resval = r.Payload
		received <- struct{}{}
	}, SubmitAsyncOptionNone)
	assert.Nil(t, err)

	select {
	case <-received:
		assert.Equal(t, rid, resid)
		if success, ok := resval.(*models.Pong); ok {
			assert.NotNil(t, success)
		} else {
			assert.Fail(t, "unexpected response: %v", resval)
		}
	case <-time.After(2 * time.Second):
		assert.Fail(t, "response not received")
	}

	rid, err = client.EchoAsync(NewEchoParams().WithBody("hola"), func(reqid string, r *EchoOK, e error) {
		resid = reqid
		resval = r.Payload
		received <- struct{}{}
	}, SubmitAsyncOptionNone)
	assert.Nil(t, err)

	select {
	case <-received:
		assert.Equal(t, rid, resid)
		if success, ok := resval.(*models.EchoReply); ok {
			assert.NotNil(t, success)
		} else {
			assert.Fail(t, "unexpected response: %v", resval)
		}
	case <-time.After(2 * time.Second):
		assert.Fail(t, "response not received")
	}

	// subscribe
	var resvals []*models.GreetingReply
	subid, err := client.SubscribeAsync(NewSubscribeParams().WithName("dog"), func(reqid string, r *SubscribeOK, e error) {
		resid = reqid
		resvals = append(resvals, r.Payload)
		received <- struct{}{}
	}, SubmitAsyncOptionSubscribe)
	assert.Nil(t, err)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		assert.Fail(t, "response not received")
	}
	assert.Equal(t, subid, resid)
	assert.Equal(t, 1, len(resvals))
	assert.Equal(t, 1, len(conf.ResponseMediator.Subscribed()))

	// echo after subscribed
	_, err = client.EchoAsync(NewEchoParams().WithBody("hola"), func(reqid string, r *EchoOK, e error) {
	}, SubmitAsyncOptionNone)
	assert.Nil(t, err)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		assert.Fail(t, "response not received")
	}
	assert.Equal(t, 2, len(resvals))
	assert.Equal(t, "system", resvals[1].From)
	assert.Equal(t, "hola", resvals[1].Text)

	// unsubscribe
	_, err = client.UnsubscribeAsync(NewUnsubscribeParams().WithSid(subid), func(reqid string, r *UnsubscribeOK, e error) {
		resid = reqid
		received <- struct{}{}
	}, SubmitAsyncOptionUnsubscribe(subid))
	assert.Nil(t, err)

	select {
	case <-received:
	case <-time.After(2 * time.Second):
		assert.Fail(t, "response not received")
	}
	assert.Equal(t, 0, len(conf.ResponseMediator.Subscribed()))
}

// DO NOT EDIT BELOW
// the code are copied from github.com/elakito/swagsock/examples/greeter-client/client and adjusted to avoid creating a cyclic dependency
func New(transport runtime.ClientTransport, formats strfmt.Registry) *Client {
	return &Client{transport: transport, formats: formats}
}

/*
Client for operations API
*/
type Client struct {
	transport runtime.ClientTransport
	formats   strfmt.Registry
}

/*
Echo echos back the message

Echo back the message
*/
func (a *Client) Echo(params *EchoParams) (*EchoOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewEchoParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "echo",
		Method:             "POST",
		PathPattern:        "/v1/echo",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"text/plain"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &EchoReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*EchoOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for echo: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
EchoAsync echos back the message
Limitation: currently only one successful return type is supported
Echo back the message
*/
func (a *Client) EchoAsync(params *EchoParams, cb func(string, *EchoOK, error), sam SubmitAsyncOption) (string, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewEchoParams()
	}
	requid, err := a.transport.(ClientTransport).SubmitAsync(&runtime.ClientOperation{
		ID:                 "echo",
		Method:             "POST",
		PathPattern:        "/v1/echo",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{"text/plain"},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &EchoReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}, func(k string, v interface{}) {
		switch value := v.(type) {
		case *EchoOK:
			cb(k, value, nil)
		case error:
			cb(k, nil, value)
		}
	}, sam)
	if err != nil {
		return "", err
	}
	return requid, nil
}

func (a *Client) Ping(params *PingParams) (*PingOK, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPingParams()
	}

	result, err := a.transport.Submit(&runtime.ClientOperation{
		ID:                 "ping",
		Method:             "GET",
		PathPattern:        "/v1/ping",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{""},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &PingReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	})
	if err != nil {
		return nil, err
	}
	success, ok := result.(*PingOK)
	if ok {
		return success, nil
	}
	// unexpected success response
	// safeguard: normally, absent a default response, unknown success responses return an error above: so this is a codegen issue
	msg := fmt.Sprintf("unexpected success response for ping: API contract not enforced by server. Client expected to get an error, but got: %T", result)
	panic(msg)
}

/*
PingAsync pings the greeter service
Limitation: currently only one successful return type is supported
Ping the greeter service
*/
func (a *Client) PingAsync(params *PingParams, cb func(string, *PingOK, error), sam SubmitAsyncOption) (string, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewPingParams()
	}
	requid, err := a.transport.(ClientTransport).SubmitAsync(&runtime.ClientOperation{
		ID:                 "ping",
		Method:             "GET",
		PathPattern:        "/v1/ping",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{""},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &PingReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}, func(k string, v interface{}) {
		switch value := v.(type) {
		case *PingOK:
			cb(k, value, nil)
		case error:
			cb(k, nil, value)
		}
	}, sam)
	if err != nil {
		return "", err
	}
	return requid, nil
}

/*
SubscribeAsync subscribes to the greeting events
Limitation: currently only one successful return type is supported
Subscribe to the greeteing events
*/
func (a *Client) SubscribeAsync(params *SubscribeParams, cb func(string, *SubscribeOK, error), sam SubmitAsyncOption) (string, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewSubscribeParams()
	}
	requid, err := a.transport.(ClientTransport).SubmitAsync(&runtime.ClientOperation{
		ID:                 "subscribe",
		Method:             "GET",
		PathPattern:        "/v1/subscribe/{name}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{""},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &SubscribeReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}, func(k string, v interface{}) {
		switch value := v.(type) {
		case *SubscribeOK:
			cb(k, value, nil)
		case error:
			cb(k, nil, value)
		}
	}, sam)
	if err != nil {
		return "", err
	}
	return requid, nil
}

/*
UnsubscribeAsync unsubscribes from the greeting events
Limitation: currently only one successful return type is supported
Unsubscribe from the greeting events
*/
func (a *Client) UnsubscribeAsync(params *UnsubscribeParams, cb func(string, *UnsubscribeOK, error), sam SubmitAsyncOption) (string, error) {
	// TODO: Validate the params before sending
	if params == nil {
		params = NewUnsubscribeParams()
	}
	requid, err := a.transport.(ClientTransport).SubmitAsync(&runtime.ClientOperation{
		ID:                 "unsubscribe",
		Method:             "DELETE",
		PathPattern:        "/v1/unsubscribe/{sid}",
		ProducesMediaTypes: []string{"application/json"},
		ConsumesMediaTypes: []string{""},
		Schemes:            []string{"http", "https"},
		Params:             params,
		Reader:             &UnsubscribeReader{formats: a.formats},
		Context:            params.Context,
		Client:             params.HTTPClient,
	}, func(k string, v interface{}) {
		switch value := v.(type) {
		case *UnsubscribeOK:
			cb(k, value, nil)
		case error:
			cb(k, nil, value)
		}
	}, sam)
	if err != nil {
		return "", err
	}
	return requid, nil
}

////////////////////////////////////////////////// operations/echo_parameters.go
// NewEchoParams creates a new EchoParams object
// with the default values initialized.
func NewEchoParams() *EchoParams {
	var ()
	return &EchoParams{

		timeout: cr.DefaultTimeout,
	}
}

/*EchoParams contains all the parameters to send to the API endpoint
for the echo operation typically these are written to a http.Request
*/
type EchoParams struct {

	/*Body
	  A message

	*/
	Body string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the echo params
func (o *EchoParams) WithTimeout(timeout time.Duration) *EchoParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the echo params
func (o *EchoParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the echo params
func (o *EchoParams) WithContext(ctx context.Context) *EchoParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the echo params
func (o *EchoParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the echo params
func (o *EchoParams) WithHTTPClient(client *http.Client) *EchoParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the echo params
func (o *EchoParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithBody adds the body to the echo params
func (o *EchoParams) WithBody(body string) *EchoParams {
	o.SetBody(body)
	return o
}

// SetBody adds the body to the echo params
func (o *EchoParams) SetBody(body string) {
	o.Body = body
}

// WriteToRequest writes these params to a swagger request
func (o *EchoParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if err := r.SetBodyParam(o.Body); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

////////////////////////////////////////////////// operations/ping_parameters.go
// NewPingParams creates a new PingParams object
// with the default values initialized.
func NewPingParams() *PingParams {

	return &PingParams{

		timeout: cr.DefaultTimeout,
	}
}

/*PingParams contains all the parameters to send to the API endpoint
for the ping operation typically these are written to a http.Request
*/
type PingParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the ping params
func (o *PingParams) WithTimeout(timeout time.Duration) *PingParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the ping params
func (o *PingParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the ping params
func (o *PingParams) WithContext(ctx context.Context) *PingParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the ping params
func (o *PingParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the ping params
func (o *PingParams) WithHTTPClient(client *http.Client) *PingParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the ping params
func (o *PingParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *PingParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

////////////////////////////////////////////////// operations/subscribe_parameters.go
// NewSubscribeParams creates a new SubscribeParams object
// with the default values initialized.
func NewSubscribeParams() *SubscribeParams {
	var ()
	return &SubscribeParams{

		timeout: cr.DefaultTimeout,
	}
}

/*SubscribeParams contains all the parameters to send to the API endpoint
for the subscribe operation typically these are written to a http.Request
*/
type SubscribeParams struct {

	/*Name
	  greeted name

	*/
	Name string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the subscribe params
func (o *SubscribeParams) WithTimeout(timeout time.Duration) *SubscribeParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the subscribe params
func (o *SubscribeParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the subscribe params
func (o *SubscribeParams) WithContext(ctx context.Context) *SubscribeParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the subscribe params
func (o *SubscribeParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the subscribe params
func (o *SubscribeParams) WithHTTPClient(client *http.Client) *SubscribeParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the subscribe params
func (o *SubscribeParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithName adds the name to the subscribe params
func (o *SubscribeParams) WithName(name string) *SubscribeParams {
	o.SetName(name)
	return o
}

// SetName adds the name to the subscribe params
func (o *SubscribeParams) SetName(name string) {
	o.Name = name
}

// WriteToRequest writes these params to a swagger request
func (o *SubscribeParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param name
	if err := r.SetPathParam("name", o.Name); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

////////////////////////////////////////////////// operations/unsubscribe_parameters.go
// NewUnsubscribeParams creates a new UnsubscribeParams object
// with the default values initialized.
func NewUnsubscribeParams() *UnsubscribeParams {
	var ()
	return &UnsubscribeParams{

		timeout: cr.DefaultTimeout,
	}
}

/*UnsubscribeParams contains all the parameters to send to the API endpoint
for the unsubscribe operation typically these are written to a http.Request
*/
type UnsubscribeParams struct {

	/*Sid
	  subscription id which corresponds to the id used to start a subscription

	*/
	Sid string

	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the unsubscribe params
func (o *UnsubscribeParams) WithTimeout(timeout time.Duration) *UnsubscribeParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the unsubscribe params
func (o *UnsubscribeParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the unsubscribe params
func (o *UnsubscribeParams) WithContext(ctx context.Context) *UnsubscribeParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the unsubscribe params
func (o *UnsubscribeParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the unsubscribe params
func (o *UnsubscribeParams) WithHTTPClient(client *http.Client) *UnsubscribeParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the unsubscribe params
func (o *UnsubscribeParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WithSid adds the sid to the unsubscribe params
func (o *UnsubscribeParams) WithSid(sid string) *UnsubscribeParams {
	o.SetSid(sid)
	return o
}

// SetSid adds the sid to the unsubscribe params
func (o *UnsubscribeParams) SetSid(sid string) {
	o.Sid = sid
}

// WriteToRequest writes these params to a swagger request
func (o *UnsubscribeParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	// path param sid
	if err := r.SetPathParam("sid", o.Sid); err != nil {
		return err
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

////////////////////////////////////////////////// operations/echo_responses.go
// EchoReader is a Reader for the Echo structure.
type EchoReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *EchoReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewEchoOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewEchoOK creates a EchoOK with default headers values
func NewEchoOK() *EchoOK {
	return &EchoOK{}
}

/*EchoOK handles this case with default header values.

Returns an echo message
*/
type EchoOK struct {
	Payload *models.EchoReply
}

func (o *EchoOK) Error() string {
	return fmt.Sprintf("[POST /v1/echo][%d] echoOK  %+v", 200, o.Payload)
}

func (o *EchoOK) GetPayload() *models.EchoReply {
	return o.Payload
}

func (o *EchoOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.EchoReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

////////////////////////////////////////////////// operations/ping_responses.go
// PingReader is a Reader for the Ping structure.
type PingReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *PingReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewPingOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewPingOK creates a PingOK with default headers values
func NewPingOK() *PingOK {
	return &PingOK{}
}

/*PingOK handles this case with default header values.

Returns a pong
*/
type PingOK struct {
	Payload *models.Pong
}

func (o *PingOK) Error() string {
	return fmt.Sprintf("[GET /v1/ping][%d] pingOK  %+v", 200, o.Payload)
}

func (o *PingOK) GetPayload() *models.Pong {
	return o.Payload
}

func (o *PingOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.Pong)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

////////////////////////////////////////////////// operations/subscribe_responses.go
// SubscribeReader is a Reader for the Subscribe structure.
type SubscribeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *SubscribeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewSubscribeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewSubscribeOK creates a SubscribeOK with default headers values
func NewSubscribeOK() *SubscribeOK {
	return &SubscribeOK{}
}

/*SubscribeOK handles this case with default header values.

Returns an empty or a sequence of greeting events
*/
type SubscribeOK struct {
	Payload *models.GreetingReply
}

func (o *SubscribeOK) Error() string {
	return fmt.Sprintf("[GET /v1/subscribe/{name}][%d] subscribeOK  %+v", 200, o.Payload)
}

func (o *SubscribeOK) GetPayload() *models.GreetingReply {
	return o.Payload
}

func (o *SubscribeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GreetingReply)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

////////////////////////////////////////////////// operations/unsubscribe_responses.go
// UnsubscribeReader is a Reader for the Unsubscribe structure.
type UnsubscribeReader struct {
	formats strfmt.Registry
}

// ReadResponse reads a server response into the received o.
func (o *UnsubscribeReader) ReadResponse(response runtime.ClientResponse, consumer runtime.Consumer) (interface{}, error) {
	switch response.Code() {
	case 200:
		result := NewUnsubscribeOK()
		if err := result.readResponse(response, consumer, o.formats); err != nil {
			return nil, err
		}
		return result, nil

	default:
		return nil, runtime.NewAPIError("unknown error", response, response.Code())
	}
}

// NewUnsubscribeOK creates a UnsubscribeOK with default headers values
func NewUnsubscribeOK() *UnsubscribeOK {
	return &UnsubscribeOK{}
}

/*UnsubscribeOK handles this case with default header values.

Returns the subscription summary
*/
type UnsubscribeOK struct {
	Payload *models.GreetingSummary
}

func (o *UnsubscribeOK) Error() string {
	return fmt.Sprintf("[DELETE /v1/unsubscribe/{sid}][%d] unsubscribeOK  %+v", 200, o.Payload)
}

func (o *UnsubscribeOK) GetPayload() *models.GreetingSummary {
	return o.Payload
}

func (o *UnsubscribeOK) readResponse(response runtime.ClientResponse, consumer runtime.Consumer, formats strfmt.Registry) error {

	o.Payload = new(models.GreetingSummary)

	// response payload
	if err := consumer.Consume(response.Body(), o.Payload); err != nil && err != io.EOF {
		return err
	}

	return nil
}

//
// a copy go-openapi/runtime/runtime/client/request_test to test the package private code copied from
// go-openapi/runtime/runtime/client/request
//
//////////////////// DO NOT EDIT THIS PART
func TestBuildRequest_SetHeaders(t *testing.T) {
	r, _ := newRequest("GET", "/flats/{id}/", nil)
	// single value
	_ = r.SetHeaderParam("X-Rate-Limit", "500")
	assert.Equal(t, "500", r.header.Get("X-Rate-Limit"))
	_ = r.SetHeaderParam("X-Rate-Limit", "400")
	assert.Equal(t, "400", r.header.Get("X-Rate-Limit"))

	// multi value
	_ = r.SetHeaderParam("X-Accepts", "json", "xml", "yaml")
	assert.EqualValues(t, []string{"json", "xml", "yaml"}, r.header["X-Accepts"])
}

func TestBuildRequest_SetPath(t *testing.T) {
	r, _ := newRequest("GET", "/flats/{id}/?hello=world", nil)

	_ = r.SetPathParam("id", "1345")
	assert.Equal(t, "1345", r.pathParams["id"])
}

func TestBuildRequest_SetQuery(t *testing.T) {
	r, _ := newRequest("GET", "/flats/{id}/", nil)

	// single value
	_ = r.SetQueryParam("hello", "there")
	assert.Equal(t, "there", r.query.Get("hello"))

	// multi value
	_ = r.SetQueryParam("goodbye", "cruel", "world")
	assert.Equal(t, []string{"cruel", "world"}, r.query["goodbye"])
}

func TestBuildRequest_SetForm(t *testing.T) {
	// non-multipart
	r, _ := newRequest("POST", "/flats", nil)
	_ = r.SetFormParam("hello", "world")
	assert.Equal(t, "world", r.formFields.Get("hello"))
	_ = r.SetFormParam("goodbye", "cruel", "world")
	assert.Equal(t, []string{"cruel", "world"}, r.formFields["goodbye"])
}

func TestBuildRequest_SetFile(t *testing.T) {
	// needs to convert form to multipart
	r, _ := newRequest("POST", "/flats/{id}/image", nil)
	// error if it isn't there
	err := r.SetFileParam("not there", os.NewFile(0, "./i-dont-exist"))
	assert.Error(t, err)
	// error if it isn't a file
	err = r.SetFileParam("directory", os.NewFile(0, "../client"))
	assert.Error(t, err)
	// success adds it to the map
	err = r.SetFileParam("file", mustGetFile("./api.go"))
	if assert.NoError(t, err) {
		fl, ok := r.fileFields["file"]
		if assert.True(t, ok) {
			assert.Equal(t, "api.go", filepath.Base(fl[0].Name()))
		}
	}
	// success adds a file param with multiple files
	err = r.SetFileParam("otherfiles", mustGetFile("./api.go"), mustGetFile("./client.go"))
	if assert.NoError(t, err) {
		fl, ok := r.fileFields["otherfiles"]
		if assert.True(t, ok) {
			assert.Equal(t, "api.go", filepath.Base(fl[0].Name()))
			assert.Equal(t, "client.go", filepath.Base(fl[1].Name()))
		}
	}
}
func mustGetFile(path string) *os.File {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	return f
}

//////////////////// DO NOT EDIT END
