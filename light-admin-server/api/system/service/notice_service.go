package service

import (
	"context"
	"strings"

	"github.com/top-system/light-admin/api/system/repository"
	"github.com/top-system/light-admin/db/store"
	"github.com/top-system/light-admin/errors"
	"github.com/top-system/light-admin/lib"
	"github.com/top-system/light-admin/models/dto"
	"github.com/top-system/light-admin/models/system"
)

// NoticeService service layer
type NoticeService struct {
	logger               lib.Logger
	txManager            lib.TxManager
	noticeRepository     repository.NoticeRepository
	userNoticeRepository repository.UserNoticeRepository
	userRepository       repository.UserRepository
}

// NewNoticeService creates a new notice service
func NewNoticeService(
	logger lib.Logger,
	txManager lib.TxManager,
	noticeRepository repository.NoticeRepository,
	userNoticeRepository repository.UserNoticeRepository,
	userRepository repository.UserRepository,
) NoticeService {
	return NoticeService{
		logger:               logger,
		txManager:            txManager,
		noticeRepository:     noticeRepository,
		userNoticeRepository: userNoticeRepository,
		userRepository:       userRepository,
	}
}

// Query 分页查询通知公告
func (a NoticeService) Query(param *system.NoticeQueryParam) (*system.NoticeQueryResult, error) {
	return a.noticeRepository.Query(param)
}

// Get 获取通知公告
func (a NoticeService) Get(id string) (*system.Notice, error) {
	return a.noticeRepository.Get(id)
}

// GetForm 获取通知公告表单数据
func (a NoticeService) GetForm(id string) (*system.NoticeForm, error) {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return nil, err
	}

	var targetUserIds []string
	if notice.TargetUserIds != "" {
		targetUserIds = strings.Split(notice.TargetUserIds, ",")
	}

	return &system.NoticeForm{
		ID:            notice.ID,
		Title:         notice.Title,
		Content:       notice.Content,
		Type:          dto.FlexInt(notice.Type),
		Level:         notice.Level,
		TargetType:    notice.TargetType,
		TargetUserIds: targetUserIds,
	}, nil
}

// GetDetail 获取通知公告详情并标记为已读
func (a NoticeService) GetDetail(id string, userID string) (*system.NoticeDetailVO, error) {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return nil, err
	}

	// 标记为已读
	_ = a.userNoticeRepository.MarkAsRead(id, userID)

	// 获取发布人信息
	var publisherName string
	if notice.PublisherId != "" {
		publisher, err := a.userRepository.Get(notice.PublisherId)
		if err == nil && publisher != nil {
			publisherName = publisher.Nickname
		}
	}

	return &system.NoticeDetailVO{
		ID:            notice.ID,
		Title:         notice.Title,
		Content:       notice.Content,
		Type:          notice.Type,
		Level:         notice.Level,
		PublisherId:   notice.PublisherId,
		PublisherName: publisherName,
		PublishTime:   notice.PublishTime,
	}, nil
}

// Create 创建通知公告
func (a NoticeService) Create(form *system.NoticeForm, createdBy string) error {
	// 如果目标类型是指定用户，则必须填写目标用户
	if form.TargetType == 2 && len(form.TargetUserIds) == 0 {
		return errors.New("推送指定用户不能为空")
	}

	notice := &system.Notice{
		Title:         form.Title,
		Content:       form.Content,
		Type:          form.Type.Value(),
		Level:         form.Level,
		TargetType:    form.TargetType,
		TargetUserIds: strings.Join(form.TargetUserIds, ","),
		PublishStatus: 0, // 未发布
		CreateBy:      createdBy,
		IsDeleted:     0,
	}

	return a.noticeRepository.Create(notice)
}

// Update 更新通知公告
func (a NoticeService) Update(id string, form *system.NoticeForm, updatedBy string) error {
	// 检查通知是否存在
	_, err := a.noticeRepository.Get(id)
	if err != nil {
		return err
	}

	// 如果目标类型是指定用户，则必须填写目标用户
	if form.TargetType == 2 && len(form.TargetUserIds) == 0 {
		return errors.New("推送指定用户不能为空")
	}

	notice := &system.Notice{
		ID:            id,
		Title:         form.Title,
		Content:       form.Content,
		Type:          form.Type.Value(),
		Level:         form.Level,
		TargetType:    form.TargetType,
		TargetUserIds: strings.Join(form.TargetUserIds, ","),
		UpdateBy:      updatedBy,
	}

	return a.noticeRepository.Update(id, notice)
}

// Delete 删除通知公告
func (a NoticeService) Delete(ids string, deletedBy string) error {
	if ids == "" {
		return errors.New("删除的通知公告数据为空")
	}

	idStrs := strings.Split(ids, ",")
	idList := make([]string, 0, len(idStrs))
	for _, idStr := range idStrs {
		id := strings.TrimSpace(idStr)
		if id == "" {
			continue
		}
		idList = append(idList, id)
	}

	if len(idList) == 0 {
		return errors.New("删除的通知公告数据为空")
	}

	// Delete the notices and their per-user state atomically.
	return a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		if err := a.noticeRepository.WithTx(q).BatchDelete(idList, deletedBy); err != nil {
			return err
		}
		return a.userNoticeRepository.WithTx(q).DeleteByNoticeIDs(idList)
	})
}

// Publish 发布通知公告
func (a NoticeService) Publish(id string, publisherId string) error {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return err
	}

	if notice.PublishStatus == 1 {
		return errors.New("通知公告已发布")
	}

	if notice.TargetType == 2 && notice.TargetUserIds == "" {
		return errors.New("推送指定用户不能为空")
	}

	// 获取目标用户列表
	var targetUsers system.Users
	if notice.TargetType == 1 {
		// 全体用户
		userQR, err := a.userRepository.Query(&system.UserQueryParam{})
		if err != nil {
			return err
		}
		targetUsers = userQR.List
	} else {
		// 指定用户
		targetUserIds := strings.Split(notice.TargetUserIds, ",")
		userIds := make([]string, 0, len(targetUserIds))
		for _, idStr := range targetUserIds {
			id := strings.TrimSpace(idStr)
			if id == "" {
				continue
			}
			userIds = append(userIds, id)
		}
		// 这里简化处理，实际应该根据userIds查询用户
		userQR, err := a.userRepository.Query(&system.UserQueryParam{})
		if err != nil {
			return err
		}
		for _, user := range userQR.List {
			for _, uid := range userIds {
				if user.ID == uid {
					targetUsers = append(targetUsers, user)
					break
				}
			}
		}
	}

	// 创建用户通知记录
	userNotices := make([]*system.UserNotice, 0, len(targetUsers))
	for _, user := range targetUsers {
		userNotices = append(userNotices, &system.UserNotice{
			NoticeID: id,
			UserID:   user.ID,
			IsRead:   0,
		})
	}

	// Publish the notice, clear any prior recipient state (re-publish), and fan out
	// the new per-user records atomically.
	return a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		if err := a.noticeRepository.WithTx(q).UpdateStatus(id, 1, publisherId); err != nil {
			return err
		}
		userNoticeRepo := a.userNoticeRepository.WithTx(q)
		if err := userNoticeRepo.DeleteByNoticeID(id); err != nil {
			return err
		}
		if len(userNotices) > 0 {
			return userNoticeRepo.BatchCreate(userNotices)
		}
		return nil
	})
}

// Revoke 撤回通知公告
func (a NoticeService) Revoke(id string, updatedBy string) error {
	notice, err := a.noticeRepository.Get(id)
	if err != nil {
		return err
	}

	if notice.PublishStatus != 1 {
		return errors.New("通知公告未发布或已撤回")
	}

	// Revoke the notice and clear its per-user state atomically.
	return a.txManager.RunInTx(context.Background(), func(q store.Store) error {
		if err := a.noticeRepository.WithTx(q).UpdateStatus(id, -1, updatedBy); err != nil {
			return err
		}
		return a.userNoticeRepository.WithTx(q).DeleteByNoticeID(id)
	})
}

// GetMyNoticePage 获取我的通知公告分页列表
func (a NoticeService) GetMyNoticePage(param *system.NoticeQueryParam) ([]system.UserNoticePageVO, int64, error) {
	return a.userNoticeRepository.GetMyNoticePage(param)
}

// ReadAll 全部标记为已读
func (a NoticeService) ReadAll(userID string) error {
	return a.userNoticeRepository.MarkAllAsRead(userID)
}
