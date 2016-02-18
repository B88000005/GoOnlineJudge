package schedule

import (
	"GoOnlineJudge/model"
	"html/template"
	"html"
	"io/ioutil"
	"log"
	"net/http"
    "sync"
    "net/url"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
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

	titlePat := `<td>Title:</td>\s*<td>(.*?)</td>`
    h.titleRx = regexp.MustCompile(titlePat)

	TimePat := `<td>Time Limit:</td>\s*<td>(\d+) MS</td>`
	h.TimeRx = regexp.MustCompile(TimePat)
	MemoryPat := `<td>Memory Limit:</td>\s*<td>(\d+) KB</td>`
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

////////////////////////////////

type Jar struct {
    lk      sync.Mutex
    cookies map[string][]*http.Cookie
}
func NewJar() *Jar {
    jar := new(Jar)
    jar.cookies = make(map[string][]*http.Cookie)
    return jar
}
func (jar *Jar) SetCookies(u *url.URL, cookies []*http.Cookie) {
    jar.lk.Lock()
    jar.cookies[u.Host] = cookies
    jar.lk.Unlock()
}
func (jar *Jar) Cookies(u *url.URL) []*http.Cookie {
    return jar.cookies[u.Host]
}

////////////////////////////////

func (h *VJJudger) GetProblemPage(pid string) (string, error) {
    jar := NewJar()
    h.client = &http.Client{Jar: jar}
    resp, err := h.client.PostForm("http://acm.hust.edu.cn/vjudge/user/login.action", url.Values{
        "username": {"vsake"},
        "password": {"JC945312"},
    })
	resp.Body.Close()

	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)


	//resp, err = h.client.Get("http://acm.hust.edu.cn/vjudge/user/checkLogInStatus.action")

	resp, err = h.client.Get("http://acm.hust.edu.cn/vjudge/problem/viewProblem.action?id=" + pid)
	if err != nil {
		return "", ErrConnectFailed
	}
	b, _ = ioutil.ReadAll(resp.Body)
	html = string(b)
	eidPat := `"/vjudge/problem/toEditDescription.action.id=(.*?)"`
	eidRx := regexp.MustCompile(eidPat)
	eidMatch := eidRx.FindStringSubmatch(html)
	if len(eidMatch) != 2 {
		log.Println(eidMatch)
		return "", ErrConnectFailed
	}
	eid := eidMatch[1]

	resp, err = h.client.Get("http://acm.hust.edu.cn/vjudge/problem/toEditDescription.action?id=" + eid)
	if err != nil {
		return "", ErrConnectFailed
	}
	b, _ = ioutil.ReadAll(resp.Body)
	html = string(b)
	return html, nil

}
func (h *VJJudger) IsExist(page string) bool {
	return strings.Index(page, "Can not find problem") < 0
}
func (h *VJJudger) ReplaceHtml(text string) string {
	return html.UnescapeString(text)
}

func (h *VJJudger) SetDetail(pid string, html string) error {
	log.Println(pid)
	pro := model.Problem{}
	pro.RPid, _ = strconv.Atoi(pid)
	pro.ROJ = "VJ"
	pro.Status = StatusReverse

	titleMatch := h.titleRx.FindStringSubmatch(html)
	if len(titleMatch) != 2 {
		log.Println(titleMatch)
		return ErrMatchFailed
	}
	pro.Title = titleMatch[1]

//	if strings.Index(html, "Special Judge") >= 0 {
//		pro.Special = 1
//	}

	TimeMatch := h.TimeRx.FindStringSubmatch(html)
	if len(TimeMatch) != 2 {
		log.Println(TimeMatch)
		return ErrMatchFailed
	}
	pro.Time, _ = strconv.Atoi(TimeMatch[1])

	MemoryMatch := h.MemoryRx.FindStringSubmatch(html)
	if len(MemoryMatch) != 2 {
		log.Println(MemoryMatch)
		return ErrMatchFailed
	}
	pro.Memory, _ = strconv.Atoi(MemoryMatch[1])

	DescriptionMatch := h.DescriptionRx.FindStringSubmatch(html)
	if len(DescriptionMatch) != 2 {
		log.Println(DescriptionMatch)
		return ErrMatchFailed
	}
	pro.Description = template.HTML(h.ReplaceHtml(DescriptionMatch[1]))
	InputMatch := h.InputRx.FindStringSubmatch(html)
	if len(InputMatch) != 2 {
		log.Println(InputMatch)
		return ErrMatchFailed
	}
	pro.Input = template.HTML(h.ReplaceHtml(InputMatch[1]))
	OutputMatch := h.OutputRx.FindStringSubmatch(html)
	if len(OutputMatch) != 2 {
		log.Println(OutputMatch)
		return ErrMatchFailed
	}

	testIn := h.testInRx.FindStringSubmatch(html)
	if len(testIn) != 2 {
		log.Println(testIn)
		return ErrMatchFailed
	}
	pro.In = ""
	testOut := h.testOutRx.FindStringSubmatch(html)
	if len(testOut) != 2 {
		log.Println(testOut)
		return ErrMatchFailed
	}
	pro.Out = ""

	pro.Output = template.HTML(h.ReplaceHtml(OutputMatch[1]+"\n"+testIn[1]+"\n"+testOut[1]))

	src := h.srcRx.FindStringSubmatch(html)
	if len(src) >= 2 {
		pro.Source = src[1]
	}

	hint := h.hintRx.FindStringSubmatch(html)
	if len(hint) != 2 {
		log.Println(hint)
		return ErrMatchFailed
	}
	pro.Hint = template.HTML(h.ReplaceHtml(hint[1]))

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
