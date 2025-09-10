package mqtt

import "net/url"

type ConnectionNotificationType int64

const (
	ConnectionNotificationTypeUndefined ConnectionNotificationType = iota
	ConnectionNotificationTypeConnected
	ConnectionNotificationTypeConnectionLost
	ConnectionNotificationTypeReconnecting
	ConnectionNotificationTypeConnectionAttempt
	ConnectionNotificationTypeConnectionAttemptFailed
	// ConnectionNotificationTypeConnectionRetry
	// ConnectionNotificationTypeConnectionRetryFailed
)

type ConnectionNotification interface {
	Type() ConnectionNotificationType
}

// Connected

type ConnectionNotificationConnected struct {
	Client Client
}

func (n ConnectionNotificationConnected) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnected
}

// ConnectionLost

type ConnectionNotificationConnectionLost struct {
	Client Client
	Reason error // may be nil
}

func (n ConnectionNotificationConnectionLost) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnectionLost
}

// Reconnecting

type ConnectionNotificationReconnecting struct {
	Client Client
	Reason error // may be nil
}

func (n ConnectionNotificationReconnecting) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeReconnecting
}

// ConnectionAttempt

type ConnectionNotificationConnectionAttempt struct {
	Broker *url.URL
}

func (n ConnectionNotificationConnectionAttempt) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnectionAttempt
}

// ConnectionAttemptFailed

type ConnectionNotificationConnectionAttemptFailed struct {
	Broker *url.URL
	Reason error
}

func (n ConnectionNotificationConnectionAttemptFailed) Type() ConnectionNotificationType {
	return ConnectionNotificationTypeConnectionAttemptFailed
}
