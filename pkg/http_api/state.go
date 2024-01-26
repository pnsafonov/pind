package http_api

type State struct {
	Error string `json:"error"`

	Pool Pool `json:"pool"`
}

type Pool struct {
	Idle []int `json:"idle"`
	Load Load  `json:"load"`
}

type Load struct {
	Used []int `json:"used"`
	Free []int `json:"free"`
}
