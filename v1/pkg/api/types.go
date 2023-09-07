package api

type MutatingAdmissionPolicy struct {
	Spec struct {
		Mutation []struct {
			Expressions []string `json:"expressions"`
		} `json:"mutation"`
	} `json:"spec"`
}
