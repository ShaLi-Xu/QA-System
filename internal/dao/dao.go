package dao

import (
	"context"
	"time"

	"QA-System/internal/model"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

// Dao 数据访问对象
type Dao struct {
	orm   *gorm.DB
	mongo *mongo.Database
}

// New 实例化数据访问对象
func New(orm *gorm.DB, mongodb *mongo.Database) Daos {
	dao := &Dao{
		orm:   orm,
		mongo: mongodb,
	}
	// 初始化用户邮箱缓存
	dao.InitializeCache()
	return dao
}

// Daos 数据访问对象接口
type Daos interface {
	SaveAnswerSheet(ctx context.Context, answerSheet AnswerSheet, qids []int) error
	GetAnswerSheetBySurveyID(
		ctx context.Context, surveyID int64, pageNum int, pageSize int, text string, unique bool) (
		[]AnswerSheet, *int64, error)
	GetAnswerSheetBySurveyIDAndStudentID(
		ctx context.Context, surveyID int64, studentID string, pageNum int, pageSize int, text string, unique bool) (
		[]AnswerSheet, *int64, error)
	DeleteAnswerSheetBySurveyID(ctx context.Context, surveyID int64) error
	DeleteAnswerSheetByAnswerID(ctx context.Context, answerID primitive.ObjectID) error
	GetAnswerSheetByAnswerID(ctx context.Context, answerID primitive.ObjectID) error

	CreateManage(ctx context.Context, id int, surveyID int64) error
	DeleteManage(ctx context.Context, id int, surveyID int64) error
	DeleteManageBySurveyID(ctx context.Context, surveyID int64) error
	CheckManage(ctx context.Context, id int, surveyID int64) error
	GetManageByUIDAndSID(ctx context.Context, uid int, sid int64) (*model.Manage, error)
	GetManageByUserID(ctx context.Context, uid int) ([]model.Manage, error)

	CreateOption(ctx context.Context, option model.Option) error
	GetOptionsByQuestionID(ctx context.Context, questionID int) ([]model.Option, error)
	DeleteOption(ctx context.Context, questionID int) error
	GetOptionByQIDAndAnswer(ctx context.Context, qid int, answer string) (*model.Option, error)
	GetOptionByQIDAndSerialNum(ctx context.Context, qid int, serialNum int) (*model.Option, error)

	CreateQuestion(ctx context.Context, question model.Question) (model.Question, error)
	GetQuestionsBySurveyID(ctx context.Context, surveyID int64) ([]model.Question, error)
	GetQuestionByID(ctx context.Context, questionID int) (*model.Question, error)
	DeleteQuestion(ctx context.Context, questionID int) error
	DeleteQuestionBySurveyID(ctx context.Context, surveyID int64) error
	CreateType(ctx context.Context, name string, value string) error
	GetType(ctx context.Context, name string) (string, error)

	SaveRecordSheet(ctx context.Context, answerSheet RecordSheet, sid int64) error
	DeleteRecordSheets(ctx context.Context, surveyID int64) error

	CreateSurvey(ctx context.Context, survey model.Survey) (model.Survey, error)
	UpdateSurveyStatus(ctx context.Context, surveyID int64, status int) error
	UpdateSurvey(ctx context.Context, id int64, surveyType, limit uint,
		sumLimit uint, verify bool, undergradOnly bool, desc string, title string, deadline, startTime time.Time,
		needNotify bool) error
	GetSurveyByUserID(ctx context.Context, userId int) ([]model.Survey, error)
	GetSurveyByID(ctx context.Context, surveyID int64) (*model.Survey, error)
	GetAllSurvey(ctx context.Context) ([]model.Survey, error)
	IncreaseSurveyNum(ctx context.Context, sid int64) error
	DeleteSurvey(ctx context.Context, surveyID int64) error

	GetUserByUsername(ctx context.Context, username string) (*model.User, error)
	GetUserByID(ctx context.Context, id int) (*model.User, error)
	CreateUser(ctx context.Context, user *model.User) error
	UpdateUserPassword(ctx context.Context, uid int, password string) error
	UpdateUserEmail(ctx context.Context, uid int, email string) error
	GetUserEmailByID(ctx context.Context, uid int) (string, error)
}
