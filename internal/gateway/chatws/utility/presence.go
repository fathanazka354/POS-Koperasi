package utility

import (
	"strconv"
	"sync"
)

// Presence menghitung koneksi WebSocket chat aktif per member / karyawan (referensi tab/app).
type Presence struct {
	mu     sync.RWMutex
	counts map[string]int
}

func NewPresence() *Presence {
	return &Presence{counts: make(map[string]int)}
}

func memberKey(id int) string { return "m:" + strconv.Itoa(id) }
func employeeKey(id int) string { return "e:" + strconv.Itoa(id) }

func (p *Presence) ConnectMember(memberID int) {
	if memberID <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	k := memberKey(memberID)
	p.counts[k]++
}

func (p *Presence) ConnectEmployee(employeeID int) {
	if employeeID <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	k := employeeKey(employeeID)
	p.counts[k]++
}

func (p *Presence) DisconnectMember(memberID int) {
	if memberID <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	k := memberKey(memberID)
	p.counts[k]--
	if p.counts[k] <= 0 {
		delete(p.counts, k)
	}
}

func (p *Presence) DisconnectEmployee(employeeID int) {
	if employeeID <= 0 {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	k := employeeKey(employeeID)
	p.counts[k]--
	if p.counts[k] <= 0 {
		delete(p.counts, k)
	}
}

func (p *Presence) IsMemberOnline(memberID int) bool {
	if memberID <= 0 {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.counts[memberKey(memberID)] > 0
}

func (p *Presence) IsEmployeeOnline(employeeID int) bool {
	if employeeID <= 0 {
		return false
	}
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.counts[employeeKey(employeeID)] > 0
}

