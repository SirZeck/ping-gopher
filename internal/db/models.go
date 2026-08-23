package db

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MonitorStatus defines the operational status of a monitor target.
type MonitorStatus string

const (
	StatusUp       MonitorStatus = "UP"
	StatusDown     MonitorStatus = "DOWN"
	StatusDegraded MonitorStatus = "DEGRADED"
	StatusPaused   MonitorStatus = "PAUSED"
)

// IncidentStatus defines the resolution lifecycle of an incident.
type IncidentStatus string

const (
	IncidentOpen          IncidentStatus = "OPEN"
	IncidentInvestigating IncidentStatus = "INVESTIGATING"
	IncidentResolved      IncidentStatus = "RESOLVED"
)

// User represents tenant user accounts.
type User struct {
	ID           uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	Email        string    `gorm:"type:varchar(255);uniqueIndex;not null" json:"email"`
	PasswordHash string    `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	Monitors []Monitor `gorm:"foreignKey:UserID;constraint:OnDelete:CASCADE" json:"monitors,omitempty"`
}

// BeforeCreate GORM hook to generate UUID if not set.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

// Monitor represents an external uptime or synthetic monitoring target.
type Monitor struct {
	ID                   uuid.UUID     `gorm:"type:uuid;primaryKey" json:"id"`
	UserID               uuid.UUID     `gorm:"type:uuid;index;not null" json:"user_id"`
	User                 User          `gorm:"foreignKey:UserID" json:"-"`
	Name                 string        `gorm:"type:varchar(255);not null" json:"name"`
	URL                  string        `gorm:"type:text;not null" json:"url"`
	CheckIntervalSeconds int           `gorm:"default:60;not null" json:"check_interval_seconds"`
	Status               MonitorStatus `gorm:"type:varchar(50);default:'PAUSED';not null" json:"status"`
	SSLExpirationDate    *time.Time    `gorm:"index" json:"ssl_expiration_date,omitempty"`
	WebhookURL           string        `gorm:"type:text" json:"webhook_url,omitempty"`
	IsPublic             bool          `gorm:"default:true;not null" json:"is_public"`
	CreatedAt            time.Time     `json:"created_at"`
	UpdatedAt            time.Time     `json:"updated_at"`

	PingLogs  []PingLog  `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE" json:"ping_logs,omitempty"`
	Incidents []Incident `gorm:"foreignKey:MonitorID;constraint:OnDelete:CASCADE" json:"incidents,omitempty"`
}

// BeforeCreate GORM hook to generate UUID if not set.
func (m *Monitor) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

// PingLog records the result of an individual synthetic probe or HTTP check execution.
type PingLog struct {
	ID             uuid.UUID `gorm:"type:uuid;primaryKey" json:"id"`
	MonitorID      uuid.UUID `gorm:"type:uuid;index:idx_monitor_created,priority:1;not null" json:"monitor_id"`
	Monitor        Monitor   `gorm:"foreignKey:MonitorID" json:"-"`
	StatusCode     int       `gorm:"not null" json:"status_code"`
	ResponseTimeMS int       `gorm:"not null" json:"response_time_ms"`
	ErrorMessage   string    `gorm:"type:text" json:"error_message,omitempty"`
	CreatedAt      time.Time `gorm:"index:idx_monitor_created,priority:2" json:"created_at"`
}

// BeforeCreate GORM hook to generate UUID if not set.
func (p *PingLog) BeforeCreate(tx *gorm.DB) error {
	if p.ID == uuid.Nil {
		p.ID = uuid.New()
	}
	return nil
}

// Incident tracks outages, degraded performance events, and resolution workflows.
type Incident struct {
	ID        uuid.UUID      `gorm:"type:uuid;primaryKey" json:"id"`
	MonitorID uuid.UUID      `gorm:"type:uuid;index;not null" json:"monitor_id"`
	Monitor   Monitor        `gorm:"foreignKey:MonitorID" json:"-"`
	StartedAt time.Time      `gorm:"not null" json:"started_at"`
	ResolvedAt *time.Time    `json:"resolved_at,omitempty"`
	Cause     string         `gorm:"type:text" json:"cause"`
	Status    IncidentStatus `gorm:"type:varchar(50);default:'OPEN';not null" json:"status"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
}

// BeforeCreate GORM hook to generate UUID if not set.
func (i *Incident) BeforeCreate(tx *gorm.DB) error {
	if i.ID == uuid.Nil {
		i.ID = uuid.New()
	}
	return nil
}
