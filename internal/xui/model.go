package xui

type VpnKey struct {
	ID     string
	ChatID int64
	Title  string
	Key    string
}

type Inbound struct {
	ID             int    `json:"id"`
	Port           int    `json:"port"`
	Protocol       string `json:"protocol"`
	Remark         string `json:"remark"`
	Settings       string `json:"settings"`
	StreamSettings string `json:"streamSettings"`
}

type InboundSettings struct {
	Clients []InboundClient `json:"clients"`
}

type InboundClient struct {
	ID      string `json:"id"`
	Flow    string `json:"flow"`
	Email   string `json:"email"`
	Enable  bool   `json:"enable"`
	Comment string `json:"comment"`
}

type InboundStreamSettings struct {
	Network         string `json:"network"`
	Security        string `json:"security"`
	RealitySettings struct {
		ShortIDS    []string `json:"shortIds"`
		ServerNames []string `json:"serverNames"`
		Settings    struct {
			PublikKey   string `json:"publicKey"`
			Fingerprint string `json:"fingerprint"`
			SpiderX     string `json:"spiderX"`
		} `json:"settings"`
	} `json:"realitySettings"`
}

type CreateClientRequest struct {
	ID       int    `json:"id"`
	Settings string `json:"settings"`
}

type CreateClientResponse struct {
	Success bool   `json:"success"`
	Msg     string `json:"msg"`
}
