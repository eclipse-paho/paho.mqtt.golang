package mqtt

import "net/url"

type ConnectionNotificationType int64

const (
	ConnectionNotificationTypeUndefined ConnectionNotificationType = iota
	ConnectionNotificationTypeConnected
	ConnectionNotificationTypeFailed
	ConnectionNotificationTypeLost
	ConnectionNotificationTypeReconnecting
	ConnectionNotificationTypeAttempt
	ConnectionNotificationTypeAttemptFailed
	ConnectionNotificationTypeRetry
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

// Reconnecting

type ConnectionNotificationReconnecting struct {
	Reason error // may be nil
}

func (n ConnectionNotificationReconnecting) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeReconnecting
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

// Connected

type ConnectionNotificationRetry struct {
	Count  int
	Reason error
}

func (n ConnectionNotificationRetry) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeRetry
}
