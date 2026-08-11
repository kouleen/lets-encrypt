package modle

type BasePage struct {
	Current int   `json:"current"`
	Size    int   `json:"size"`
	Total   int64 `json:"total"`
	Records any   `json:"records"`
}

func ToPage(list any, total int64) BasePage {
	return BasePage{
		Total:   total,
		Records: list,
	}
}
