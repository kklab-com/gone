package http

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"mime/multipart"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/kklab-com/gone-httpheadername"
	"github.com/kklab-com/gone-httpstatus"
	"github.com/kklab-com/gone/channel"
	"github.com/kklab-com/gone/http/httpsession"
	"github.com/kklab-com/goth-base62"
	"github.com/kklab-com/goth-kklogger"
	"github.com/kklab-com/goth-kkutil/buf"
	"github.com/kklab-com/goth-kkutil/validate"
	"golang.org/x/text/language"
)

const MaxMultiPartMemory = 1024 * 1024 * 256

type Request struct {
	request     *http.Request
	channel     channel.Channel
	trackID     string
	createdAt   time.Time
	remoteAddrs []string
	body        buf.ByteBuf
	session     httpsession.Session
	op          sync.Mutex
}

func WrapRequest(ch channel.Channel, req *http.Request) *Request {
	u := uuid.New()
	request := Request{
		request:   req,
		channel:   ch,
		trackID:   base62.ShiftEncoding.EncodeToString(u[:]),
		createdAt: time.Now(),
	}

	if bs, e := ioutil.ReadAll(request.request.Body); e == nil {
		request.request.Body.Close()
		request.body = buf.NewByteBuf(bs)
		request.request.Body = ioutil.NopCloser(buf.NewByteBuf(bs))
	}

	if xForwardFor := request.Header().Get(httpheadername.XForwardedFor); xForwardFor != "" {
		xffu := strings.Split(xForwardFor, ", ")
		rAddr := request.request.RemoteAddr
		for i := len(xffu) - 1; i >= 0; i-- {
			current := xffu[i]
			if validate.IsPublicIP(net.ParseIP(current)) {
				request.request.RemoteAddr = current
				break
			}
		}

		if xffu == nil {
			xffu = []string{}
		}

		request.remoteAddrs = append(xffu, rAddr)
	} else {
		request.remoteAddrs = []string{request.request.RemoteAddr}
	}

	return &request
}

func (r *Request) Request() *http.Request {
	return r.request
}

func (r *Request) Channel() channel.Channel {
	return r.channel
}

func (r *Request) CreatedAt() *time.Time {
	return &r.createdAt
}

func (r *Request) TrackID() string {
	return r.trackID
}

func (r *Request) Host() string {
	return r.request.Host
}

func (r *Request) Method() string {
	return r.request.Method
}

func (r *Request) Url() *url.URL {
	return r.request.URL
}

func (r *Request) ContentLength() int64 {
	return r.request.ContentLength
}

func (r *Request) Proto() string {
	return r.request.Proto
}

func (r *Request) UserAgent() string {
	return r.request.UserAgent()
}

func (r *Request) RequestURI() string {
	return r.request.RequestURI
}

func (r *Request) RemoteAddr() (ip net.IP, port string) {
	addr := r.request.RemoteAddr
	if strings.Count(addr, ":") > 1 {
		if strings.LastIndex(addr, "]") == -1 {
			return net.ParseIP(addr), ""
		}

		ip := net.ParseIP(addr[1 : strings.LastIndex(addr, ":")-1])
		port := addr[strings.LastIndex(addr, ":")+1:]
		return ip, port
	} else {
		if strings.LastIndex(addr, ":") == -1 {
			return net.ParseIP(addr), ""
		}

		ip := net.ParseIP(addr[:strings.LastIndex(addr, ":")])
		port := addr[strings.LastIndex(addr, ":")+1:]
		return ip, port
	}
}

func (r *Request) RemoteAddrs() []string {
	return r.remoteAddrs
}

func (r *Request) Cookies() []*http.Cookie {
	return r.request.Cookies()
}

func (r *Request) Cookie(name string) (*http.Cookie, error) {
	return r.request.Cookie(name)
}

func (r *Request) AddCookie(c *http.Cookie) {
	r.request.AddCookie(c)
}

func (r *Request) Header() http.Header {
	return r.request.Header
}

func (r *Request) Body() buf.ByteBuf {
	return r.body
}

func (r *Request) FormFile(key string) (multipart.File, *multipart.FileHeader, error) {
	return r.request.FormFile(key)
}

func (r *Request) FormValue(key string) string {
	return r.request.FormValue(key)
}

func (r *Request) RemoteIP() (ip net.IP) {
	rip, _ := r.RemoteAddr()
	return rip
}

func (r *Request) Session() httpsession.Session {
	if r.session == nil {
		r.op.Lock()
		defer r.op.Unlock()
		if r.session == nil {
			r.session = GetSession(r)
		}
	}

	return r.session
}

func (r *Request) RenewSession() {
	r.session = _NewSession(r)
}

func (r *Request) AcceptLanguage() []LanguageValue {
	var languageValues []LanguageValue
	if tags, q, err := language.ParseAcceptLanguage(r.Header().Get(httpheadername.AcceptLanguage)); err == nil {
		for i, tag := range tags {
			languageValues = append(languageValues, LanguageValue{
				Value:  tag,
				Factor: q[i],
			})
		}

		return languageValues
	}

	return languageValues
}

func (r *Request) Accept() []QualityValue {
	return _DecodeQualityValueField(r.Header().Get(httpheadername.Accept))
}

func (r *Request) AcceptEncoding() []QualityValue {
	return _DecodeQualityValueField(r.Header().Get(httpheadername.AcceptEncoding))
}

func (r *Request) TE() []QualityValue {
	return _DecodeQualityValueField(r.Header().Get(httpheadername.TE))
}

func (r *Request) AcceptCharset() []QualityValue {
	return _DecodeQualityValueField(r.Header().Get(httpheadername.AcceptCharset))
}

type QualityValue struct {
	Value  string
	Factor float32
}

type LanguageValue struct {
	Value  language.Tag
	Factor float32
}

func _DecodeQualityValueField(text string) []QualityValue {
	var qualities []QualityValue
	if text == "" || text == "*" {
		return qualities
	}

	for _, entity := range strings.Split(strings.TrimSpace(text), ",") {
		qv := QualityValue{}
		split := strings.Split(entity, ";")
		if len(split) == 2 {
			factor := strings.Split(split[1], "=")
			if len(factor) == 2 {
				if cf, err := strconv.ParseFloat(factor[1], 32); err == nil {
					qv.Factor = float32(cf)
				}
			}
		}

		qv.Value = split[0]
		qualities = append(qualities, qv)
	}

	return qualities
}

func (r *Request) PreferLang() *language.Tag {
	lang := r.AcceptLanguage()
	if len(lang) > 0 {
		return &lang[0].Value
	} else {
		return nil
	}
}

func (r *Request) Referer() string {
	return r.Header().Get(httpheadername.Referer)
}

func (r *Request) Origin() string {
	return r.Header().Get(httpheadername.Origin)
}

type Response struct {
	response   *http.Response
	request    *Request
	statusCode int
	header     http.Header
	cookies    map[string][]http.Cookie
	body       buf.ByteBuf
}

func WrapResponse(ch channel.NetChannel, response *http.Response) *Response {
	resp := &Response{
		request: &Request{
			request: response.Request,
			channel: ch,
		},
		statusCode: response.StatusCode,
		header:     response.Header,
	}

	return resp
}

func (r *Response) Request() *Request {
	return r.request
}

func NewResponse(request *Request) *Response {
	response := &Response{}
	response.request = request
	response.statusCode = httpstatus.OK
	response.header = map[string][]string{}
	response.cookies = map[string][]http.Cookie{}
	response.body = buf.EmptyByteBuf()
	return response
}

func EmptyResponse() *Response {
	response := &Response{}
	response.header = map[string][]string{}
	response.cookies = map[string][]http.Cookie{}
	response.body = buf.EmptyByteBuf()
	return response
}

func (r *Response) ResponseError(er ErrorResponse) {
	er = er.Clone()
	er.ErrorData()["cid"] = r.request.Channel().ID()
	er.ErrorData()["tid"] = r.request.TrackID()
	r.SetStatusCode(er.ErrorStatusCode()).
		JsonResponse(buf.NewByteBufString(er.Error()))
}

func (r *Response) Redirect(redirectUrl string) {
	r.SetStatusCode(httpstatus.Found).
		SetHeader(httpheadername.Location, redirectUrl)
}

func (r *Response) StatusCode() int {
	return r.statusCode
}

func (r *Response) SetStatusCode(statusCode int) *Response {
	r.statusCode = statusCode
	return r
}

func (r *Response) AddHeader(name string, value string) *Response {
	r.header.Add(name, value)
	return r
}

func (r *Response) SetHeader(name string, value string) *Response {
	r.header.Set(name, value)
	return r
}

func (r *Response) DelHeader(name string) *Response {
	r.header.Del(name)
	return r
}

func (r *Response) Header() http.Header {
	return r.header
}

func (r *Response) GetHeader(name string) string {
	return r.header.Get(name)
}

func (r *Response) Cookie(name string) *http.Cookie {
	if cookie, f := r.cookies[name]; f {
		return &cookie[0]
	}

	return nil
}

func (r *Response) SetCookie(cookie *http.Cookie) *Response {
	r.cookies[cookie.Name] = []http.Cookie{*cookie}
	return r
}

func (r *Response) Cookies() map[string][]http.Cookie {
	return r.cookies
}

func (r *Response) Body() []byte {
	return r.body.Bytes()
}

func (r *Response) SetBody(buf buf.ByteBuf) {
	r.body = buf
}

func (r *Response) SetContentType(ct string) {
	r.SetHeader(httpheadername.ContentType, ct)
}

func (r *Response) TextResponse(buf buf.ByteBuf) {
	r.
		SetHeader(httpheadername.ContentType, "text/plain").
		SetBody(buf)
}

func (r *Response) JsonResponse(obj interface{}) {
	r.SetHeader(httpheadername.ContentType, "application/json")

	switch body := obj.(type) {
	case *bytes.Buffer:
		r.SetBody(buf.NewByteBuf(body.Bytes()))
	case buf.ByteBuf:
		r.SetBody(body)
	case []byte:
		r.SetBody(buf.NewByteBuf(body))
	case string:
		obj = struct{ Data string }{Data: body}
	default:
		if body, e := json.Marshal(obj); e == nil {
			r.SetBody(buf.NewByteBuf(body))
			return
		} else {
			kklogger.ErrorJ("gone:Response.JsonResponse#JsonMarshal", e.Error())
		}
	}
}

type ResponseWriter interface {
	http.ResponseWriter
}

type Pack struct {
	Request   *Request               `json:"request"`
	Response  *Response              `json:"response"`
	RouteNode RouteNode              `json:"route_node"`
	Params    map[string]interface{} `json:"params"`
	Writer    ResponseWriter         `json:"writer"`
}

func _UnPack(obj interface{}) *Pack {
	if pkg, true := obj.(*Pack); true {
		return pkg
	}

	return nil
}
