package model

import (
	"time"

	"gorm.io/gorm"
)

// 歌词稿件
type LyricsDraft struct {
	gorm.Model

	// ===== 基本信息 =====
	Title    string
	Artists  string // 先 string(JSON) 或 text
	Album    string
	Language string

	// ===== 权限 =====
	OwnerUserID uint

	// ===== 状态机 =====
	Status        DraftStatus    `gorm:"type:varchar(20);index"`
	WorkflowStage *WorkflowStage `gorm:"type:varchar(30)"` // 仅 PRE_REVIEW 时存在

	// ===== 审核相关 =====
	RejectCount  uint
	LastRejectAt *time.Time

	// ===== 流程配置 =====
	AllowStageRollback bool `gorm:"default:true"` //是否允许 预审核阶段主动回退

	// ===== 冻结快照（进入审核时）=====
	ReviewSnapshotID *uint

	// ===== 发布 / 扩展 =====
	GithubPRURL   string
	PublishTarget string // AMLX / GITHUB / BOTH（先留）
}

// 歌词审核
type LyricsReview struct {
	gorm.Model

	DraftID        uint `gorm:"index"`
	ReviewerUserID uint

	Result string `gorm:"type:varchar(20)"` // APPROVED / REJECTED

	// 驳回专用
	RejectReason  string
	RejectToStage *WorkflowStage `gorm:"type:varchar(30)"`
}

// 阶段回滚
type StageRollback struct {
	gorm.Model
	DraftID uint `gorm:"index"`

	FromStage WorkflowStage `gorm:"type:varchar(30)"`
	ToStage   WorkflowStage `gorm:"type:varchar(30)"`

	RequestedBy uint // 贡献者
	ApprovedBy  uint // Owner

	RollbackMode RollbackMode `gorm:"type:varchar(20)"`

	Reason string
}

// 歌词版本
type LyricsVersion struct {
	gorm.Model
	DraftID uint `gorm:"index"`

	WorkflowStage WorkflowStage `gorm:"type:varchar(30)"`
	Content       string        `gorm:"type:longtext"`

	IsSnapshot bool // 是否为审核冻结版本

	CreatedBy uint
}
