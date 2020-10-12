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

	templateParams := map[string]interface{}{
		"SessionID": sessionID,
	}
	tmpl.Execute(w, templateParams)
}

func (s *Sessions) Command(w http.ResponseWriter, r *http.Request) {
	encoder := json.NewEncoder(w)

	id, err := r.Cookie(sessionCookie)
	if err != nil || id.Value == "" {
		encoder.Encode(jsonErrMessage)
		return
	}

	command := r.FormValue("cmd")

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
	<label>command: <input id="command" type="text" name="command"></label>
	<input id="send" type="button" value="Send">
	<h3 id="answer"></h3>
<script>
(function() {
	var command = document.getElementById("command");
	var send = document.getElementById("send");
	send.onclick = function() {
		fetch('/command?connection={{.ConnectionID}}&cmd='+command.value)
		  .then(response => response.json())
		  .then(data => {
			var ans = document.getElementById("answer");
			ans.innerHTML = data.ans
		  });
	}
}())
</script>
</body>
</html>
`))
