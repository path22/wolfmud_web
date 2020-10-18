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
	sessions map[string]session // map[sessionID]Session

	shutdown      bool
	cleanInterval time.Duration
	liveTime      time.Duration
}

type session struct {
	id         string
	tcpConnect *net.Conn
	lastUpdate time.Time
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
		sessions:      make(map[string]session),
		cleanInterval: sessionsCleanInterval,
		liveTime:      sessionsLiveTime,
	}
}

func (s *Sessions) add(id string) {
	tcpConn, err := net.Dial("tcp", fmt.Sprintf("%s:%s", config.Server.Host, config.Server.Port))
	if err != nil {
		panic(err)
	}
	s.mux.Lock()
	s.sessions[id] = session{
		id:         id,
		tcpConnect: &tcpConn,
		lastUpdate: time.Now(),
	}
	s.mux.Unlock()
}

func (s *Sessions) drop(id string) {
	s.mux.Lock()
	delete(s.sessions, id)
	s.mux.Unlock()
}

func (s *Sessions) update(id string) bool {
	s.mux.Lock()
	session, ok := s.sessions[id]
	session.lastUpdate = time.Now()
	s.sessions[id] = session
	s.mux.Unlock()
	return ok
}

func (s *Sessions) get(id string) *net.Conn {
	s.mux.Lock()
	defer s.mux.Unlock()

	return s.sessions[id].tcpConnect
}

func (s *Sessions) manager() {
	for !s.shutdown {
		time.Sleep(s.cleanInterval)
		fmt.Println("manager running")

		oldestLastUpdate := time.Now().Add(-s.liveTime)

		s.mux.Lock()
		var sessionsCopy = make(map[string]session)
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

	ok := s.update(sessionID)
	if !ok {
		s.add(sessionID)
	}

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

	flusher, ok := w.(http.Flusher)

	notify := w.(http.CloseNotifier).CloseNotify()

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	ok = s.update(cookie.Value)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	conn := s.get(sessionID)
	connR := bufio.NewReader(*conn)

loop:
	for {
		select {
		case <-notify:
			break loop
		case <-time.After(time.Millisecond):
		}
		line, _, err := connR.ReadLine()
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		coloredLine := replaceColors(string(line))
		w.Write([]byte(coloredLine + "\n"))
		flusher.Flush()
	}
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
	connection := s.get(id.Value)
	conn := *connection
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
	<iframe id="message">
	</iframe>
	<label>command: <input id="command" type="text" name="command"></label>
	<input id="send" type="button" value="Enter">
	<!-- <input id="otherAccount" type="button" value="Other Account"> -->
<script>
(function() {
	var messageField = document.getElementById("message");
	messageField.src = window.location.href + "message"
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
