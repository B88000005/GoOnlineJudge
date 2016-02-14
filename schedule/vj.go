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
)

type VJJudger struct {
	client    *http.Client
	testRx    *regexp.Regexp
	titleRx   *regexp.Regexp
	resLimtRx *regexp.Regexp
	ctxRx     *regexp.Regexp
	srcRx     *regexp.Regexp
	hintRx    *regexp.Regexp
}

var VJlogger *log.Logger

func (p *VJJudger) Host() string {
	return "VJ"
}
func (p *VJJudger) Ping() error {
	p.client = &http.Client{Timeout: time.Second * 10}
	resp, err := p.client.Get("http://codeforces.com/problemset/problem/1/A")
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

	titlePat := `<a href="/vjudge/problem/visitOriginUrl.action?id=.*?" target="_blank">(.*?)</a>`

	h.titleRx = regexp.MustCompile(titlePat)

	resLimtPat := `<td><b>Time Limit:</b> (\d+)MS</td><td width="10px"></td><td><b>Memory Limit:</b> (\d+)KB</td>`
	h.resLimtRx = regexp.MustCompile(resLimtPat)

	ctxPat := `(?s)<div class="textBG"><p>(.*?)</p></div>`
	h.ctxRx = regexp.MustCompile(ctxPat)

	testPat := `(?s)put</p><div class="textBG">(.*?)</div></div>`
	h.testRx = regexp.MustCompile(testPat)

	srcPat := `<a style="color: black" href="http://codeforces.com/contest/.*?">(.*?)</a>`
	h.srcRx = regexp.MustCompile(srcPat)

//	hintPat := `(?s)<p class="pst">Hint</p><div class="ptx" lang=".*?">(.*?)</div>`
//	h.hintRx = regexp.MustCompile(hintPat)

	VJLogfile, err := os.Create("log/vj.log")
	if err != nil {
		log.Println(err)
		return
	}
	VJlogger = log.New(VJLogfile, "[VJ]", log.Ldate|log.Ltime)
}

func (h *VJJudger) GetProblemPage(pid string) (string, error) {
	resp, err := h.client.Get("http://acm.hust.edu.cn/vjudge/problem/viewProblem.action?id=" + pid)
	if err != nil {
		return "", ErrConnectFailed
	}
	b, _ := ioutil.ReadAll(resp.Body)
	html := string(b)
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
	if len(titleMatch) < 1 {
		log.Println(titleMatch)
		return ErrMatchFailed
	}
	pro.Title = titleMatch[1]

	if strings.Index(html, "Special Judge") >= 0 {
		pro.Special = 1
	}

	resMatch := h.resLimtRx.FindStringSubmatch(html)
	if len(resMatch) < 3 {
		log.Println(resMatch)
		return ErrMatchFailed
	}
	pro.Time, _ = strconv.Atoi(resMatch[1])
	pro.Memory, _ = 1024 * strconv.Atoi(resMatch[2])

	cxtMatch := h.ctxRx.FindAllStringSubmatch(html, 4)
	if len(cxtMatch) < 3 {
		log.Println("ctx match error, VJ pid is", pid)
		return ErrMatchFailed
	}
	pro.Description = template.HTML(h.ReplaceImg(cxtMatch[0][1]))
	pro.Input = template.HTML(h.ReplaceImg(cxtMatch[1][1]))
	pro.Output = template.HTML(h.ReplaceImg(cxtMatch[2][1]))

	test := h.testRx.FindAllStringSubmatch(html, 2)
	if len(test) < 2 {
		log.Println("test data error, VJ pid is", pid)
		return ErrMatchFailed
	}
	pro.In = test[0][1]
	pro.Out = test[1][1]

	src := h.srcRx.FindStringSubmatch(html)
	if len(src) > 1 {
		pro.Source = src[1]
	}

	if len(cxtMatch) >= 4 {
		pro.Hint = template.HTML(cxtMatch[3][1])
	}

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
