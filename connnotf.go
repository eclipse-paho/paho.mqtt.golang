package mqtt

import "net/url"

type ConnectionNotificationType int64

const (
	ConnectionNotificationTypeUndefined ConnectionNotificationType = iota
	ConnectionNotificationTypeConnected
	ConnectionNotificationTypeConnecting
	ConnectionNotificationTypeFailed
	ConnectionNotificationTypeLost
	ConnectionNotificationTypeAttempt
	ConnectionNotificationTypeAttemptFailed
)

type ConnectionNotification interface {
	Type() ConnectionNotificationType
}

// Connected

type ConnectionNotificationConnected struct {
}

func (n ConnectionNotificationConnected) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnected
}

// Connecting

type ConnectionNotificationConnecting struct {
	IsReconnect bool
	Attempt     int
}

func (n ConnectionNotificationConnecting) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnecting
}

// ConnectionFailed

type ConnectionNotificationFailed struct {
	Reason error
}

func (n ConnectionNotificationFailed) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeFailed
}

// ConnectionLost

type ConnectionNotificationLost struct {
	Reason error // may be nil
}

func (n ConnectionNotificationLost) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeLost
}

// ConnectionAttempt

type ConnectionNotificationAttempt struct {
	Broker *url.URL
}

func (n ConnectionNotificationAttempt) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeAttempt
}

// ConnectionAttemptFailed

type ConnectionNotificationAttemptFailed struct {
	Broker *url.URL
	Reason error
}

func (n ConnectionNotificationAttemptFailed) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeAttemptFailed
}
