package model

type DraftStatus string

const (
	DraftPreReview  DraftStatus = "PRE_REVIEW"
	DraftInReview   DraftStatus = "IN_REVIEW"
	DraftReviewDone DraftStatus = "REVIEW_DONE"
)

type WorkflowStage string

const (
	StageLyricRequest   WorkflowStage = "LYRIC_REQUEST"
	StageLyricCompleted WorkflowStage = "LYRIC_COMPLETED"
	StageRough          WorkflowStage = "ROUGH"
	StageFine           WorkflowStage = "FINE"
	StageCheck          WorkflowStage = "CHECK"
)

type RollbackMode string

const (
	RollbackKeepFiles RollbackMode = "KEEP_FILES"
	RollbackDropFiles RollbackMode = "DROP_FILES"
)
