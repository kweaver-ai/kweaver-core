package agentsvc

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/kweaver-ai/decision-agent/agent-factory/src/domain/valueobject/agentrespvo"
)

const (
	minSentenceLength = 10
	indexKeyTemp      = "%d:%d"
	docMarkTagRTemp   = `<i score=${score} slice_idx=${slice_idx} p='r'>%s</i>` // 参考信息
	docMarkTagSTemp   = `<i score=${score} slice_idx=${slice_idx} p='s'>%s</i>` // 句子末尾
	cutScore          = 0.8
	capScore          = 0.2
)

// 不同大模型的返回格式不同，需要更新
var docRefPatternList = []*regexp.Regexp{
	regexp.MustCompile(`第(\d+)个(?:、第(\d+)个)*和第(\d+)个参考信息`), // 匹配"第1个和第4个参考信息"
	regexp.MustCompile(`第(\d+)个参考信息`),                      // 匹配"第1个参考信息"
	regexp.MustCompile(`（参考信息(\d+)(?:, (\d+))*）`),          // 匹配"（参考信息1, 4）"
	regexp.MustCompile(`（参考信息(\d+)(?:、(\d+))*）`),           // 匹配"（参考信息1、4）"
	regexp.MustCompile(`（参考文档ID：第(\d+)个）`),                 // 匹配"（参考文档ID：第1个）"
	regexp.MustCompile(`参考文档第(\d+)个`),                      // 匹配"参考文档第1个"
	regexp.MustCompile(`参考文档第(\d+)个(?:、第(\d+)个)*和第(\d+)个`), // 匹配"参考文档第1个和第4个"
	regexp.MustCompile(`参考文档第(\d+)个(?:、第(\d+)个)*、第(\d+)个`), // 匹配"参考文档第1个、第4个"
	regexp.MustCompile(`参考文档(\d+)`),                        // 匹配"参考文档1"
	regexp.MustCompile(`参考文档(\d+)(?:, (\d+))*`),            // 匹配"参考文档1, 4"
	regexp.MustCompile(`参考信息(\d+)(?:,(\d+))*`),             // 匹配"参考信息1, 4"
	regexp.MustCompile(`参考文档(\d+)和(\d+)`),                  // 匹配"参考文档1和4"
	regexp.MustCompile(`参考文档(\d+)(?:、(\d+))*和(\d+)`),       // 匹配"参考文档1、4和5"
	regexp.MustCompile(`参考文档(\d+)(?:、(\d+))`),              // 匹配"参考文档1、4"
	regexp.MustCompile(`参考文档(\d+)(?:、(\d+))、(\d+)`),        // 匹配"参考文档1、4、5"
}

type docCite struct {
	Content string
	Index   int
	Slices  []*agentrespvo.V1Slice
}

func (agentSvc *agentSvc) addCiteDocMark(answer string, cites []*agentrespvo.CiteDoc) string {
	if len(cites) == 0 || len(answer) == 0 {
		return answer
	}

	// 1. 分句
	sentences := splitSentences(answer, minSentenceLength)

	docCites := map[int]*docCite{}
	for i, cite := range cites {
		docCites[i+1] = &docCite{
			Content: cite.Content,
			Index:   i + 1,
			Slices:  cite.Slices,
		}
	}
	// 2. 计算每个句子与每个文档切片的相似度，替换特定句子中的引用标签
	builder := strings.Builder{}

	for _, sentence := range sentences {
		// 2.1 计算句子与文档切片的相似度
		_sentenceInfo := agentSvc.getSentenceDocScore(sentence, docCites)

		// 2.2
		_sentenceInfo.HasRefrence, _sentenceInfo.DocIndexs, _sentenceInfo.Text = markInDocIndex(sentence, docRefPatternList)

		// 2.3 如果没有引用
		if !_sentenceInfo.HasRefrence {
			agentSvc.hlRef(_sentenceInfo, sentence)
		}

		// 2.4
		for _, docIndex := range _sentenceInfo.DocIndexs {
			mss, ok := _sentenceInfo.MaxScoreMap[docIndex]
			if ok {
				for _, temp := range []string{docMarkTagRTemp, docMarkTagSTemp} {
					tempStr := fmt.Sprintf(temp, docIndex)

					replace := strings.ReplaceAll(tempStr, "${score}", fmt.Sprintf("%.2f", mss.Score))

					replace = strings.ReplaceAll(replace, "${slice_idx}", fmt.Sprintf("%d", mss.SliceIndex))

					_sentenceInfo.Text = strings.ReplaceAll(_sentenceInfo.Text, tempStr, replace)
				}
			}
		}

		builder.WriteString(_sentenceInfo.Text)
	}

	// 3. 返回
	return builder.String()
}

// splitSentences 将文本按句子分割，并限制每个句子的长度
func splitSentences(text string, length int) []string {
	re := regexp.MustCompile(`[^。！？]+[。！？]*`)
	sentences := re.FindAllString(text, -1)

	var combinedSentences []string

	tempSentence := ""

	for _, sentence := range sentences {
		tempSentence += sentence
		if len(tempSentence) >= length {
			combinedSentences = append(combinedSentences, tempSentence)
			tempSentence = ""
		}
	}

	if tempSentence != "" {
		combinedSentences = append(combinedSentences, tempSentence)
	}

	return combinedSentences
}

type maxScoreSlice struct {
	DocIndex   int
	Score      float64
	SliceIndex int
}
type sentenceInfo struct {
	MaxScoreMap map[string]*maxScoreSlice
	AvgScore    float64
	HasRefrence bool
	DocIndexs   []string
	Text        string
}

func (agentSvc *agentSvc) getSentenceDocScore(sentence string, docCites map[int]*docCite) *sentenceInfo {
	sentenceInfo := &sentenceInfo{
		MaxScoreMap: map[string]*maxScoreSlice{},
		Text:        sentence,
		DocIndexs:   []string{},
	}
	totalScore := 0.0

	for _, cite := range docCites {
		for i, slice := range cite.Slices {
			text := sentence
			dom, err := goquery.NewDocumentFromReader(strings.NewReader(sentence))

			if err == nil {
				text = dom.Text()
			}

			score := agentSvc.sameWordsPercentage(text, slice.Content)
			mss, ok := sentenceInfo.MaxScoreMap[fmt.Sprint(cite.Index)]

			if ok {
				if score > mss.Score {
					mss.Score = score
					mss.SliceIndex = i
				}
			} else {
				sentenceInfo.MaxScoreMap[fmt.Sprint(cite.Index)] = &maxScoreSlice{
					DocIndex:   cite.Index,
					Score:      score,
					SliceIndex: i,
				}
			}
		}
	}

	for _, mss := range sentenceInfo.MaxScoreMap {
		totalScore += mss.Score
	}

	sentenceInfo.AvgScore = totalScore / float64(len(docCites))

	return sentenceInfo
}

func (agentSvc *agentSvc) sameWordsPercentage(sentence string, sliceContent string) float64 {
	return 0.0
}

type stringIndexInfo struct {
	Start int
	End   int
	Value string
}

func (s *stringIndexInfo) key() string {
	return fmt.Sprintf(indexKeyTemp, s.Start, s.End)
}

// 添加文档引用标签
func markInDocIndex(text string, res []*regexp.Regexp) (has bool, docIndexs []string, newText string) {
	indexMap := map[string]*stringIndexInfo{}
	sepList := []int{}
	docIndexs = []string{}

	for _, re := range res {
		matches := re.FindAllStringSubmatchIndex(text, -1)
		for _, match := range matches {
			l := len(match)
			if l > 0 && l%2 == 0 {
				for i := 1; i < l/2; i++ {
					if match[i*2] == match[i*2+1] {
						continue
					}

					info := &stringIndexInfo{
						Start: match[i*2],
						End:   match[i*2+1],
						Value: fmt.Sprintf(docMarkTagRTemp, text[match[i*2]:match[i*2+1]]),
					}
					_, ok := indexMap[info.key()]

					if !ok {
						indexMap[info.key()] = info

						sepList = append(sepList, match[i*2], match[i*2+1])
						docIndexs = append(docIndexs, text[match[i*2]:match[i*2+1]])
					}
				}
			}
		}
	}

	if len(sepList) == 0 {
		newText = text
		return
	}

	has = true

	sort.Ints(sepList)

	newStr := strings.Builder{}

	for i := 0; i < len(sepList); i++ {
		if i == 0 && sepList[i] > 0 {
			newStr.WriteString(text[:sepList[i]])
		}

		if i == len(sepList)-1 {
			if sepList[i] < len(text) {
				newStr.WriteString(text[sepList[i]:])
			}
		} else {
			info, ok := indexMap[fmt.Sprintf(indexKeyTemp, sepList[i], sepList[i+1])]
			if ok {
				newStr.WriteString(info.Value)
			} else {
				newStr.WriteString(text[sepList[i]:sepList[i+1]])
			}
		}
	}

	newText = newStr.String()

	return
}

func (agentSvc *agentSvc) hlRef(_sentenceInfo *sentenceInfo, sentence string) {
	nextScore := -1.0
	if _sentenceInfo.AvgScore+capScore < cutScore {
		nextScore = _sentenceInfo.AvgScore + capScore
	}

	var docIndexs []int

	var nextDocIndexs []int

	for _, mss := range _sentenceInfo.MaxScoreMap {
		if mss.Score > cutScore {
			docIndexs = append(docIndexs, mss.DocIndex)
		}

		if nextScore > 0 && mss.Score > nextScore {
			nextDocIndexs = append(nextDocIndexs, mss.DocIndex)
		}
	}

	if len(docIndexs) == 0 {
		docIndexs = nextDocIndexs
	}

	sort.Ints(docIndexs)

	runes := []rune(sentence)

	_sentenceInfo.Text = string(runes[:len(runes)-1])
	for _, docIndex := range docIndexs {
		_sentenceInfo.DocIndexs = append(_sentenceInfo.DocIndexs, fmt.Sprint(docIndex))
		_sentenceInfo.Text += fmt.Sprintf(docMarkTagSTemp, fmt.Sprint(docIndex))
	}

	_sentenceInfo.Text += string(runes[len(runes)-1:])
}
