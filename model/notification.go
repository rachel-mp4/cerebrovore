package model

type notificationMessage struct {
	Type string            `json:"type"`
	Data notificationEvent `json:"data"`
}

type notificationEvent struct {
	Count *int  `json:"count,omitempty"` // amount to increment our notification count by
	Clear *bool `json:"clear,omitempty"`
	// mostly we care because certain gets might have a lot of replies to you resolve at once
	// which means your channel would get clogged otherwise
}

func (m *Model) DispatchNotification(username string, count *int) {
	m.dispatch(username, notificationEvent{Count: count})
}

func (m *Model) BulkDispatch(usernames []string, count *int) {
	for _, username := range usernames {
		m.dispatch(username, notificationEvent{Count: count})
	}
}

func (m *Model) BULKDispatch() {
	nm := notificationMessage{Type: "notification", Data: notificationEvent{}}
	m.watchersmu.RLock()
	defer m.watchersmu.RUnlock()
	for _, uwctx := range m.watchers {
		uwctx.ocmu.RLock()
		for conn := range uwctx.openConns {
			select {
			case conn.ch <- nm:
			default:
			}
		}
		uwctx.ocmu.RUnlock()
	}
}

func (m *Model) DispatchClear(username string) {
	beep := true
	m.dispatch(username, notificationEvent{Clear: &beep})
}

func (m *Model) dispatch(username string, ne notificationEvent) {
	m.watchersmu.RLock()
	uwctx, ok := m.watchers[username]
	m.watchersmu.RUnlock()
	if !ok {
		return
	}
	nm := notificationMessage{Type: "notification", Data: ne}
	uwctx.ocmu.RLock()
	defer uwctx.ocmu.RUnlock()
	for conn := range uwctx.openConns {
		select {
		case conn.ch <- nm:
		default:
		}
	}
}
