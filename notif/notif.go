package notif

import (
	"context"
	notiftypes "git.sr.ht/~bouncepaw/betula/types/notif"
	"sync"
	"time"
)

type Service struct {
	mu         sync.Mutex
	countCache *uint

	repo notiftypes.Repository
}

func (svc *Service) Count() (uint, error) {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	if svc.countCache != nil {
		return *svc.countCache, nil
	}
	count, err := svc.repo.Count(context.Background())
	if err != nil {
		return 0, err
	}
	ucount := uint(count)
	svc.countCache = &ucount
	return ucount, nil
}

func (svc *Service) InvalidateCache() {
	svc.mu.Lock()
	defer svc.mu.Unlock()

	svc.countCache = nil
}

func (svc *Service) GetAll() ([]notiftypes.NotificationGroup, error) {
	notifs, err := svc.repo.GetAll(context.Background())
	if err != nil {
		return nil, err
	}

	return GroupNotificationsByDay(notifs), nil
}

func (svc *Service) MarkAllAsRead() error {
	return svc.repo.DeleteAll(context.Background())
}

func (svc *Service) MarkAsRead(date string) error {
	_, err := time.Parse(time.DateOnly, date)
	if err != nil {
		return err
	}

	return svc.repo.DeleteDate(context.Background(), date)
}

var _ notiftypes.Service = &Service{}

func New(repo notiftypes.Repository) *Service {
	return &Service{
		repo: repo,
	}
}
