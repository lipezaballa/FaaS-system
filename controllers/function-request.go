package controllers

type FunctionRequest struct {
	Name     string `json:"name"`
	ImageRef string `json:"image_ref"`
}

type FunctionParameter struct {
	Param string `json:"param"`
}