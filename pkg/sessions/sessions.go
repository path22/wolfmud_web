package sessions

import (
	"bufio"
	"encoding/json"
	"fmt"
	webconfig "github.com/path22/wolfmud_web/pkg/config"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"sync"
	"text/template"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
)

const (
	sessionCookie = "session"
)

var (
	jsonErrMessage = map[string]interface{}{
		"err": "Something went wrong, please update the page",
	}
)

type Sessions struct {
	mux      sync.Mutex
	sessions map[string]*session // map[sessionID]Session

	shutdown      bool
	cleanInterval time.Duration
	liveTime      time.Duration
}

type session struct {
	id         string
	tcpConnect *net.Conn
	bufMux     sync.Mutex
	buffer     []byte
	lastUpdate time.Time
}

func (s *session) run() {
	//(*s.tcpConnect).SetReadDeadline(time.Now().Add(time.Millisecond*200))
	connR := bufio.NewReader(*s.tcpConnect)

	for {
		line, _, err := connR.ReadLine()
		if err != nil {
			fmt.Println("err read session")
		}
		coloredLine := replaceColors(string(line))
		s.bufMux.Lock()
		s.buffer = append(s.buffer, []byte(coloredLine+"<br>")...)
		s.bufMux.Unlock()
	}
}

func New(conf *webconfig.System) *Sessions {
	sessionsCleanInterval, err := time.ParseDuration(conf.SessionsCleanInterval)
	if err != nil {
		panic(err)
	}
	sessionsLiveTime, err := time.ParseDuration(conf.SessionsLiveTime)
	if err != nil {
		panic(err)
	}
	return &Sessions{
		sessions:      make(map[string]*session),
		cleanInterval: sessionsCleanInterval,
		liveTime:      sessionsLiveTime,
	}
}

//
//func (s *Sessions) add(id string) {
//	tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
//	if err != nil {
//		panic(err)
//	}
//	sess := &session{
//		id:         id,
//		tcpConnect: &tcpConn,
//		lastUpdate: time.Now(),
//		buffer: []byte(""),
//	}
//	go sess.run()
//	s.mux.Lock()
//	s.sessions[id] = sess
//	s.mux.Unlock()
//}

func (s *Sessions) drop(id string) {
	s.mux.Lock()
	delete(s.sessions, id)
	s.mux.Unlock()
}

func (s *Sessions) update(id string) bool {
	s.mux.Lock()
	defer s.mux.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
		if err != nil {
			panic(err)
		}
		sess = &session{
			id:         id,
			tcpConnect: &tcpConn,
			lastUpdate: time.Now(),
			buffer:     []byte(""),
		}
		go sess.run()
	} else {
		sess.lastUpdate = time.Now()
	}
	s.sessions[id] = sess
	return ok
}

func (s *Sessions) get(id string) *session {
	s.mux.Lock()
	defer s.mux.Unlock()
	return s.sessions[id]
}

func (s *Sessions) manager() {
	for !s.shutdown {
		time.Sleep(s.cleanInterval)
		fmt.Println("manager running")

		oldestLastUpdate := time.Now().Add(-s.liveTime)

		s.mux.Lock()
		var sessionsCopy = make(map[string]*session)
		for id, session := range s.sessions {
			if session.lastUpdate.After(oldestLastUpdate) {
				sessionsCopy[id] = session
			}
		}
		s.sessions = sessionsCopy
		s.mux.Unlock()
	}
}

func (s *Sessions) Shutdown() {
	s.shutdown = true
}

func (s *Sessions) Interface(w http.ResponseWriter, r *http.Request) {
	var sessionID string

	cookie, err := r.Cookie(sessionCookie)
	if err != nil || cookie.Value == "" {
		sessionID = strconv.Itoa(rand.Int())
	} else {
		sessionID = cookie.Value
	}

	s.update(sessionID)
	//if !ok {
	//s.add(sessionID)
	//}

	setCookieDays(w, 1, sessionID)

	tmpl.Execute(w, nil)
}

func (s *Sessions) Message(w http.ResponseWriter, r *http.Request) {

	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sessionID := cookie.Value

	//flusher, ok := w.(http.Flusher)

	//if !ok {
	//	http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
	//	return
	//}
	//w.Header().Set("Content-Type", "text/event-stream")
	//w.Header().Set("Cache-Control", "no-cache")
	//w.Header().Set("Connection", "keep-alive")
	//w.Header().Set("Access-Control-Allow-Origin", "*")

	ok := s.update(sessionID)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	sess := s.get(sessionID)
	sess.bufMux.Lock()
	w.Write(sess.buffer)
	sess.buffer = []byte("")
	sess.bufMux.Unlock()

	//flusher.Flush()
}

func (s *Sessions) Command(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	id, err := r.Cookie(sessionCookie)
	if err != nil || id.Value == "" {
		encoder.Encode(jsonErrMessage)
		return
	}

	command := r.FormValue("cmd")
	if command == "QUIT" {
		s.drop(id.Value)
		setCookieDays(w, 0, id.Value)
		http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
		return
	}

	ok := s.update(id.Value)
	if !ok {
		setCookieDays(w, 0, id.Value)
		http.Redirect(w, r, r.URL.String(), http.StatusPermanentRedirect)
		return
	}
	sess := s.get(id.Value)
	conn := *sess.tcpConnect
	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		encoder.Encode(jsonErrMessage)
	}

	setCookieDays(w, 1, id.Value)
}

func setCookieDays(w http.ResponseWriter, days int, cookie string) {
	expire := time.Now().AddDate(0, 0, days)
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   cookie,
		Expires: expire,
	})
}

var tmpl = template.Must(template.New("main_page").Parse(`
<html>
<head>
</head>
<body>
	<div id="message"></div>
	<label>command: <input id="command" type="text" name="command"></label>
	<input id="send" type="button" value="Enter">
	<!-- <input id="otherAccount" type="button" value="Other Account"> -->
<script>
(function() {
	var messageField = document.getElementById("message");
	setInterval(function() {
		fetch('/message')
  		.then(response => response.text())
		.then(data => {
			if (data == "" || data == undefined || data == null) {
				return
			}
			messageField.innerHTML += data+"\n";
			window.scrollTo(0,document.body.scrollHeight);
		});
	}, 1000);
	var commandField = document.getElementById("command");
	commandField.onkeydown = function(e) {
		if (e.keyCode != 13) {
			return
		}
		fetch('/command?cmd='+command.value);
		command.value = "";
	}
	var sendButton = document.getElementById("send");
	send.onclick = function(e) {
		fetch('/command?cmd='+command.value);
		command.value = "";
	}
	
	//var otherAccountButton = document.getElementById("otherAccount");
	//otherAccountButton.onclick = function() {
	//	var cookies = document.cookie.split(";");
	//	for (var i = 0; i < cookies.length; i++) {
	//		var cookie = cookies[i];
	//		var eqPos = cookie.indexOf("=");
	//		var name = eqPos > -1 ? cookie.substr(0, eqPos) : cookie;
	//		document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT";
	//	}
	//	window.location.reload(true);
	//}
}())
</script>
</body>
</html>
`))
