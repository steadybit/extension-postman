package extpostman

type NewmanJsonReport struct {
	Run Run `json:"Run"`
}
type Run struct {
	Stats *Stats `json:"Stats"`
}
type Stats struct {
	Requests   *Stat `json:"Requests"`
	Assertions *Stat `json:"Assertions"`
}
type Stat struct {
	Total   int `json:"total"`
	Pending int `json:"pending"`
	Failed  int `json:"failed"`
}
