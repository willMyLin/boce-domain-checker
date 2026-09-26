package main

type DetectRequest struct {
	APIKey            string   `json:"apiKey"`
	EnablePollution   bool     `json:"enablePollution"`
	EnableHijack      bool     `json:"enableHijack"`
	EnableWechat      bool     `json:"enableWechat"`
	EnableICP         bool     `json:"enableIcp"`
	EnableBlacklist   bool     `json:"enableBlacklist"`
	EnableWall        bool     `json:"enableWall"`
	Concurrency       int      `json:"concurrency"`
	TimeoutSeconds    int      `json:"timeoutSeconds"`
	ImportedTargetCnt int      `json:"importedTargetCount"`
	Targets           []string `json:"targets"`
}

type DetectRow struct {
	ID          int    `json:"id"`
	Type        string `json:"type"`
	Target      string `json:"target"`
	Status      string `json:"status"`
	CheckedAt   string `json:"checkedAt"`
	ErrorRemark string `json:"errorRemark"`
	Domain      string `json:"domain"`
	BeianCode   string `json:"beianCode"`
	SiteName    string `json:"siteName"`
}

type DetectSummary struct {
	Total        int `json:"total"`
	Checked      int `json:"checked"`
	Pollution    int `json:"pollution"`
	Normal       int `json:"normal"`
	Unregistered int `json:"unregistered"`
	Failed       int `json:"failed"`
}

type DetectResponse struct {
	Rows       []DetectRow   `json:"rows"`
	Summary    DetectSummary `json:"summary"`
	Progress   int           `json:"progress"`
	ExportPath string        `json:"exportPath"`
	Message    string        `json:"message"`
}

type ImportResponse struct {
	Targets  []string `json:"targets"`
	Message  string   `json:"message"`
	Canceled bool     `json:"canceled"`
}

type ExportResponse struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

type DetectStartEvent struct {
	Total int `json:"total"`
}

type bocePolluteResponse struct {
	ErrorCode int            `json:"error_code"`
	Error     string         `json:"error"`
	Data      map[string]int `json:"data"`
}

type boceQQResponse struct {
	ErrorCode int            `json:"error_code"`
	Error     string         `json:"error"`
	Data      map[string]int `json:"data"`
}

type boceICPResponse struct {
	ErrorCode int                    `json:"error_code"`
	Error     string                 `json:"error"`
	Data      map[string]boceICPInfo `json:"data"`
}

type boceICPInfo struct {
	CompanyBeianCode string `json:"companyBeianCode"`
	SiteBeianCode    string `json:"siteBeianCode"`
	Domain           string `json:"domain"`
	Name             string `json:"name"`
	Status           int    `json:"status"`
	CompanyName      string `json:"companyName"`
	WebsiteIndex     string `json:"websiteIndex"`
	Type             string `json:"type"`
	VerifyTime       string `json:"verifyTime"`
	UpdateTime       string `json:"updateTime"`
}
