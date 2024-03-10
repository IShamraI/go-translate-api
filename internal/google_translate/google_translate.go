package googletranslate

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

var GOOGLE_TRANSLATE_URL = "https://translate.googleapis.com/translate_a/single?client=gtx&sl=%s&tl=%s&dt=t&q=%s"

type GTranslateService struct {
	SourceLang string
	TargetLang string
}

func (s *GTranslateService) EncodeSource(source string) string {
	return url.QueryEscape(source)
}

func (s *GTranslateService) Translate(source string) (string, error) {
	var text []string
	var result []interface{}

	encodedSource := s.EncodeSource(source)
	url := fmt.Sprintf(GOOGLE_TRANSLATE_URL, s.SourceLang, s.TargetLang, encodedSource)

	r, err := http.Get(url)
	if err != nil {
		return "err", errors.New("error getting translate.googleapis.com")
	}
	defer r.Body.Close()

	body, err := io.ReadAll(r.Body)
	if err != nil {
		return "err", errors.New("error reading response body")
	}

	bReq := strings.Contains(string(body), `<title>Error 400 (Bad Request)`)
	if bReq {
		return "err", errors.New("error 400 (Bad Request)")
	}

	err = json.Unmarshal(body, &result)
	if err != nil {
		return "err", errors.New("error unmarshaling data")
	}

	if len(result) > 0 {
		inner := result[0]
		for _, slice := range inner.([]interface{}) {
			for _, translatedText := range slice.([]interface{}) {
				text = append(text, fmt.Sprintf("%v", translatedText))
				break
			}
		}
		cText := strings.Join(text, "")

		return cText, nil
	} else {
		return "err", errors.New("no translated data in responce")
	}
}

func New(sourceLang, targetLang string) *GTranslateService {
	return &GTranslateService{
		SourceLang: sourceLang,
		TargetLang: targetLang,
	}
}

func NewRuEn() *GTranslateService {
	return &GTranslateService{
		SourceLang: "ru",
		TargetLang: "en",
	}
}
