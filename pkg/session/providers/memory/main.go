package providers

import (
	"container/list"
	"sync"
	"time"

	"github.com/meddion/web-blog/pkg/session"
)

var provider = &Provider{
	sessions: make(map[string]*list.Element, 0),
	list:     list.New(),
}

func init() {
	session.Register("memory", provider)
}

type SessionStore struct {
	id           string
	timeAccessed time.Time
	value        map[interface{}]interface{}
}

func (s *SessionStore) Set(key, value interface{}) error {
	provider.SessionUpdate(s.id)
	s.value[key] = value
	return nil
}

func (s *SessionStore) Get(key interface{}) interface{} {
	provider.SessionUpdate(s.id)
	if v, ok := s.value[key]; ok {
		return v
	}
	return nil
}

func (s *SessionStore) Delete(key interface{}) error {
	provider.SessionUpdate(s.id)
	delete(s.value, key)
	return nil
}

func (s *SessionStore) IsValuePresent(key interface{}) bool {
	provider.SessionUpdate(s.id)
	_, ok := s.value[key]
	return ok
}

func (s *SessionStore) GetSessionID() string {
	provider.SessionUpdate(s.id)
	return s.id
}

type Provider struct {
	sessions map[string]*list.Element
	list     *list.List
	sync.Mutex
}

func (p *Provider) SessionInit(id string) (session.Session, error) {
	p.Lock()
	defer p.Unlock()
	newSession := &SessionStore{
		id:           id,
		timeAccessed: time.Now(),
		value:        make(map[interface{}]interface{}, 0),
	}
	p.sessions[id] = p.list.PushFront(newSession)
	return newSession, nil
}

func (p *Provider) SessionRead(id string) (session.Session, error) {
	if element, ok := p.sessions[id]; ok {
		return element.Value.(*SessionStore), nil
	}
	newSession, err := p.SessionInit(id)
	return newSession, err
}

func (p *Provider) SessionDestroy(id string) error {
	if element, ok := p.sessions[id]; ok {
		delete(p.sessions, id)
		p.list.Remove(element)
	}
	return nil
}

func (p *Provider) SessionUpdate(id string) error {
	p.Lock()
	defer p.Unlock()
	if element, ok := p.sessions[id]; ok {
		element.Value.(*SessionStore).timeAccessed = time.Now()
		p.list.MoveToFront(element)
		return nil
	}
	return nil
}

func (p *Provider) SessionGC(maxLifeTime int64) {
	p.Lock()
	defer p.Unlock()
	for {
		element := p.list.Back()
		if element == nil ||
			element.Value.(*SessionStore).timeAccessed.Unix() < time.Now().Unix()+maxLifeTime {
			break
		}
		p.list.Remove(element)
		delete(p.sessions, element.Value.(*SessionStore).id)
	}
}
