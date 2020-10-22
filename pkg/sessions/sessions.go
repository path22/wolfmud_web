package sessions

import (
	"bufio"
	"fmt"
	webconfig "github.com/path22/wolfmud_web/pkg/config"
	"io"
	"math/rand"
	"net"
	"net/http"
	"path"
	"strconv"
	"sync"
	"text/template"
	"time"

	"code.wolfmud.org/WolfMUD.git/config"
)

const (
	sessionCookie     = "session"
	disconnectPattern = "EOFDisconnect"
)

var (
	errMessage = []byte(`<span style="color:white;">Something went wrong, please update the page</span><br>`)
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
	lastBuffer []byte
	lastUpdate time.Time
}

func (s *session) run() {
	//(*s.tcpConnect).SetReadDeadline(time.Now().Add(time.Millisecond*200))
	connR := bufio.NewReader(*s.tcpConnect)

	for {
		line, err := connR.ReadString('>')
		if err != nil {
			if err == io.EOF {
				s.bufMux.Lock()
				s.buffer = append(s.buffer, []byte(disconnectPattern)...)
				s.bufMux.Unlock()
			}
			fmt.Println(err)
			(*s.tcpConnect).Close()
			break
		}
		coloredLine := replaceColors(line)
		s.bufMux.Lock()
		s.buffer = append(s.buffer, []byte(coloredLine)...)
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

	setCookieDays(w, 1, sessionID)

	tmpl.Execute(w, nil)
}

func (s *Sessions) Message(w http.ResponseWriter, r *http.Request) {

	useLastBuffer, _ := strconv.ParseBool(r.FormValue("last"))

	cookie, err := r.Cookie(sessionCookie)
	if err != nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	sessionID := cookie.Value

	ok := s.update(sessionID)
	if !ok {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	sess := s.get(sessionID)
	func() {
		sess.bufMux.Lock()
		defer sess.bufMux.Unlock()
		w.Write(sess.buffer)
		if useLastBuffer {
			w.Write(sess.lastBuffer)
		}
		if len(sess.buffer) != 0 {
			sess.lastBuffer = sess.buffer
		}
		sess.buffer = []byte("")
	}()
}

func (s *Sessions) Command(w http.ResponseWriter, r *http.Request) {

	id, err := r.Cookie(sessionCookie)
	if err != nil || id.Value == "" {
		w.Write(errMessage)
		return
	}

	command := r.FormValue("cmd")
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
		w.Write(errMessage)
		return
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

var tmpl = template.Must(template.ParseGlob(path.Join("templates", "*")))
