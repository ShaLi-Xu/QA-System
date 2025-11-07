package service

import (
	"bytes"
	"image"
	_ "image/gif" // 注册解码器
	"image/jpeg"
	_ "image/png" // 注册解码器
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"QA-System/internal/dao"
	"QA-System/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/zjutjh/WeJH-SDK/oauth"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.uber.org/zap"
	_ "golang.org/x/image/bmp" // 注册解码器
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// GetSurveyByID 根据ID获取问卷
func GetSurveyByID(id int64) (*model.Survey, error) {
	survey, err := d.GetSurveyByID(ctx, id)
	return survey, err
}

// GetQuestionsBySurveyID 根据问卷ID获取问题
func GetQuestionsBySurveyID(sid int64) ([]model.Question, error) {
	var questions []model.Question
	questions, err := d.GetQuestionsBySurveyID(ctx, sid)
	return questions, err
}

// GetOptionsByQuestionID 根据问题ID获取选项
func GetOptionsByQuestionID(questionId int) ([]model.Option, error) {
	var options []model.Option
	options, err := d.GetOptionsByQuestionID(ctx, questionId)
	return options, err
}

// GetQuestionByID 根据问题ID获取问题
func GetQuestionByID(id int) (*model.Question, error) {
	var question *model.Question
	question, err := d.GetQuestionByID(ctx, id)
	return question, err
}

// SubmitSurvey 提交问卷
func SubmitSurvey(sid int64, data []dao.QuestionsList, t string, stuId string) error {
	var answerSheet dao.AnswerSheet
	answerSheet.StudentID = stuId
	answerSheet.SurveyID = sid
	answerSheet.Time = t
	answerSheet.Unique = true
	answerSheet.AnswerID = primitive.NewObjectID()
	qids := make([]int, 0)
	for _, q := range data {
		var answer dao.Answer
		question, err := d.GetQuestionByID(ctx, q.QuestionID)
		if err != nil {
			return err
		}
		if question.QuestionType == 3 && question.Unique {
			qids = append(qids, q.QuestionID)
		}
		answer.QuestionID = q.QuestionID
		answer.Content = q.Answer
		answerSheet.Answers = append(answerSheet.Answers, answer)
	}
	err := d.SaveAnswerSheet(ctx, answerSheet, qids)
	if err != nil {
		return err
	}
	err = d.IncreaseSurveyNum(ctx, sid)
	if err != nil {
		return err
	}
	err = FromSurveyIDToMsg(sid)
	return err
}

// CreateOauthRecord 创建一条统一验证记录
func CreateOauthRecord(userInfo oauth.UserInfo, t time.Time, sid int64) error {
	sheet := dao.RecordSheet{
		College:      userInfo.College,
		Name:         userInfo.Name,
		StudentID:    userInfo.StudentID,
		UserType:     userInfo.UserType,
		UserTypeDesc: userInfo.UserTypeDesc,
		Gender:       userInfo.Gender,
		Time:         t,
	}
	return d.SaveRecordSheet(ctx, sheet, sid)
}

// ConvertToJPEG 将图片转换为 JPEG 格式
func ConvertToJPEG(reader io.Reader) (io.Reader, error) {
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	err = jpeg.Encode(&buf, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(buf.Bytes()), nil
}

// SaveFile 保存文件
func SaveFile(reader io.Reader, path string) error {
	dst := filepath.Clean(path)
	err := os.MkdirAll(filepath.Dir(dst), 0750)
	if err != nil {
		return err
	}

	// 创建文件
	outFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			zap.L().Error("Failed to close file", zap.Error(err))
		}
	}(outFile)

	// 写入文件
	_, err = io.Copy(outFile, reader)
	return err
}

// UpdateVoteLimit 更新投票限制
func UpdateVoteLimit(c *gin.Context, stuId string, surveyID int64, isNew bool, durationType string) error {
	if isNew {
		if durationType == "dailyLimit" {
			return SetUserLimit(c, stuId, surveyID, 1, durationType)
		}
		return SetUserSumLimit(c, stuId, surveyID, 1, durationType)
	}
	return InscUserLimit(c, stuId, surveyID, durationType)
}

func ensureMap(m map[int]map[int]int, key int) map[int]int {
	if m[key] == nil {
		m[key] = make(map[int]int)
	}
	return m[key]
}

type getOptionCount struct {
	SerialNum int    `json:"serial_num"` // 选项序号
	Content   string `json:"content"`    // 选项内容
	Count     int    `json:"count"`      // 选项数量
}

// GetAnswerRecordResponse 答卷记录的返回模型
type GetAnswerRecordResponse struct {
	SerialNum    int              `json:"serial_num"`    // 问题序号
	Question     string           `json:"question"`      // 问题内容
	QuestionType int              `json:"question_type"` // 问题类型  1:单选 2:多选
	Options      []getOptionCount `json:"options"`       // 选项内容
}

// CreateRecordResponse 创建答卷记录返回数据
func CreateRecordResponse(userAnswerSheets []dao.AnswerSheet,
	questions []model.Question) ([]GetAnswerRecordResponse, error) {
	if len(userAnswerSheets) == 0 {
		response := make([]GetAnswerRecordResponse, 0, len(questions))
		for _, q := range questions {
			options, err := GetOptionsByQuestionID(q.ID)
			if err != nil {
				return nil, err
			}

			qOptions := make([]getOptionCount, 0, len(options)+1)
			for _, option := range options {
				qOptions = append(qOptions, getOptionCount{
					SerialNum: option.SerialNum,
					Content:   option.Content,
					Count:     0,
				})
			}

			// 如果支持 "其他" 选项，添加一项
			if q.OtherOption {
				qOptions = append(qOptions, getOptionCount{
					SerialNum: 0,
					Content:   "其他",
					Count:     0,
				})
			}

			response = append(response, GetAnswerRecordResponse{
				SerialNum:    q.SerialNum,
				Question:     q.Subject,
				QuestionType: q.QuestionType,
				Options:      qOptions,
			})
		}
		return response, nil
	}

	// 问题编号对应的问题
	questionMap := make(map[int]model.Question)
	// 问题编号对应的选项们
	optionsMap := make(map[int][]model.Option)
	// 问题编号与选项内容对应的选项
	optionAnswerMap := make(map[int]map[string]model.Option)
	// 问题编号与选项序号对应的选项
	optionSerialNumMap := make(map[int]map[int]model.Option)
	for _, question := range questions {
		questionMap[question.ID] = question
		optionAnswerMap[question.ID] = make(map[string]model.Option)
		optionSerialNumMap[question.ID] = make(map[int]model.Option)
		options, err := GetOptionsByQuestionID(question.ID)
		if err != nil {
			return nil, err
		}
		optionsMap[question.ID] = options
		for _, option := range options {
			optionAnswerMap[question.ID][option.Content] = option
			optionSerialNumMap[question.ID][option.SerialNum] = option
		}
	}

	// 问题编号对应的选项编号对应的选项数量
	optionCounts := make(map[int]map[int]int)
	for _, sheet := range userAnswerSheets {
		for _, answer := range sheet.Answers {
			options := optionsMap[answer.QuestionID]
			question := questionMap[answer.QuestionID]
			// 初始化选项统计（确保每个选项的计数存在且为 0）
			if _, initialized := optionCounts[question.ID]; !initialized {
				counts := ensureMap(optionCounts, question.ID)
				for _, option := range options {
					counts[option.SerialNum] = 0
				}
			}
			if question.QuestionType == 1 {
				answerOptions := strings.Split(answer.Content, "┋")
				questionOptions := optionAnswerMap[answer.QuestionID]
				for _, answerOption := range answerOptions {
					// 查找选项
					if questionOptions != nil {
						option, exists := questionOptions[answerOption]
						if exists {
							// 如果找到选项，处理逻辑
							ensureMap(optionCounts, answer.QuestionID)[option.SerialNum]++
							continue
						}
					}
					// 如果选项不存在，处理为 "其他" 选项
					ensureMap(optionCounts, answer.QuestionID)[0]++
				}
			}
		}
	}

	response := make([]GetAnswerRecordResponse, 0, len(optionCounts))

	for qid, options := range optionCounts {
		q := questionMap[qid]
		var qOptions []getOptionCount
		if q.OtherOption {
			qOptions = make([]getOptionCount, 0, len(options)+1)
			// 添加其他选项
			qOptions = append(qOptions, getOptionCount{
				SerialNum: 0,
				Content:   "其他",
				Count:     options[0],
			})
		} else {
			qOptions = make([]getOptionCount, 0, len(options))
		}
		// 按序号排序
		sortedSerialNums := make([]int, 0, len(options))
		for oSerialNum := range options {
			sortedSerialNums = append(sortedSerialNums, oSerialNum)
		}
		sort.Ints(sortedSerialNums)
		for _, oSerialNum := range sortedSerialNums {
			count := options[oSerialNum]
			op := optionSerialNumMap[qid][oSerialNum]
			qOptions = append(qOptions, getOptionCount{
				SerialNum: op.SerialNum,
				Content:   op.Content,
				Count:     count,
			})
		}

		response = append(response, GetAnswerRecordResponse{
			SerialNum:    q.SerialNum,
			Question:     q.Subject,
			QuestionType: q.QuestionType,
			Options:      qOptions,
		})
	}
	return response, nil
}

// GetSurveyAnswersBySurveyIDAndStudentID 根据问卷编号和学生学号获取问卷答案
func GetSurveyAnswersBySurveyIDAndStudentID(surveyid int64, studentid string) ([]dao.AnswerSheet, error) {
	answerSheets, _, err := d.GetAnswerSheetBySurveyIDAndStudentID(ctx, surveyid, studentid, 0, 0, "", true)
	return answerSheets, err
}
