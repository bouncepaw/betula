package notiftypes

import "context"

type Service interface {
	Count() (uint, error)
	InvalidateCache()
	GetAll() ([]NotificationGroup, error)
	MarkAllAsRead() error
	MarkAsRead(date string) error
}

type Repository interface {
	Store(context.Context, ...Notification) error
	GetAll(context.Context) ([]Notification, error)
	Count(context.Context) (int64, error)
	DeleteAll(context.Context) error
	DeleteDate(context.Context, string) error
}
