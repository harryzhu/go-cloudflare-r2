package cfr2

import (
	"encoding/json"
	"log"
)

type JsonResponse struct {
	Error    int    `json: error`
	Message  string `json: message`
	Data     any    `json: data`
	Function string `json: function`
	Step     int    `json: step`
}

func NewJsonResponse() JsonResponse {
	jr := JsonResponse{}
	jr.Data = ""
	jr.Message = ""
	jr.Error = 0
	jr.Function = ""
	jr.Step = 0
	return jr
}

func (j *JsonResponse) WithFunction(f string) *JsonResponse {
	j.Function = f
	return j
}

func (j *JsonResponse) WithStep(i int) *JsonResponse {
	j.Step = i
	return j
}

func (j *JsonResponse) AutoStep() *JsonResponse {
	j.Step = j.Step + 1
	return j
}

func (j *JsonResponse) WithErrorMessage(i int, m string) *JsonResponse {
	j.Error = i
	j.Message = m
	return j
}

func (j *JsonResponse) WithData(d any) *JsonResponse {
	j.Data = d
	return j
}

func (j *JsonResponse) Jsonify() string {
	s, err := json.Marshal(j)
	if err != nil {
		log.Println(err)
	}
	return string(s)
}
