package main

import (
	"fmt"
	"github.com/shell909090/goproxy/msocks"
	"github.com/shell909090/goproxy/sutils"
	"net/http"
	"net/http/pprof"
	"text/template"
)

const (
	str_sess = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
  <head>
    <title>session list</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="author" content="Shell.Xu">
  </head>
  <body>
    <table>
      <tr>
        <td><a href="cutoff">cutoff</a></td>
	<td>
	  <form action="lookup">
	    <input type="text" name="host" value="www.google.com">
	    <input type="submit" value="lookup">
	  </form>
	</td>
      </tr>
    </table>
    <table>
      <tr>
	<th>Sess</th><th>Id</th><th>State</th>
        <th>Recv-Q</th><th>Send-Q</th><th width="50%">Address</th>
      </tr>
      {{if .GetSize}}
      {{range $sess := .GetSess}}
      <tr>
	<td>{{$sess.GetId}}</td>
	<td></td>
	<td>{{$sess.GetSize}}/{{printf "%0.2fs" $sess.GetLastPing.Seconds}}</td>
	<td>{{$sess.GetReadSpeed}}</td>
	<td>{{$sess.GetWriteSpeed}}</td>
	<td>{{$sess.RemoteAddr}}</td>
      </tr>
      {{range $conn := $sess.GetPorts}}
      <tr>
	{{with $conn}}
	<td></td>
	<td>{{$conn.GetStreamId}}</td>
	<td>{{$conn.GetStatus}}</td>
	<td>{{$conn.GetReadBufSize}}</td>
	<td>{{$conn.GetWriteBufSize}}</td>
	<td>{{$conn.Address}}</td>
	{{else}}
	<td></td>
	<td>half closed</td>
	{{end}}
      </tr>
      {{end}}
      {{end}}
      {{else}}
      <tr><td>no session</td></tr>
      {{end}}
    </table>
  </body>
</html>`
	str_addrs = `
<!DOCTYPE HTML PUBLIC "-//W3C//DTD HTML 4.01//EN" "http://www.w3.org/TR/html4/strict.dtd">
<html>
  <head>
    <title>address list</title>
    <meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
    <meta name="author" content="Shell.Xu">
  </head>
  <body>
    <table>
      {{range $addr := .}}
	<tr>
	  <td>{{$addr}}</td>
	</tr>
      {{end}}
    </table>
  </body>
</html>`
)

var (
	tmpl_sess *template.Template
	tmpl_addr *template.Template
)

func init() {
	var err error
	tmpl_sess, err = template.New("session").Parse(str_sess)
	if err != nil {
		panic(err)
	}
	tmpl_addr, err = template.New("address").Parse(str_addrs)
	if err != nil {
		panic(err)
	}
}

type MsocksManager struct {
	sp       *msocks.SessionPool
	lookuper sutils.Lookuper
}

func NewMsocksManager(sp *msocks.SessionPool, lookuper sutils.Lookuper) (mm *MsocksManager) {
	mm = &MsocksManager{
		sp:       sp,
		lookuper: lookuper,
	}
	return
}

func (mm *MsocksManager) Register(mux *http.ServeMux) {
	mux.HandleFunc("/", mm.HandlerMain)
	mux.HandleFunc("/lookup", mm.HandlerLookup)
	mux.HandleFunc("/cutoff", mm.HandlerCutoff)
	mux.HandleFunc("/debug/pprof/", pprof.Index)
	mux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	mux.HandleFunc("/debug/pprof/profile", pprof.Profile)
	mux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
}

func (mm *MsocksManager) HandlerMain(w http.ResponseWriter, req *http.Request) {
	err := tmpl_sess.Execute(w, mm.sp)
	if err != nil {
		log.Error("%s", err)
	}
	return
}

func (mm *MsocksManager) HandlerLookup(w http.ResponseWriter, req *http.Request) {
	q := req.URL.Query()
	hosts, ok := q["host"]
	if !ok {
		w.WriteHeader(400)
		w.Write([]byte("no domain"))
		return
	}

	addrs, err := mm.lookuper.LookupIP(hosts[0])
	if err != nil {
		w.WriteHeader(500)
		fmt.Fprintf(w, "error %s", err)
		return
	}

	err = tmpl_addr.Execute(w, addrs)
	if err != nil {
		log.Error("%s", err)
	}
	return
}

func (mm *MsocksManager) HandlerCutoff(w http.ResponseWriter, req *http.Request) {
	mm.sp.CutAll()
	return
}
