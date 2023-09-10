package memdb

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/marcsello/marcsellocorp-bot/utils"
	"sync"
	"time"
)

const (
	questionKeyPrefix = "QST:"
	inflightExpire    = 300 * time.Second // un-closed entries will automatically disappear
	answeredExpire    = 2 * time.Hour     // store data about answered questions for this long
)

type NewQuestionTx struct {
	randomId  string
	data      QuestionData
	dataMutex sync.Mutex
	ctx       context.Context
}

func randomIdToKey(randomId string) string {
	return questionKeyPrefix + randomId
}

func (q *NewQuestionTx) Close() error {
	q.dataMutex.Lock()
	defer q.dataMutex.Unlock()

	if q.ctx.Err() != nil {
		return q.ctx.Err() // context cancelled probably
	}
	q.data.Ready = true

	// write final version
	var dataBytes []byte
	var err error
	dataBytes, err = json.Marshal(q.data)
	if err != nil {
		return err
	}
	result := redisClient.Set(q.ctx, q.key(), dataBytes, 0)
	return result.Err()
}

func (q *NewQuestionTx) AddRelatedMessage(message StoredMessage) {
	q.dataMutex.Lock()
	defer q.dataMutex.Unlock()
	q.data.RelatedMessages = append(q.data.RelatedMessages, message)
}

func (q *NewQuestionTx) AddOption(data, label string) {
	q.dataMutex.Lock()
	defer q.dataMutex.Unlock()
	q.data.Options = append(q.data.Options, QuestionOption{
		Data:  data,
		Label: label,
	})
}

func (q *NewQuestionTx) key() string {
	return randomIdToKey(q.randomId)
}

func (q *NewQuestionTx) RandomID() string {
	return q.randomId
}

func BeginNewQuestion(ctx context.Context, sourceToken uint) (NewQuestionTx, error) {
	var err error
	data := QuestionData{
		AnsweredAt:      nil,
		AnswererID:      nil,
		AnswerData:      nil,
		RelatedMessages: make([]StoredMessage, 0),
		SourceTokenID:   sourceToken,
		Ready:           false, // <- messages are being sent out
	}

	var dataBytes []byte
	dataBytes, err = json.Marshal(data)
	if err != nil {
		return NewQuestionTx{}, err
	}

	var newId string
	for {
		newId, err = utils.GenerateRandomString(32)
		if err != nil {
			return NewQuestionTx{}, err
		}
		newKey := randomIdToKey(newId)
		result := redisClient.SetNX(ctx, newKey, dataBytes, inflightExpire)
		var val bool
		val, err = result.Result()
		if err != nil {
			return NewQuestionTx{}, err
		}
		if val {
			break
		}
		if ctx.Err() != nil {
			return NewQuestionTx{}, ctx.Err() // context cancelled probably
		}
	}

	return NewQuestionTx{
		randomId:  newId,
		data:      data,
		dataMutex: sync.Mutex{},
		ctx:       ctx,
	}, nil

}

func GetQuestionData(ctx context.Context, randomId string) (*QuestionData, error) {
	key := randomIdToKey(randomId)
	getResult := redisClient.Get(ctx, key)

	var dataBytes []byte
	var err error
	dataBytes, err = getResult.Bytes()
	if err != nil {
		return nil, err
	}

	var data QuestionData
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return nil, err
	}

	return &data, nil
}

func AnswerQuestion(ctx context.Context, randomId string, answererID int64, answerData string) (*QuestionData, error) {
	key := randomIdToKey(randomId)
	getResult := redisClient.Get(ctx, key)

	var dataBytes []byte
	var err error
	dataBytes, err = getResult.Bytes()
	if err != nil {
		return nil, err
	}

	var data QuestionData
	err = json.Unmarshal(dataBytes, &data)
	if err != nil {
		return nil, err
	}

	if !data.Ready { // unready question does not have possible options stored either
		// TODO: eh?
		return nil, fmt.Errorf("question not delivered to all recipients, please wait")
	}

	var dataIsValid bool
	for _, validOp := range data.Options {
		if validOp.Data == answerData {
			dataIsValid = true
			break
		}
	}
	if !dataIsValid {
		return nil, fmt.Errorf("invalid answerData")
	}

	data.AnswererID = &answererID
	data.AnswerData = &answerData
	now := time.Now()
	data.AnsweredAt = &now

	dataBytes, err = json.Marshal(data)
	if err != nil {
		return nil, err
	}

	setResult := redisClient.Set(ctx, key, dataBytes, answeredExpire)
	return &data, setResult.Err()
}
