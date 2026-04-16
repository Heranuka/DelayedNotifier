package domain

type NotificationChannel string

const (
	ChannelEmail NotificationChannel = "email"
)

func (c NotificationChannel) String() string {
	return string(c)
}
