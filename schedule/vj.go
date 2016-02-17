package schedule

import (
	"GoOnlineJudge/model"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
    "net/url"
    "fmt"
)

type VJJudger struct {
	client    *http.Client
	titleRx   *regexp.Regexp
	TimeRx   *regexp.Regexp
	MemoryRx   *regexp.Regexp
	DescriptionRx   *regexp.Regexp
	InputRx   *regexp.Regexp
	OutputRx   *regexp.Regexp
	testInRx   *regexp.Regexp
	testOutRx   *regexp.Regexp
	srcRx   *regexp.Regexp
	hintRx   *regexp.Regexp
}

var VJlogger *log.Logger

func (p *VJJudger) Host() string {
	return "VJ"
}
func (p *VJJudger) Ping() error {
	p.client = &http.Client{Timeout: time.Second * 10}
	resp, err := p.client.Get("http://acm.hust.edu.cn/vjudge/problem/viewProblem.action?id=19972")
	if err != nil {
		return ErrConnectFailed
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		return nil
	}
	return ErrResponse
}

func (h *VJJudger) Init() {
	h.client = &http.Client{Timeout: time.Second * 10}

	titlePat := `<td>Title:</td><td>(.*?)</td>`
    h.titleRx = regexp.MustCompile(titlePat)

	TimePat := `<td>Time Limit:</td><td>(\d+) MS</td>`
	h.TimeRx = regexp.MustCompile(TimePat)
	MemoryPat := `<td>Memory Limit:</td><td>(\d+) KB</td>`
	h.MemoryRx = regexp.MustCompile(MemoryPat)

	DescriptionPat := `<textarea name="description.description" cols="120" rows="15" id="description">(.*?)</textarea>`
	h.DescriptionRx = regexp.MustCompile(DescriptionPat)
	InputPat := `<textarea name="description.input" cols="120" rows="15" id="input">(.*?)</textarea>`
	h.InputRx = regexp.MustCompile(InputPat)
	OutputPat := `<textarea name="description.output" cols="120" rows="15" id="output">(.*?)</textarea>`
	h.OutputRx = regexp.MustCompile(OutputPat)

	testInPat := `<textarea name="description.sampleInput" cols="120" rows="15" id="sampleInput">(.*?)</textarea>`
	h.testInRx = regexp.MustCompile(testInPat)
	testOutPat := `<textarea name="description.sampleOutput" cols="120" rows="15" id="sampleOutput">(.*?)</textarea>`
	h.testOutRx = regexp.MustCompile(testOutPat)

	srcPat := `<td>Source:</td><td>(.*?)</td>`
	h.srcRx = regexp.MustCompile(srcPat)

	hintPat := `<textarea name="description.hint" cols="120" rows="15" id="hint">(.*?)</textarea>`
	h.hintRx = regexp.MustCompile(hintPat)

	VJLogfile, err := os.Create("log/vj.log")
	if err != nil {
		log.Println(err)
		return
	}
	VJlogger = log.New(VJLogfile, "[VJ]", log.Ldate|log.Ltime)
}

func (h *VJJudger) GetProblemPage(pid string) (string, error) {
    uv := url.Values{}
    uv.Add("username", "vsake")
    uv.Add("password", "JC945312")

    req, err := http.NewRequest("POST", "http://acm.hust.edu.cn/vjudge/user/login.action", strings.NewReader(uv.Encode()))
    if err != nil {
        return "", ErrConnectFailed
    }
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := h.client.Do(req)
	if err != nil {
        return "", ErrConnectFailed
	}
	defer resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)
	fmt.Println("FFF: ",html)


	resp, err = h.client.Get("http://acm.hust.edu.cn/vjudge/user/checkLogInStatus.action")
	//resp, err = h.client.Get("http://acm.hust.edu.cn/vjudge/problem/toEditDescription.action?id=" + pid)
	if err != nil {
		return "", ErrConnectFailed
	}
	b, _ = ioutil.ReadAll(resp.Body)
	html = string(b)
	fmt.Println("SSS: ",html)
	return html, nil

}
func (h *VJJudger) IsExist(page string) bool {
	return strings.Index(page, "Can not find problem") < 0
}
func (h *VJJudger) ReplaceImg(text string) string {
	return text
}

func (h *VJJudger) SetDetail(pid string, html string) error {
	log.Println(pid)
	pro := model.Problem{}
	pro.RPid, _ = strconv.Atoi(pid)
	pro.ROJ = "VJ"
	pro.Status = StatusReverse

	titleMatch := h.titleRx.FindStringSubmatch(html)
	if len(titleMatch) != 1 {
		log.Println(titleMatch)
		return ErrMatchFailed
	}
	pro.Title = titleMatch[0]

//	if strings.Index(html, "Special Judge") >= 0 {
//		pro.Special = 1
//	}

	TimeMatch := h.TimeRx.FindStringSubmatch(html)
	if len(TimeMatch) != 1 {
		log.Println(TimeMatch)
		return ErrMatchFailed
	}
	pro.Time, _ = strconv.Atoi(TimeMatch[0])

	MemoryMatch := h.MemoryRx.FindStringSubmatch(html)
	if len(MemoryMatch) != 1 {
		log.Println(MemoryMatch)
		return ErrMatchFailed
	}
	pro.Memory, _ = strconv.Atoi(MemoryMatch[0])

	DescriptionMatch := h.DescriptionRx.FindStringSubmatch(html)
	if len(DescriptionMatch) != 1 {
		log.Println(DescriptionMatch)
		return ErrMatchFailed
	}
	pro.Description = template.HTML(h.ReplaceImg(DescriptionMatch[0]))
	InputMatch := h.InputRx.FindStringSubmatch(html)
	if len(InputMatch) != 1 {
		log.Println(InputMatch)
		return ErrMatchFailed
	}
	pro.Input = template.HTML(h.ReplaceImg(InputMatch[0]))
	OutputMatch := h.OutputRx.FindStringSubmatch(html)
	if len(OutputMatch) != 1 {
		log.Println(OutputMatch)
		return ErrMatchFailed
	}
	pro.Output = template.HTML(h.ReplaceImg(OutputMatch[0]))

	testIn := h.testInRx.FindStringSubmatch(html)
	if len(testIn) != 1 {
		log.Println(testIn)
		return ErrMatchFailed
	}
	pro.In = testIn[0]
	testOut := h.testOutRx.FindStringSubmatch(html)
	if len(testOut) != 1 {
		log.Println(testOut)
		return ErrMatchFailed
	}
	pro.Out = testOut[0]

	src := h.srcRx.FindStringSubmatch(html)
	if len(src) >= 1 {
		pro.Source = src[0]
	}

	hint := h.hintRx.FindStringSubmatch(html)
	if len(hint) != 1 {
		log.Println(hint)
		return ErrMatchFailed
	}
	pro.Hint = template.HTML(hint[0])

	proModel := &model.ProblemModel{}
	proModel.Insert(pro)
	return nil
}

func (h *VJJudger) GetProblem(probId int) error {
    pid := strconv.Itoa(probId)
    page, err := h.GetProblemPage(pid)
    if err != nil { //offline
        VJlogger.Println("pid["+pid+"]: ", err, ".")
        return err
    }
    if h.IsExist(page) {
        err := h.SetDetail(pid, page)
        if err != nil {
            VJlogger.Println("pid["+pid+"]: ", "import error.")
        }
    } else {
        VJlogger.Println("pid["+pid+"]: ", "not exist.")
    }

	VJlogger.Println("add problem from VJ, pid is ", probId, ".")
	return nil
}
