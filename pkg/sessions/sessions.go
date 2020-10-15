package sessions

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	webconfig "github.com/path22/wolfmud_web/pkg/config"
	"math/rand"
	"net"
	"net/http"
	"strconv"
	"strings"
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

func (s *Sessions) update(id string) {
	s.mux.Lock()
	session := s.sessions[id]
	session.lastUpdate = time.Now()
	s.sessions[id] = session
	s.mux.Unlock()
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

	s.add(sessionID)

	expire := time.Now().AddDate(0, 0, 1)
	http.SetCookie(w, &http.Cookie{
		Name:    sessionCookie,
		Value:   sessionID,
		Expires: expire,
	})

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

	if !ok {
		http.Error(w, "Streaming unsupported!", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	conn := s.get(sessionID)
	connR := bufio.NewReader(*conn)

	for {
		line, _, err := connR.ReadLine()
		if err != nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Write(bytes.Join([][]byte{line, []byte("\n")}, []byte("")))
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
		return
	}

	connection := s.get(id.Value)
	conn := *connection
	_, err = conn.Write([]byte(command + "\n"))
	if err != nil {
		encoder.Encode(jsonErrMessage)
	}
	var response string
	rdr := bufio.NewReader(conn)
	conn.SetReadDeadline(time.Now().Add(time.Millisecond * 10))
	func() {
		defer func() {
			err := recover().(error)
			if !strings.Contains(err.Error(), "timeout") {
				panic(err)
			}
		}()
		for {
			line, err := rdr.ReadString('\n')
			if err != nil {
				panic(err)
			}
			response += line
		}
	}()
	encoder.Encode(map[string]interface{}{
		"ans": string(response),
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
	<input id="otherAccount" type="button" value="Other Account">
<script>
(function() {
	var messageField = document.getElementById("message");
	messageField.src = window.location + "/message"
	var commandField = document.getElementById("command");
	commandField.onkeydown = function(e) {
		if (e.keyCode != 13) {
			return
		}
		fetch('/command?cmd='+command.value)
		  .then(response => response.json())
		  .then(data => {
			var ans = document.getElementById("answer");
			ans.innerHTML = data.ans
		  });
	}
	var sendButton = document.getElementById("send");
	send.onclick = function(e) {
		fetch('/command?cmd='+command.value)
		  .then(response => response.json())
		  .then(data => {
			var ans = document.getElementById("answer");
			ans.innerHTML = data.ans
		  });
	}
	
	var otherAccountButton = document.getElementById("otherAccount");
	otherAccountButton.onclick = function() {
		var cookies = document.cookie.split(";");
		for (var i = 0; i < cookies.length; i++) {
			var cookie = cookies[i];
			var eqPos = cookie.indexOf("=");
			var name = eqPos > -1 ? cookie.substr(0, eqPos) : cookie;
			document.cookie = name + "=;expires=Thu, 01 Jan 1970 00:00:00 GMT";
		}
	window.location.reload(true);
	}
}())
</script>
</body>
</html>
`))
